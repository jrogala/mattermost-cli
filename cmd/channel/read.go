package channel

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/jrogala/mattermost-cli/pkg/ops"
	"github.com/spf13/cobra"
)

var (
	readLimit int
	readTrunc int
	readSince string
)

func init() {
	readCmd.Flags().IntVarP(&readLimit, "limit", "n", 50, "max number of messages to show")
	readCmd.Flags().IntVar(&readTrunc, "truncate", 120, "max chars per message (0 = no truncation)")
	readCmd.Flags().StringVar(&readSince, "since", "", "show messages since (e.g. 1h, 24h, 2d, 1w, 2026-03-20)")
	Cmd.AddCommand(readCmd)
}

var readCmd = &cobra.Command{
	Use:   "read <channel_id_or_name>",
	Short: "Read recent messages from a channel. Use --since to filter by time.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdutil.NewClient()

		channelID, err := ops.ResolveChannelArg(c, args[0])
		if err != nil {
			return err
		}

		messages, err := ops.ReadMessages(c, channelID, ops.ReadOptions{
			Limit: readLimit,
			Since: readSince,
		})
		if err != nil {
			return err
		}

		cmdutil.Render(cmd, messages, func() {
			if len(messages) == 0 {
				fmt.Println("No messages.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			for _, m := range messages {
				ts := m.Time.Format("Jan 02 15:04")
				msg := cmdutil.TruncateMsg(m.Text, readTrunc)
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", ts, m.User, msg)
			}
			_ = w.Flush()
		})
		return nil
	},
}
