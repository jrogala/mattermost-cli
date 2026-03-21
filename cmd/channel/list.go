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
	listTeam         string
	listType         string
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
		entries, err := ops.ListChannels(c, ops.ListOptions{
			TeamID:       listTeam,
			Type:         listType,
			IncludeMuted: listIncludeMuted,
		})
		if err != nil {
			return err
		}

		cmdutil.Render(cmd, entries, func() {
			if len(entries) == 0 {
				fmt.Println("No channels found.")
				return
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tTYPE\tNAME")
			for _, e := range entries {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", e.ID, e.Type, e.DisplayName)
			}
			_ = w.Flush()
		})
		return nil
	},
}
