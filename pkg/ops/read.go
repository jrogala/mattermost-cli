package ops

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jrogala/mattermost-cli/client"
)

// Message represents a single channel message (from REST or WebSocket).
type Message struct {
	ID        string    `json:"id"`
	ChannelID string    `json:"channel_id"`
	Channel   string    `json:"channel_name,omitempty"`
	User      string    `json:"user"`
	UserID    string    `json:"user_id,omitempty"`
	Text      string    `json:"message"`
	Time      time.Time `json:"time"`
	Event     string    `json:"event,omitempty"`     // WS event type: "posted", "post_deleted", etc.
	PostType  string    `json:"post_type,omitempty"` // system post type if any
}

// ReadOptions configures message reading.
type ReadOptions struct {
	Limit int    // max messages, 0 = default (50)
	Since string // duration (1h, 2d) or date (2026-03-20)
}

// ReadMessages reads messages from a channel by ID.
func ReadMessages(c *client.Client, channelID string, opts ReadOptions) ([]Message, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	params := map[string]string{
		"per_page": strconv.Itoa(limit),
	}

	if opts.Since != "" {
		sinceMs, err := ParseSince(opts.Since)
		if err != nil {
			return nil, fmt.Errorf("invalid since value %q: %w", opts.Since, err)
		}
		params["since"] = strconv.FormatInt(sinceMs, 10)
	}

	pl, err := c.GetChannelPosts(channelID, params)
	if err != nil {
		return nil, err
	}

	userCache := map[string]string{}
	resolveUser := func(userID string) string {
		if name, ok := userCache[userID]; ok {
			return name
		}
		u, err := c.GetUser(userID)
		if err != nil {
			userCache[userID] = userID[:8]
			return userCache[userID]
		}
		userCache[userID] = u.Username
		return u.Username
	}

	// Reverse order: oldest first, skip system messages
	var messages []Message
	for i := len(pl.Order) - 1; i >= 0; i-- {
		post := pl.Posts[pl.Order[i]]
		if post.IsSystem() {
			continue
		}
		messages = append(messages, Message{
			ID:        post.ID,
			ChannelID: post.ChannelID,
			User:      resolveUser(post.UserID),
			UserID:    post.UserID,
			Text:      post.Message,
			Time:      time.UnixMilli(post.CreateAt),
		})
	}
	return messages, nil
}

// ReadMessagesByName resolves a channel name then reads messages.
func ReadMessagesByName(c *client.Client, channelName string, opts ReadOptions) ([]Message, error) {
	channelID, err := ResolveChannelArg(c, channelName)
	if err != nil {
		return nil, err
	}
	return ReadMessages(c, channelID, opts)
}

// ParseSince converts a duration or date string to a Unix millisecond timestamp.
// Supports: "1h", "24h", "2d", "1w", "YYYY-MM-DD".
func ParseSince(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UnixMilli(), nil
	}

	if len(s) < 2 {
		return 0, fmt.Errorf("use: 1h, 24h, 2d, 1w, or YYYY-MM-DD")
	}

	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("cannot parse number from %q", s)
	}

	var d time.Duration
	switch unit {
	case 'h':
		d = time.Duration(num) * time.Hour
	case 'd':
		d = time.Duration(num) * 24 * time.Hour
	case 'w':
		d = time.Duration(num) * 7 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("unknown unit %q, use h/d/w", string(unit))
	}

	return time.Now().Add(-d).UnixMilli(), nil
}
