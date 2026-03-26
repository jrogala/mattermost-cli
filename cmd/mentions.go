package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/jrogala/mattermost-cli/pkg/ops"
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
		mentions, err := ops.GetMentions(c, ops.MentionOptions{
			Limit: mentionsLimit,
		})
		if err != nil {
			return err
		}

		cmdutil.Render(cmd, mentions, func() {
			if len(mentions) == 0 {
				fmt.Println("No mentions.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "TIME\tCHANNEL\tFROM\tMESSAGE")
			for _, m := range mentions {
				ts := m.Time.Format("Jan 02 15:04")
				msg := truncateMsg(m.Text, 150)
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ts, m.Channel, m.User, msg)
			}
			_ = w.Flush()
		})
		return nil
	},
}
