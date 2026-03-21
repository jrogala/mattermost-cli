package channel

import (
	"fmt"
	"strings"

	"github.com/jrogala/mattermost-cli/internal/cmdutil"
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
		channelID, err := resolveChannelArg(c, args[0])
		if err != nil {
			return err
		}
		message := strings.Join(args[1:], " ")

		post, err := c.SendPost(channelID, message)
		if err != nil {
			return err
		}

		if cmdutil.IsJSON(cmd) {
			return cmdutil.PrintJSON(post)
		}

		fmt.Printf("Message sent to %s\n", post.ChannelID)
		return nil
	},
}
