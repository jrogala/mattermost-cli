// Package cmdutil provides shared helpers for CLI commands.
package cmdutil

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

// ResolveTeam returns the team ID. If only one team, returns it automatically.
func ResolveTeam(c *client.Client, teamID string) (string, error) {
	if teamID != "" {
		return teamID, nil
	}
	teams, err := c.GetTeams()
	if err != nil {
		return "", err
	}
	if len(teams) == 0 {
		return "", fmt.Errorf("you are not a member of any team")
	}
	if len(teams) == 1 {
		return teams[0].ID, nil
	}
	fmt.Fprintln(os.Stderr, "Multiple teams found, use --team <id>:")
	for _, t := range teams {
		fmt.Fprintf(os.Stderr, "  %s  %s\n", t.ID, t.DisplayName)
	}
	return "", fmt.Errorf("--team is required when you belong to multiple teams")
}

// ResolveChannelName returns a human-friendly name for a channel.
func ResolveChannelName(c *client.Client, ch client.Channel, myUserID string) string {
	if ch.DisplayName != "" && ch.Type != "D" {
		return ch.DisplayName
	}
	if ch.Type == "D" {
		parts := strings.SplitN(ch.Name, "__", 2)
		if len(parts) == 2 {
			otherID := parts[0]
			if otherID == myUserID {
				otherID = parts[1]
			}
			u, err := c.GetUser(otherID)
			if err == nil {
				return u.Username + " (DM)"
			}
		}
		return ch.Name
	}
	if ch.DisplayName != "" {
		return ch.DisplayName
	}
	return ch.Name
}
