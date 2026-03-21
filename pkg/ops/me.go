// Package ops contains business logic for Mattermost operations.
// Functions return Go structs and errors — zero I/O, zero formatting.
package ops

import "github.com/jrogala/mattermost-cli/client"

// UserInfo holds the authenticated user's profile.
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// GetMe returns the authenticated user's profile.
func GetMe(c *client.Client) (*UserInfo, error) {
	u, err := c.Me()
	if err != nil {
		return nil, err
	}
	return &UserInfo{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
	}, nil
}
