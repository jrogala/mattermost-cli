package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/jrogala/mattermost-cli/pkg/ops"
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
		entries, err := ops.GetUnread(c, ops.UnreadOptions{
			IncludeMuted: unreadShowMuted,
		})
		if err != nil {
			return err
		}

		cmdutil.Render(cmd, entries, func() {
			if len(entries) == 0 {
				fmt.Println("All caught up!")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "UNREAD\tMENTIONS\tCHANNEL")
			for _, e := range entries {
				muted := ""
				if e.Muted {
					muted = " (muted)"
				}
				_, _ = fmt.Fprintf(w, "%d\t%d\t%s%s\n", e.Unread, e.Mentions, e.DisplayName, muted)
			}
			_ = w.Flush()
		})
		return nil
	},
}
