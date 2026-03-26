// Package client provides an HTTP client for the Mattermost API v4.
package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jrogala/mattermost-cli/config"
)

// Client is an authenticated Mattermost API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	tlsCfg     *tls.Config
}

// New creates a Client from the given config.
func New(cfg *config.Config) *Client {
	baseURL := strings.TrimRight(cfg.URL, "/")
	if !strings.HasSuffix(baseURL, "/api/v4") {
		baseURL += "/api/v4"
	}
	tc := tlsConfig(cfg.TLSSkipVerify)
	return &Client{
		baseURL: baseURL,
		token:   cfg.Token,
		tlsCfg:  tc,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: tc,
			},
		},
	}
}

func (c *Client) do(method, path string, body io.Reader) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// User represents a Mattermost user.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

// Team represents a Mattermost team.
type Team struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// Channel represents a Mattermost channel.
type Channel struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`
	Type          string `json:"type"` // O=public, P=private, D=DM, G=group
	TeamID        string `json:"team_id"`
	TotalMsgCount int64  `json:"total_msg_count"`
	LastPostAt    int64  `json:"last_post_at"`
}

// NotifyProps holds channel notification settings.
type NotifyProps struct {
	MarkUnread string `json:"mark_unread"` // "all" or "mention" (mention = muted)
}

// ChannelMember represents a user's membership in a channel.
type ChannelMember struct {
	ChannelID    string      `json:"channel_id"`
	UserID       string      `json:"user_id"`
	MsgCount     int64       `json:"msg_count"`
	MentionCount int         `json:"mention_count"`
	NotifyProps  NotifyProps `json:"notify_props"`
	LastViewedAt int64       `json:"last_viewed_at"`
}

// Post represents a Mattermost message.
type Post struct {
	ID        string `json:"id"`
	ChannelID string `json:"channel_id"`
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	Type      string `json:"type"` // empty for user posts, "system_*" for system messages
	CreateAt  int64  `json:"create_at"`
	UpdateAt  int64  `json:"update_at"`
}

// IsSystem returns true if this is a system-generated post (join, leave, etc).
func (p *Post) IsSystem() bool {
	return p.Type != ""
}

// PostList is the API response for channel posts.
type PostList struct {
	Order []string         `json:"order"`
	Posts map[string]*Post `json:"posts"`
}

// Me returns the authenticated user.
func (c *Client) Me() (*User, error) {
	data, err := c.do("GET", "/users/me", nil)
	if err != nil {
		return nil, err
	}
	var u User
	return &u, json.Unmarshal(data, &u)
}

// GetTeams returns the user's teams.
func (c *Client) GetTeams() ([]Team, error) {
	me, err := c.Me()
	if err != nil {
		return nil, err
	}
	data, err := c.do("GET", fmt.Sprintf("/users/%s/teams", me.ID), nil)
	if err != nil {
		return nil, err
	}
	var teams []Team
	return teams, json.Unmarshal(data, &teams)
}

// GetChannels returns channels for a team that the user belongs to.
func (c *Client) GetChannels(teamID string) ([]Channel, error) {
	me, err := c.Me()
	if err != nil {
		return nil, err
	}
	data, err := c.do("GET", fmt.Sprintf("/users/%s/teams/%s/channels", me.ID, teamID), nil)
	if err != nil {
		return nil, err
	}
	var channels []Channel
	return channels, json.Unmarshal(data, &channels)
}

// GetChannelPosts returns recent posts in a channel.
func (c *Client) GetChannelPosts(channelID string, params map[string]string) (*PostList, error) {
	path := fmt.Sprintf("/channels/%s/posts", channelID)
	if len(params) > 0 {
		v := url.Values{}
		for k, val := range params {
			v.Set(k, val)
		}
		path += "?" + v.Encode()
	}
	data, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	var pl PostList
	return &pl, json.Unmarshal(data, &pl)
}

// SendPost sends a message to a channel.
func (c *Client) SendPost(channelID, message string) (*Post, error) {
	payload := map[string]string{
		"channel_id": channelID,
		"message":    message,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	data, err := c.do("POST", "/posts", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	var p Post
	return &p, json.Unmarshal(data, &p)
}

// GetUser returns a user by ID.
func (c *Client) GetUser(userID string) (*User, error) {
	data, err := c.do("GET", fmt.Sprintf("/users/%s", userID), nil)
	if err != nil {
		return nil, err
	}
	var u User
	return &u, json.Unmarshal(data, &u)
}

// SearchChannels searches channels by name in a team.
func (c *Client) SearchChannels(teamID, term string) ([]Channel, error) {
	payload := map[string]string{"term": term}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	data, err := c.do("POST", fmt.Sprintf("/teams/%s/channels/search", teamID), strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	var channels []Channel
	return channels, json.Unmarshal(data, &channels)
}

// GetChannelMembers returns all channel memberships for the user in a team.
func (c *Client) GetChannelMembers(teamID string) ([]ChannelMember, error) {
	me, err := c.Me()
	if err != nil {
		return nil, err
	}
	data, err := c.do("GET", fmt.Sprintf("/users/%s/teams/%s/channels/members", me.ID, teamID), nil)
	if err != nil {
		return nil, err
	}
	var members []ChannelMember
	return members, json.Unmarshal(data, &members)
}

// GetChannel returns a single channel by ID.
func (c *Client) GetChannel(channelID string) (*Channel, error) {
	data, err := c.do("GET", fmt.Sprintf("/channels/%s", channelID), nil)
	if err != nil {
		return nil, err
	}
	var ch Channel
	return &ch, json.Unmarshal(data, &ch)
}

// SearchPosts searches for posts in a team.
func (c *Client) SearchPosts(teamID string, terms string) (*PostList, error) {
	payload := map[string]any{
		"terms":      terms,
		"is_or_search": false,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	data, err := c.do("POST", fmt.Sprintf("/teams/%s/posts/search", teamID), strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	var pl PostList
	return &pl, json.Unmarshal(data, &pl)
}

// IsMuted returns true if the channel member has muted the channel.
func (m *ChannelMember) IsMuted() bool {
	return m.NotifyProps.MarkUnread == "mention"
}

// GetUserByUsername returns a user by username.
func (c *Client) GetUserByUsername(username string) (*User, error) {
	data, err := c.do("GET", fmt.Sprintf("/users/username/%s", username), nil)
	if err != nil {
		return nil, err
	}
	var u User
	return &u, json.Unmarshal(data, &u)
}

// GetOrCreateDM returns the DM channel between the current user and another user.
func (c *Client) GetOrCreateDM(otherUserID string) (*Channel, error) {
	me, err := c.Me()
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal([]string{me.ID, otherUserID})
	if err != nil {
		return nil, err
	}
	data, err := c.do("POST", "/channels/direct", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	var ch Channel
	return &ch, json.Unmarshal(data, &ch)
}
