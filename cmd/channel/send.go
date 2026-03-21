package channel

import (
	"fmt"
	"strings"

	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/jrogala/mattermost-cli/pkg/ops"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(sendCmd)
}

var sendCmd = &cobra.Command{
	Use:   "send <channel_id_or_name> <message>",
	Short: "Send a message to a channel.",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdutil.NewClient()
		message := strings.Join(args[1:], " ")

		channelID, err := ops.ResolveChannelArg(c, args[0])
		if err != nil {
			return err
		}

		result, err := ops.SendMessage(c, channelID, message)
		if err != nil {
			return err
		}

		cmdutil.Render(cmd, result, func() {
			fmt.Printf("Message sent to %s\n", result.ChannelID)
		})
		return nil
	},
}
