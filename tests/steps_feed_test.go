package tests

import (
	"fmt"
	"strings"

	"github.com/jrogala/mattermost-cli/pkg/ops"
)

// --- Unread steps ---

func (sc *scenarioCtx) channelHasUnreadMessages(name string, n int) error {
	id, err := sc.createChannel(name, "O")
	if err != nil {
		return err
	}
	// View channel first to reset unread count
	if err := sc.env.ViewChannel(id); err != nil {
		return err
	}
	// Post n messages from alice to create unread
	alice := sc.env.Users["alice"]
	for i := 0; i < n; i++ {
		if _, err := sc.env.PostMessage(alice.Token, id, fmt.Sprintf("unread msg %d", i+1)); err != nil {
			return err
		}
	}
	return nil
}

func (sc *scenarioCtx) channelHasUnreadMessagesAndIsMuted(name string, n int) error {
	if err := sc.channelHasUnreadMessages(name, n); err != nil {
		return err
	}
	return sc.env.MuteChannel(sc.channels[name])
}

func (sc *scenarioCtx) channelHasUnreadMessagesAndMentions(name string, unread, mentions int) error {
	id, err := sc.createChannel(name, "O")
	if err != nil {
		return err
	}
	if err := sc.env.ViewChannel(id); err != nil {
		return err
	}
	alice := sc.env.Users["alice"]
	for i := 0; i < unread; i++ {
		msg := fmt.Sprintf("msg %d", i+1)
		if i < mentions {
			msg = fmt.Sprintf("hey @%s check this %d", sc.env.AdminUsername, i+1)
		}
		if _, err := sc.env.PostMessage(alice.Token, id, msg); err != nil {
			return err
		}
	}
	return nil
}

func (sc *scenarioCtx) noChannelsHaveUnreadMessages() error {
	// View all channels to mark everything as read
	list, err := ops.ListChannels(sc.client, ops.ListOptions{IncludeMuted: true})
	if err != nil {
		return err
	}
	for _, ch := range list {
		_ = sc.env.ViewChannel(ch.ID)
	}
	return nil
}

func (sc *scenarioCtx) iGetUnreadChannels() error {
	entries, err := ops.GetUnread(sc.client, ops.UnreadOptions{})
	sc.lastErr = err
	sc.unreadList = entries
	return nil
}

func (sc *scenarioCtx) iGetUnreadChannelsIncludingMuted() error {
	entries, err := ops.GetUnread(sc.client, ops.UnreadOptions{IncludeMuted: true})
	sc.lastErr = err
	sc.unreadList = entries
	return nil
}

func (sc *scenarioCtx) unreadListShouldInclude(name string) error {
	expectedID := sc.channels[name]
	entries, ok := sc.unreadList.([]ops.Channel)
	if !ok {
		return fmt.Errorf("no unread list available")
	}
	for _, e := range entries {
		if e.ID == expectedID {
			return nil
		}
	}
	return fmt.Errorf("channel %q not found in unread list", name)
}

func (sc *scenarioCtx) unreadListShouldNotInclude(name string) error {
	expectedID := sc.channels[name]
	entries, ok := sc.unreadList.([]ops.Channel)
	if !ok {
		return nil
	}
	for _, e := range entries {
		if e.ID == expectedID {
			return fmt.Errorf("channel %q should not be in unread list", name)
		}
	}
	return nil
}

func (sc *scenarioCtx) shouldAppearBeforeInUnreadList(first, second string) error {
	firstID, secondID := sc.channels[first], sc.channels[second]
	entries, ok := sc.unreadList.([]ops.Channel)
	if !ok {
		return fmt.Errorf("no unread list available")
	}
	firstIdx, secondIdx := -1, -1
	for i, e := range entries {
		if e.ID == firstID {
			firstIdx = i
		}
		if e.ID == secondID {
			secondIdx = i
		}
	}
	if firstIdx < 0 {
		return fmt.Errorf("%q not found in unread list", first)
	}
	if secondIdx < 0 {
		return fmt.Errorf("%q not found in unread list", second)
	}
	if firstIdx >= secondIdx {
		return fmt.Errorf("expected %q (idx %d) before %q (idx %d)", first, firstIdx, second, secondIdx)
	}
	return nil
}

func (sc *scenarioCtx) unreadListShouldBeEmpty() error {
	entries, ok := sc.unreadList.([]ops.Channel)
	if ok && len(entries) > 0 {
		return fmt.Errorf("expected empty unread list, got %d entries", len(entries))
	}
	return nil
}

// --- Latest steps ---

func (sc *scenarioCtx) channelHasRecentMessages(name string) error {
	id, err := sc.createChannel(name, "O")
	if err != nil {
		return err
	}
	alice := sc.env.Users["alice"]
	_, err = sc.env.PostMessage(alice.Token, id, "Recent message in "+name)
	return err
}

func (sc *scenarioCtx) channelIsMutedAndHasRecentMessages(name string) error {
	if err := sc.channelHasRecentMessages(name); err != nil {
		return err
	}
	return sc.env.MuteChannel(sc.channels[name])
}

func (sc *scenarioCtx) nChannelsHaveRecentMessages(n int) error {
	alice := sc.env.Users["alice"]
	for i := 0; i < n; i++ {
		name := fmt.Sprintf("RecentCh%d", i+1)
		id, err := sc.createChannel(name, "O")
		if err != nil {
			return err
		}
		if _, err := sc.env.PostMessage(alice.Token, id, "msg in "+name); err != nil {
			return err
		}
	}
	return nil
}

