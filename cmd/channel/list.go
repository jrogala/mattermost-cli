package channel

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var (
	listTeam        string
	listType        string
	listIncludeMuted bool
)

func init() {
	listCmd.Flags().StringVarP(&listTeam, "team", "t", "", "team ID (required if multiple teams)")
	listCmd.Flags().StringVar(&listType, "type", "", "filter by type: public, private, dm, group")
	listCmd.Flags().BoolVar(&listIncludeMuted, "include-muted", false, "include muted channels")
	Cmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls", "l"},
	Short:   "List non-muted channels you belong to.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		c := cmdutil.NewClient()

		me, err := c.Me()
		if err != nil {
			return err
		}

		teamID, err := cmdutil.ResolveTeam(c, listTeam)
		if err != nil {
			return err
		}

		channels, err := c.GetChannels(teamID)
		if err != nil {
			return err
		}

		// Filter muted unless --include-muted
		if !listIncludeMuted {
			members, err := c.GetChannelMembers(teamID)
			if err != nil {
				return err
			}
			mutedSet := make(map[string]bool, len(members))
			for _, m := range members {
				if m.IsMuted() {
					mutedSet[m.ChannelID] = true
				}
			}
			var filtered []client.Channel
			for _, ch := range channels {
				if !mutedSet[ch.ID] {
					filtered = append(filtered, ch)
				}
			}
			channels = filtered
		}

		if listType != "" {
			typeCode := channelTypeCode(listType)
			var filtered []client.Channel
			for _, ch := range channels {
				if ch.Type == typeCode {
					filtered = append(filtered, ch)
				}
			}
			channels = filtered
		}

		if cmdutil.IsJSON(cmd) {
			return cmdutil.PrintJSON(channels)
		}

		if len(channels) == 0 {
			fmt.Println("No channels found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "ID\tTYPE\tNAME")
		for _, ch := range channels {
			name := cmdutil.ResolveChannelName(c, ch, me.ID)
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", ch.ID, channelTypeName(ch.Type), name)
		}
		return w.Flush()
	},
}
