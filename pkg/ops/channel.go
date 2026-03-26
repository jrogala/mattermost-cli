package ops

import (
	"fmt"
	"strings"

	"github.com/jrogala/mattermost-cli/client"
)

// Channel represents a Mattermost channel with optional unread/mute state.
type Channel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Unread      int64  `json:"unread,omitempty"`
	Mentions    int    `json:"mentions,omitempty"`
	Muted       bool   `json:"muted,omitempty"`
}

// ListOptions configures channel listing.
type ListOptions struct {
	TeamID       string
	Type         string // "public", "private", "dm", "group", or ""
	IncludeMuted bool
}

// ListChannels returns channels the user belongs to, filtered by options.
func ListChannels(c *client.Client, opts ListOptions) ([]Channel, error) {
	me, err := c.Me()
	if err != nil {
		return nil, err
	}

	teamID, err := ResolveTeam(c, opts.TeamID)
	if err != nil {
		return nil, err
	}

	channels, err := c.GetChannels(teamID)
	if err != nil {
		return nil, err
	}

	members, err := c.GetChannelMembers(teamID)
	if err != nil {
		return nil, err
	}
	mutedSet := make(map[string]bool, len(members))
	for _, m := range members {
		if m.IsMuted() {
			mutedSet[m.ChannelID] = true
		}
	}

	if !opts.IncludeMuted {
		var filtered []client.Channel
		for _, ch := range channels {
			if !mutedSet[ch.ID] {
				filtered = append(filtered, ch)
			}
		}
		channels = filtered
	}

	if opts.Type != "" {
		typeCode := channelTypeCode(opts.Type)
		var filtered []client.Channel
		for _, ch := range channels {
			if ch.Type == typeCode {
				filtered = append(filtered, ch)
			}
		}
		channels = filtered
	}

	var entries []Channel
	for _, ch := range channels {
		entries = append(entries, Channel{
			ID:          ch.ID,
			Name:        ch.Name,
			DisplayName: ResolveChannelName(c, ch, me.ID),
			Type:        channelTypeName(ch.Type),
			Muted:       mutedSet[ch.ID],
		})
	}
	return entries, nil
}

// FindChannels searches for channels matching a term.
func FindChannels(c *client.Client, term string, typeFilter string) ([]Channel, error) {
	me, err := c.Me()
	if err != nil {
		return nil, err
	}

	teamID, err := ResolveTeam(c, "")
	if err != nil {
		return nil, err
	}

	channels, err := c.GetChannels(teamID)
	if err != nil {
		return nil, err
	}

	if typeFilter != "" {
		code := channelTypeCode(typeFilter)
		var filtered []client.Channel
		for _, ch := range channels {
			if ch.Type == code {
				filtered = append(filtered, ch)
			}
		}
		channels = filtered
	}

	termLower := strings.ToLower(term)
	var results []Channel
	for _, ch := range channels {
		name := ResolveChannelName(c, ch, me.ID)
		if strings.Contains(strings.ToLower(name), termLower) ||
			strings.Contains(strings.ToLower(ch.Name), termLower) {
			results = append(results, Channel{
				ID:          ch.ID,
				Name:        ch.Name,
				DisplayName: name,
				Type:        channelTypeName(ch.Type),
			})
		}
	}
	return results, nil
}

// FindDMChannel finds or creates a DM channel with a user by username.
func FindDMChannel(c *client.Client, username string) (*Channel, error) {
	u, err := c.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("user %q not found: %w", username, err)
	}
	dm, err := c.GetOrCreateDM(u.ID)
	if err != nil {
		return nil, err
	}
	return &Channel{
		ID:          dm.ID,
		DisplayName: u.Username + " (DM)",
		Type:        "dm",
	}, nil
}

// SendResult holds the result of sending a message.
type SendResult struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
}

// SendMessage sends a message to a channel by ID.
func SendMessage(c *client.Client, channelID string, message string) (*SendResult, error) {
	post, err := c.SendPost(channelID, message)
	if err != nil {
		return nil, err
	}
	return &SendResult{
		ID:        post.ID,
		ChannelID: post.ChannelID,
		Message:   post.Message,
	}, nil
}

// SendMessageByName resolves a channel name and sends a message.
func SendMessageByName(c *client.Client, channelName string, message string) (*SendResult, error) {
	channelID, err := ResolveChannelArg(c, channelName)
	if err != nil {
		return nil, err
	}
	return SendMessage(c, channelID, message)
}

func channelTypeName(t string) string {
	switch t {
	case "O":
		return "public"
	case "P":
		return "private"
	case "D":
		return "dm"
	case "G":
		return "group"
	default:
		return t
	}
}

func channelTypeCode(name string) string {
	switch name {
	case "public":
		return "O"
	case "private":
		return "P"
	case "dm":
		return "D"
	case "group":
		return "G"
	default:
		return name
	}
}
