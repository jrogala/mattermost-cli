package ops

import (
	"fmt"
	"strings"

	"github.com/jrogala/mattermost-cli/client"
)

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
	var lines []string
	for _, t := range teams {
		lines = append(lines, fmt.Sprintf("  %s  %s", t.ID, t.DisplayName))
	}
	return "", fmt.Errorf("--team is required when you belong to multiple teams:\n%s", strings.Join(lines, "\n"))
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

// ResolveChannelArg takes a channel ID or a fuzzy name and returns the channel ID.
// If the argument looks like a Mattermost ID (26 alphanumeric chars), it's returned as-is.
// Otherwise, it searches the user's channels for a fuzzy match.
func ResolveChannelArg(c *client.Client, arg string) (string, error) {
	if looksLikeID(arg) {
		return arg, nil
	}

	me, err := c.Me()
	if err != nil {
		return "", err
	}

	teamID, err := ResolveTeam(c, "")
	if err != nil {
		return "", err
	}

	channels, err := c.GetChannels(teamID)
	if err != nil {
		return "", err
	}

	termLower := strings.ToLower(arg)
	var matches []client.Channel
	for _, ch := range channels {
		name := ResolveChannelName(c, ch, me.ID)
		if strings.Contains(strings.ToLower(name), termLower) ||
			strings.Contains(strings.ToLower(ch.Name), termLower) {
			matches = append(matches, ch)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no channel matching %q", arg)
	}
	if len(matches) == 1 {
		return matches[0].ID, nil
	}

	var lines []string
	for _, ch := range matches {
		name := ResolveChannelName(c, ch, me.ID)
		lines = append(lines, fmt.Sprintf("  %s  %s", ch.ID, name))
	}
	return "", fmt.Errorf("multiple channels match %q, use the ID:\n%s", arg, strings.Join(lines, "\n"))
}

func looksLikeID(s string) bool {
	if len(s) != 26 {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
