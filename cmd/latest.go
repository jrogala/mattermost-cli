package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/jrogala/mattermost-cli/pkg/ops"
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
		posts, err := ops.GetLatest(c, ops.LatestOptions{
			Channels:   latestLimit,
			PerChannel: latestPerChan,
		})
		if err != nil {
			return err
		}

		cmdutil.Render(cmd, posts, func() {
			if len(posts) == 0 {
				fmt.Println("No recent messages.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			currentChannel := ""
			for _, p := range posts {
				if p.Channel != currentChannel {
					if currentChannel != "" {
						_, _ = fmt.Fprintln(w)
					}
					_, _ = fmt.Fprintf(w, "--- %s ---\n", p.Channel)
					currentChannel = p.Channel
				}
				ts := p.Time.Format("Jan 02 15:04")
				msg := truncateMsg(p.Message, 120)
				_, _ = fmt.Fprintf(w, "  %s\t%s\t%s\n", ts, p.User, msg)
			}
			_ = w.Flush()
		})
		return nil
	},
}
