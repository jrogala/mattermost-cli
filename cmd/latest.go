package cmd

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var (
	latestLimit   int
	latestPerChan int
)

func init() {
	rootCmd.AddCommand(latestCmd)
	latestCmd.Flags().IntVarP(&latestLimit, "channels", "n", 10, "number of recent channels to show")
	latestCmd.Flags().IntVar(&latestPerChan, "per-channel", 3, "messages to show per channel")
}

var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Show latest messages across recent non-muted channels.",
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

		channels, err := c.GetChannels(teamID)
		if err != nil {
			return err
		}

		members, err := c.GetChannelMembers(teamID)
		if err != nil {
			return err
		}

		memberMap := make(map[string]*client.ChannelMember, len(members))
		for i := range members {
			memberMap[members[i].ChannelID] = &members[i]
		}

		var active []client.Channel
		for _, ch := range channels {
			m, ok := memberMap[ch.ID]
			if !ok {
				continue
			}
			if m.IsMuted() {
				continue
			}
			if ch.LastPostAt == 0 {
				continue
			}
			active = append(active, ch)
		}

		sort.Slice(active, func(i, j int) bool {
			return active[i].LastPostAt > active[j].LastPostAt
		})

		if latestLimit > 0 && len(active) > latestLimit {
			active = active[:latestLimit]
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

		type latestPost struct {
			Channel string `json:"channel"`
			Time    string `json:"time"`
			User    string `json:"user"`
			Message string `json:"message"`
		}

		var allPosts []latestPost

		for _, ch := range active {
			pl, err := c.GetChannelPosts(ch.ID, map[string]string{
				"per_page": strconv.Itoa(latestPerChan),
			})
			if err != nil {
				continue
			}

			name := cmdutil.ResolveChannelName(c, ch, me.ID)

			for i := len(pl.Order) - 1; i >= 0; i-- {
				post := pl.Posts[pl.Order[i]]
				ts := time.UnixMilli(post.CreateAt).Format("Jan 02 15:04")
				msg := truncateMsg(post.Message, 120)
				allPosts = append(allPosts, latestPost{
					Channel: name,
					Time:    ts,
					User:    resolveUser(post.UserID),
					Message: msg,
				})
			}
		}

		if isJSON(cmd) {
			return printJSON(allPosts)
		}

		if len(allPosts) == 0 {
			fmt.Println("No recent messages.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		currentChannel := ""
		for _, p := range allPosts {
			if p.Channel != currentChannel {
				if currentChannel != "" {
					_, _ = fmt.Fprintln(w)
				}
				_, _ = fmt.Fprintf(w, "--- %s ---\n", p.Channel)
				currentChannel = p.Channel
			}
			_, _ = fmt.Fprintf(w, "  %s\t%s\t%s\n", p.Time, p.User, p.Message)
		}
		return w.Flush()
	},
}