func (sc *scenarioCtx) channelHasNRecentMessages(name string, n int) error {
	id, err := sc.createChannel(name, "O")
	if err != nil {
		return err
	}
	alice := sc.env.Users["alice"]
	for i := 0; i < n; i++ {
		if _, err := sc.env.PostMessage(alice.Token, id, fmt.Sprintf("Latest msg %d", i+1)); err != nil {
			return err
		}
	}
	return nil
}

func (sc *scenarioCtx) noChannelsHaveRecentMessages() error {
	// Nothing to do - fresh scenario with no posts
	return nil
}

func (sc *scenarioCtx) iGetLatestMessages() error {
	posts, err := ops.GetLatest(sc.client, ops.LatestOptions{})
	sc.lastErr = err
	sc.latestList = posts
	return nil
}

func (sc *scenarioCtx) iGetLatestMessagesWithChannelLimit(n int) error {
	posts, err := ops.GetLatest(sc.client, ops.LatestOptions{Channels: n})
	sc.lastErr = err
	sc.latestList = posts
	return nil
}

func (sc *scenarioCtx) iGetLatestMessagesWithPerChannelLimit(n int) error {
	posts, err := ops.GetLatest(sc.client, ops.LatestOptions{PerChannel: n})
	sc.lastErr = err
	sc.latestList = posts
	return nil
}

func (sc *scenarioCtx) latestResultsShouldIncludeChannel(name string) error {
	posts, ok := sc.latestList.([]ops.Message)
	if !ok {
		return fmt.Errorf("no latest results available")
	}
	for _, p := range posts {
		if p.Channel == name {
			return nil
		}
	}
	return fmt.Errorf("channel %q not found in latest results", name)
}

func (sc *scenarioCtx) latestResultsShouldNotIncludeChannel(name string) error {
	posts, ok := sc.latestList.([]ops.Message)
	if !ok {
		return nil
	}
	for _, p := range posts {
		if p.Channel == name {
			return fmt.Errorf("channel %q should not be in latest results", name)
		}
	}
	return nil
}

func (sc *scenarioCtx) latestResultsShouldHaveAtMostNChannels(n int) error {
	posts, ok := sc.latestList.([]ops.Message)
	if !ok {
		return nil
	}
	channels := make(map[string]bool)
	for _, p := range posts {
		channels[p.Channel] = true
	}
	if len(channels) > n {
		return fmt.Errorf("expected at most %d channels, got %d", n, len(channels))
	}
	return nil
}

func (sc *scenarioCtx) sectionShouldHaveAtMostNMessages(name string, n int) error {
	posts, ok := sc.latestList.([]ops.Message)
	if !ok {
		return fmt.Errorf("no latest results available")
	}
	count := 0
	for _, p := range posts {
		if p.Channel == name {
			count++
		}
	}
	if count > n {
		return fmt.Errorf("expected at most %d messages in %q, got %d", n, name, count)
	}
	return nil
}

func (sc *scenarioCtx) latestResultsShouldBeEmpty() error {
	posts, ok := sc.latestList.([]ops.Message)
	if ok && len(posts) > 0 {
		return fmt.Errorf("expected empty latest results, got %d posts", len(posts))
	}
	return nil
}

// --- Mentions steps ---

func (sc *scenarioCtx) channelHasMentionFrom(channelName, username string) error {
	id, err := sc.createChannel(channelName, "O")
	if err != nil {
		return err
	}
	u, ok := sc.env.Users[username]
	if !ok {
		return fmt.Errorf("unknown test user %q", username)
	}
	msg := fmt.Sprintf("hey @%s check this out", sc.env.AdminUsername)
	_, err = sc.env.PostMessage(u.Token, id, msg)
	return err
}

func (sc *scenarioCtx) channelIsMutedAndHasMentionFrom(channelName, username string) error {
	if err := sc.channelHasMentionFrom(channelName, username); err != nil {
		return err
	}
	return sc.env.MuteChannel(sc.channels[channelName])
}

func (sc *scenarioCtx) noChannelsHaveMentions() error {
	// Nothing to do - fresh scenario
	return nil
}

func (sc *scenarioCtx) iGetMyMentions() error {
	mentions, err := ops.GetMentions(sc.client, ops.MentionOptions{})
	sc.lastErr = err
	sc.mentionList = mentions
	return nil
}

func (sc *scenarioCtx) mentionsShouldIncludeMessageFrom(username string) error {
	mentions, ok := sc.mentionList.([]ops.Message)
	if !ok {
		return fmt.Errorf("no mentions available")
	}
	for _, m := range mentions {
		if m.User == username {
			return nil
		}
	}
	return fmt.Errorf("no mention from %q found", username)
}

func (sc *scenarioCtx) mentionsShouldReferenceChannel(name string) error {
	mentions, ok := sc.mentionList.([]ops.Message)
	if !ok {
		return fmt.Errorf("no mentions available")
	}
	for _, m := range mentions {
		if strings.Contains(m.Channel, name) {
			return nil
		}
	}
	return fmt.Errorf("no mention referencing channel %q found", name)
}

func (sc *scenarioCtx) mentionsListShouldBeEmpty() error {
	mentions, ok := sc.mentionList.([]ops.Message)
	if ok && len(mentions) > 0 {
		return fmt.Errorf("expected empty mentions, got %d", len(mentions))
	}
	return nil
}
