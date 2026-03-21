package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var unreadShowMuted bool

func init() {
	rootCmd.AddCommand(unreadCmd)
	unreadCmd.Flags().BoolVar(&unreadShowMuted, "include-muted", false, "include muted channels")
}

var unreadCmd = &cobra.Command{
	Use:   "unread",
	Short: "List channels with unread messages (excludes muted channels).",
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

		type unreadEntry struct {
			Channel  client.Channel
			Unread   int64
			Mentions int
			Muted    bool
		}

		var entries []unreadEntry
		for _, ch := range channels {
			m, ok := memberMap[ch.ID]
			if !ok {
				continue
			}
			unread := ch.TotalMsgCount - m.MsgCount
			if unread <= 0 && m.MentionCount <= 0 {
				continue
			}
			if m.IsMuted() && !unreadShowMuted {
				continue
			}
			entries = append(entries, unreadEntry{
				Channel:  ch,
				Unread:   unread,
				Mentions: m.MentionCount,
				Muted:    m.IsMuted(),
			})
		}

		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Mentions != entries[j].Mentions {
				return entries[i].Mentions > entries[j].Mentions
			}
			return entries[i].Unread > entries[j].Unread
		})

		if isJSON(cmd) {
			type jsonEntry struct {
				ChannelID   string `json:"channel_id"`
				Name        string `json:"name"`
				DisplayName string `json:"display_name"`
				Type        string `json:"type"`
				Unread      int64  `json:"unread"`
				Mentions    int    `json:"mentions"`
				Muted       bool   `json:"muted"`
			}
			var out []jsonEntry
			for _, e := range entries {
				name := cmdutil.ResolveChannelName(c, e.Channel, me.ID)
				out = append(out, jsonEntry{
					ChannelID:   e.Channel.ID,
					Name:        e.Channel.Name,
					DisplayName: name,
					Type:        e.Channel.Type,
					Unread:      e.Unread,
					Mentions:    e.Mentions,
					Muted:       e.Muted,
				})
			}
			return printJSON(out)
		}

		if len(entries) == 0 {
			fmt.Println("All caught up!")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "UNREAD\tMENTIONS\tCHANNEL")
		for _, e := range entries {
			name := cmdutil.ResolveChannelName(c, e.Channel, me.ID)
			muted := ""
			if e.Muted {
				muted = " (muted)"
			}
			_, _ = fmt.Fprintf(w, "%d\t%d\t%s%s\n", e.Unread, e.Mentions, name, muted)
		}
		return w.Flush()
	},
}
