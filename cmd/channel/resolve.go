package channel

import (
	"fmt"
	"strings"

	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/internal/cmdutil"
)

// resolveChannelArg takes a channel ID or a fuzzy name and returns the channel ID.
// If the argument looks like a Mattermost ID (26 alphanumeric chars), it's returned as-is.
// Otherwise, it searches the user's channels for a fuzzy match.
// If multiple channels match, it returns an error listing them.
func resolveChannelArg(c *client.Client, arg string) (string, error) {
	// Mattermost IDs are 26 lowercase alphanumeric chars
	if looksLikeID(arg) {
		return arg, nil
	}

	me, err := c.Me()
	if err != nil {
		return "", err
	}

	teamID, err := cmdutil.ResolveTeam(c, "")
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
		name := cmdutil.ResolveChannelName(c, ch, me.ID)
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

	// Multiple matches: show them and error
	var lines []string
	for _, ch := range matches {
		name := cmdutil.ResolveChannelName(c, ch, me.ID)
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
