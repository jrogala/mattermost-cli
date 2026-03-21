package cmd

import (
	"fmt"

	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/jrogala/mattermost-cli/pkg/ops"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(meCmd)
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show authenticated user info.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		c := newClient()
		info, err := ops.GetMe(c)
		if err != nil {
			return err
		}
		cmdutil.Render(cmd, info, func() {
			fmt.Printf("Username: %s\n", info.Username)
			fmt.Printf("Email:    %s\n", info.Email)
			fmt.Printf("ID:       %s\n", info.ID)
		})
		return nil
	},
}
