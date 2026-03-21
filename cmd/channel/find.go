package channel

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/jrogala/mattermost-cli/pkg/ops"
	"github.com/spf13/cobra"
)

var findType string

func init() {
	findCmd.Flags().StringVar(&findType, "type", "", "filter by type: public, private, dm, group, user")
	Cmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:   "find <name>",
	Short: "Find a channel or user DM by name. Use --type user to find a user's DM channel.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdutil.NewClient()
		term := args[0]

		if findType == "user" {
			result, err := ops.FindDMChannel(c, term)
			if err != nil {
				return err
			}
			cmdutil.Render(cmd, result, func() {
				fmt.Printf("%s\t%s (DM)\n", result.ChannelID, result.Username)
			})
			return nil
		}

		results, err := ops.FindChannels(c, term, findType)
		if err != nil {
			return err
		}

		cmdutil.Render(cmd, results, func() {
			if len(results) == 0 {
				fmt.Printf("No channels matching %q\n", term)
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tTYPE\tNAME")
			for _, r := range results {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", r.ID, r.Type, r.Name)
			}
			_ = w.Flush()
		})
		return nil
	},
}
