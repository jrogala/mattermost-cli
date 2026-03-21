// Package cmdutil provides shared helpers for CLI commands.
package cmdutil

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/config"
	"github.com/spf13/cobra"
)

// NewClient creates an authenticated Mattermost client from config.
func NewClient() *client.Client {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	return client.New(cfg)
}

// PrintJSON encodes v as indented JSON to stdout.
func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// TruncateMsg truncates a message to maxLen chars. 0 = no truncation.
func TruncateMsg(msg string, maxLen int) string {
	if maxLen <= 0 || len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen-3] + "..."
}

// IsJSON returns true if the --json persistent flag is set on the command's root.
func IsJSON(cmd *cobra.Command) bool {
	v, _ := cmd.Root().PersistentFlags().GetBool("json")
	return v
}

// Render outputs data as JSON if --json is set, otherwise calls tableFunc.
func Render(cmd *cobra.Command, data any, tableFunc func()) {
	if IsJSON(cmd) {
		_ = PrintJSON(data)
		return
	}
	tableFunc()
}
