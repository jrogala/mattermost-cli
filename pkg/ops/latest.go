package ops

import (
	"sort"
	"strconv"
	"time"

	"github.com/jrogala/mattermost-cli/client"
)

// LatestOptions configures the latest messages query.
type LatestOptions struct {
	TeamID     string
	Channels   int // max channels to show, 0 = default (10)
	PerChannel int // messages per channel, 0 = default (3)
}

// GetLatest returns recent messages across non-muted channels.
func GetLatest(c *client.Client, opts LatestOptions) ([]Message, error) {
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

	memberMap := make(map[string]*client.ChannelMember, len(members))
	for i := range members {
		memberMap[members[i].ChannelID] = &members[i]
	}

	var active []client.Channel
	for _, ch := range channels {
		m, ok := memberMap[ch.ID]
		if !ok {
			continue
		}
		if m.IsMuted() {
			continue
		}
		if ch.LastPostAt == 0 {
			continue
		}
		active = append(active, ch)
	}

	sort.Slice(active, func(i, j int) bool {
		return active[i].LastPostAt > active[j].LastPostAt
	})

	limit := opts.Channels
	if limit <= 0 {
		limit = 10
	}
	if limit > 0 && len(active) > limit {
		active = active[:limit]
	}

	perChannel := opts.PerChannel
	if perChannel <= 0 {
		perChannel = 3
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

	var posts []Message
	for _, ch := range active {
		pl, err := c.GetChannelPosts(ch.ID, map[string]string{
			"per_page": strconv.Itoa(perChannel),
		})
		if err != nil {
			continue
		}

		name := ResolveChannelName(c, ch, me.ID)
		for i := len(pl.Order) - 1; i >= 0; i-- {
			post := pl.Posts[pl.Order[i]]
			if post.IsSystem() {
				continue
			}
			posts = append(posts, Message{
				ID:        post.ID,
				ChannelID: ch.ID,
				Channel:   name,
				User:      resolveUser(post.UserID),
				UserID:    post.UserID,
				Text:      post.Message,
				Time:      time.UnixMilli(post.CreateAt),
			})
		}
	}

	return posts, nil
}
