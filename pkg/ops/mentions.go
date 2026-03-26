package ops

import (
	"fmt"
	"sort"
	"time"

	"github.com/jrogala/mattermost-cli/client"
)

// MentionOptions configures the mentions query.
type MentionOptions struct {
	TeamID string
	Limit  int // max messages per channel with mentions, 0 = default (5)
}

// GetMentions returns recent messages that mention the authenticated user.
// Unlike other feeds, mentions always include muted channels.
func GetMentions(c *client.Client, opts MentionOptions) ([]Message, error) {
	me, err := c.Me()
	if err != nil {
		return nil, err
	}

	teamID, err := ResolveTeam(c, opts.TeamID)
	if err != nil {
		return nil, err
	}

	members, err := c.GetChannelMembers(teamID)
	if err != nil {
		return nil, err
	}

	var withMentions []client.ChannelMember
	for _, m := range members {
		if m.MentionCount > 0 {
			withMentions = append(withMentions, m)
		}
	}

	if len(withMentions) == 0 {
		return nil, nil
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 5
	}

	chanCache := map[string]string{}
	resolveChan := func(chanID string) string {
		if name, ok := chanCache[chanID]; ok {
			return name
		}
		ch, err := c.GetChannel(chanID)
		if err != nil {
			chanCache[chanID] = chanID[:8]
			return chanCache[chanID]
		}
		name := ResolveChannelName(c, *ch, me.ID)
		chanCache[chanID] = name
		return name
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

	var mentions []Message
	username := me.Username

	for _, m := range withMentions {
		pl, err := c.GetChannelPosts(m.ChannelID, map[string]string{
			"per_page": fmt.Sprintf("%d", limit),
		})
		if err != nil {
			continue
		}

		chanName := resolveChan(m.ChannelID)
		for _, pid := range pl.Order {
			post := pl.Posts[pid]
			if !containsMention(post.Message, username) {
				continue
			}
			mentions = append(mentions, Message{
				ID:        post.ID,
				ChannelID: m.ChannelID,
				Channel:   chanName,
				User:      resolveUser(post.UserID),
				UserID:    post.UserID,
				Text:      post.Message,
				Time:      time.UnixMilli(post.CreateAt),
			})
		}
	}

	sort.Slice(mentions, func(i, j int) bool {
		return mentions[i].Time.After(mentions[j].Time)
	})

	return mentions, nil
}

func containsMention(message, username string) bool {
	target := "@" + username
	for i := 0; i <= len(message)-len(target); i++ {
		if message[i:i+len(target)] == target {
			return true
		}
	}
	return false
}
