package ops

import (
	"sort"

	"github.com/jrogala/mattermost-cli/client"
)

// UnreadEntry represents a channel with unread messages.
type UnreadEntry struct {
	ChannelID   string `json:"channel_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Unread      int64  `json:"unread"`
	Mentions    int    `json:"mentions"`
	Muted       bool   `json:"muted"`
}

// UnreadOptions configures unread listing.
type UnreadOptions struct {
	TeamID       string
	IncludeMuted bool
}

// GetUnread returns channels with unread messages, sorted by mentions then unread count.
func GetUnread(c *client.Client, opts UnreadOptions) ([]UnreadEntry, error) {
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

	var entries []UnreadEntry
	for _, ch := range channels {
		m, ok := memberMap[ch.ID]
		if !ok {
			continue
		}
		unread := ch.TotalMsgCount - m.MsgCount
		if unread <= 0 && m.MentionCount <= 0 {
			continue
		}
		if m.IsMuted() && !opts.IncludeMuted {
			continue
		}
		entries = append(entries, UnreadEntry{
			ChannelID:   ch.ID,
			Name:        ch.Name,
			DisplayName: ResolveChannelName(c, ch, me.ID),
			Type:        channelTypeName(ch.Type),
			Unread:      unread,
			Mentions:    m.MentionCount,
			Muted:       m.IsMuted(),
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Mentions != entries[j].Mentions {
			return entries[i].Mentions > entries[j].Mentions
		}
		return entries[i].Unread > entries[j].Unread
	})

	return entries, nil
}
