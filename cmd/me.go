package cmd

import (
	"fmt"

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
		user, err := c.Me()
		if err != nil {
			return err
		}
		if isJSON(cmd) {
			return printJSON(user)
		}
		fmt.Printf("Username: %s\n", user.Username)
		fmt.Printf("Email:    %s\n", user.Email)
		fmt.Printf("ID:       %s\n", user.ID)
		return nil
	},
}
