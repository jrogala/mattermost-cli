package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var mentionsLimit int

func init() {
	rootCmd.AddCommand(mentionsCmd)
	mentionsCmd.Flags().IntVarP(&mentionsLimit, "limit", "n", 5, "max messages to fetch per channel with mentions")
}

var mentionsCmd = &cobra.Command{
	Use:   "mentions",
	Short: "Show recent messages that mention you (includes muted channels).",
	RunE: func(cmd *cobra.Command, _ []string) error {
		c := newClient()

		me, err := c.Me()
		if err != nil {
			return err
		}

		teamID, err := cmdutil.ResolveTeam(c, "")
		if err != nil {
			return err
		}

		members, err := c.GetChannelMembers(teamID)
		if err != nil {
			return err
		}

		var withMentions []client.ChannelMember
		for _, m := range members {
			if m.MentionCount > 0 {
				withMentions = append(withMentions, m)
			}
		}

		if len(withMentions) == 0 {
			fmt.Println("No mentions.")
			return nil
		}

		chanCache := map[string]string{}
		resolveChan := func(chanID string) string {
			if name, ok := chanCache[chanID]; ok {
				return name
			}
			ch, err := c.GetChannel(chanID)
			if err != nil {
				chanCache[chanID] = chanID[:8]
				return chanCache[chanID]
			}
			name := cmdutil.ResolveChannelName(c, *ch, me.ID)
			chanCache[chanID] = name
			return name
		}

		userCache := map[string]string{}
		resolveUser := func(userID string) string {
			if name, ok := userCache[userID]; ok {
				return name
			}
			u, err := c.GetUser(userID)
			if err != nil {
				userCache[userID] = userID[:8]
				return userCache[userID]
			}
			userCache[userID] = u.Username
			return u.Username
		}

		type mention struct {
			Channel   string `json:"channel"`
			Time      string `json:"time"`
			TimeMilli int64  `json:"time_ms"`
			User      string `json:"user"`
			Message   string `json:"message"`
		}

		var mentions []mention

		username := me.Username
		for _, m := range withMentions {
			pl, err := c.GetChannelPosts(m.ChannelID, map[string]string{
				"per_page": fmt.Sprintf("%d", mentionsLimit),
			})
			if err != nil {
				continue
			}

			chanName := resolveChan(m.ChannelID)
			for _, pid := range pl.Order {
				post := pl.Posts[pid]
				if !containsMention(post.Message, username) {
					continue
				}
				ts := time.UnixMilli(post.CreateAt)
				msg := truncateMsg(post.Message, 150)
				mentions = append(mentions, mention{
					Channel:   chanName,
					Time:      ts.Format("Jan 02 15:04"),
					TimeMilli: post.CreateAt,
					User:      resolveUser(post.UserID),
					Message:   msg,
				})
			}
		}

		sort.Slice(mentions, func(i, j int) bool {
			return mentions[i].TimeMilli > mentions[j].TimeMilli
		})

		if isJSON(cmd) {
			return printJSON(mentions)
		}

		if len(mentions) == 0 {
			fmt.Println("No mentions found in recent messages.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "TIME\tCHANNEL\tFROM\tMESSAGE")
		for _, m := range mentions {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Time, m.Channel, m.User, m.Message)
		}
		return w.Flush()
	},
}

func containsMention(message, username string) bool {
	target := "@" + username
	for i := 0; i <= len(message)-len(target); i++ {
		if message[i:i+len(target)] == target {
			return true
		}
	}
	return false
}
