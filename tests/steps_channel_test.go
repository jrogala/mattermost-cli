package tests

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/jrogala/mattermost-cli/pkg/ops"
)

// shortID returns a random 12-char hex string.
func shortID() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// --- Channel setup steps ---

// createChannel creates a channel with the given name and type.
// Name is used as both display name and slug base.
func (sc *scenarioCtx) createChannel(name, chanType string) (string, error) {
	if id, ok := sc.channels[name]; ok {
		return id, nil
	}
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-")) + "-" + shortID()
	id, err := sc.env.CreateChannel(slug, name, chanType)
	if err != nil {
		return "", err
	}
	sc.channels[name] = id
	return id, nil
}

// createRandomChannel creates a channel with a random name.
func (sc *scenarioCtx) createRandomChannel(chanType string) (string, string, error) {
	name := shortID()
	id, err := sc.createChannel(name, chanType)
	return name, id, err
}

func (sc *scenarioCtx) iBelongToChannel(name string) error {
	_, err := sc.createChannel(name, "O")
	return err
}

func (sc *scenarioCtx) iBelongToChannelMuted(name string) error {
	id, err := sc.createChannel(name, "O")
	if err != nil {
		return err
	}
	return sc.env.MuteChannel(id)
}

func (sc *scenarioCtx) iBelongToPublicChannel(name string) error {
	_, err := sc.createChannel(name, "O")
	return err
}

func (sc *scenarioCtx) iBelongToPrivateChannel(name string) error {
	_, err := sc.createChannel(name, "P")
	return err
}

func (sc *scenarioCtx) iHaveADMChannelWithUser(username string) error {
	u, ok := sc.env.Users[username]
	if !ok {
		return fmt.Errorf("unknown test user %q", username)
	}
	result, err := ops.FindDMChannel(sc.client, u.Username)
	if err != nil {
		return err
	}
	sc.channels[username+" (DM)"] = result.ChannelID
	return nil
}

func (sc *scenarioCtx) aChannelExists(name string) error {
	_, err := sc.createChannel(name, "O")
	return err
}

func (sc *scenarioCtx) aUserExists(username string) error {
	if _, ok := sc.env.Users[username]; !ok {
		return fmt.Errorf("test user %q not found (available: alice, bob, testbot)", username)
	}
	return nil
}

// --- Channel list steps ---

func (sc *scenarioCtx) iListMyChannels() error {
	list, err := ops.ListChannels(sc.client, ops.ListOptions{})
	sc.lastErr = err
	sc.channelList = list
	return nil
}

func (sc *scenarioCtx) iListMyChannelsIncludingMuted() error {
	list, err := ops.ListChannels(sc.client, ops.ListOptions{IncludeMuted: true})
	sc.lastErr = err
	sc.channelList = list
	return nil
}

func (sc *scenarioCtx) iListMyChannelsFilteredByType(typeFilter string) error {
	list, err := ops.ListChannels(sc.client, ops.ListOptions{Type: typeFilter})
	sc.lastErr = err
	sc.channelList = list
	return nil
}

func (sc *scenarioCtx) resultShouldIncludeChannel(name string) error {
	expectedID, ok := sc.channels[name]
	if !ok {
		return fmt.Errorf("channel %q was not created in this scenario", name)
	}
	list, ok := sc.channelList.([]ops.ChannelEntry)
	if !ok {
		return fmt.Errorf("no channel list available")
	}
	for _, e := range list {
		if e.ID == expectedID {
			return nil
		}
	}
	return fmt.Errorf("channel %q (id=%s) not found in list", name, expectedID)
}

func (sc *scenarioCtx) resultShouldNotIncludeChannel(name string) error {
	expectedID, ok := sc.channels[name]
	if !ok {
		return nil // not created = not in list
	}
	list, ok := sc.channelList.([]ops.ChannelEntry)
	if !ok {
		return fmt.Errorf("no channel list available")
	}
	for _, e := range list {
		if e.ID == expectedID {
			return fmt.Errorf("channel %q should not be in list but was found", name)
		}
	}
	return nil
}

func (sc *scenarioCtx) resultShouldIncludeDMWith(username string) error {
	list, ok := sc.channelList.([]ops.ChannelEntry)
	if !ok {
		return fmt.Errorf("no channel list available")
	}
	expected := username + " (DM)"
	for _, e := range list {
		if e.DisplayName == expected {
			return nil
		}
	}
	return fmt.Errorf("DM with %q not found in list", username)
}

// --- Channel find steps ---

func (sc *scenarioCtx) iSearchForChannel(term string) error {
	results, err := ops.FindChannels(sc.client, term, "")
	sc.lastErr = err
	sc.findResults = results
	return nil
}

func (sc *scenarioCtx) iSearchForDMChannelWithUser(username string) error {
	result, err := ops.FindDMChannel(sc.client, username)
	sc.lastErr = err
	sc.dmResult = result
	return nil
}

func (sc *scenarioCtx) iShouldFindExactlyNChannels(n int) error {
	results, ok := sc.findResults.([]ops.FindResult)
	if !ok {
		return fmt.Errorf("no find results available")
	}
	if len(results) != n {
		return fmt.Errorf("expected %d channels, got %d", n, len(results))
	}
	return nil
}

func (sc *scenarioCtx) iShouldFindChannel(name string) error {
	expectedID := sc.channels[name]
	results, ok := sc.findResults.([]ops.FindResult)
	if !ok {
		return fmt.Errorf("no find results available")
	}
	for _, r := range results {
		if r.ID == expectedID {
			return nil
		}
	}
	return fmt.Errorf("channel %q not found in results", name)
}

func (sc *scenarioCtx) foundChannelShouldHaveAnID() error {
	results, ok := sc.findResults.([]ops.FindResult)
	if !ok || len(results) == 0 {
		return fmt.Errorf("no find results available")
	}
	if results[0].ID == "" {
		return fmt.Errorf("found channel has empty ID")
	}
	return nil
}

func (sc *scenarioCtx) iShouldGetADMChannel() error {
	if sc.lastErr != nil {
		return fmt.Errorf("expected DM channel but got error: %v", sc.lastErr)
	}
	if sc.dmResult == nil {
		return fmt.Errorf("no DM result available")
	}
	return nil
}

func (sc *scenarioCtx) dmChannelShouldHaveAnID() error {
	result, ok := sc.dmResult.(*ops.DMResult)
	if !ok || result == nil {
		return fmt.Errorf("no DM result available")
	}
	if result.ChannelID == "" {
		return fmt.Errorf("DM channel has empty ID")
	}
	return nil
}

// --- Channel read steps ---

func (sc *scenarioCtx) aChannelWithMessages(name string, table *godog.Table) error {
	id, err := sc.createChannel(name, "O")
	if err != nil {
		return err
	}

	for i, row := range table.Rows {
		if i == 0 {
			continue // header
		}
		username := row.Cells[0].Value
		message := row.Cells[1].Value

		u, ok := sc.env.Users[username]
		if !ok {
			return fmt.Errorf("unknown test user %q", username)
		}

		postID, err := sc.env.PostMessage(u.Token, id, message)
		if err != nil {
			return fmt.Errorf("post message from %s: %w", username, err)
		}

		// If there's an "age" column, set the timestamp
		if len(row.Cells) > 2 {
			age := row.Cells[2].Value
			dur, err := parseAge(age)
			if err != nil {
				return err
			}
			ts := time.Now().Add(-dur)
			if err := sc.env.SetPostTimestamp(postID, ts); err != nil {
				return fmt.Errorf("set timestamp: %w", err)
			}
		}
	}
	return nil
}

func (sc *scenarioCtx) aChannelWithNMessages(name string, n int) error {
	id, err := sc.createChannel(name, "O")
	if err != nil {
		return err
	}
	alice := sc.env.Users["alice"]
	for i := 0; i < n; i++ {
		if _, err := sc.env.PostMessage(alice.Token, id, fmt.Sprintf("Message %d", i+1)); err != nil {
			return err
		}
	}
	return nil
}

func (sc *scenarioCtx) aChannelWithNoMessages(name string) error {
	_, err := sc.createChannel(name, "O")
	return err
}

func (sc *scenarioCtx) iReadMessagesFromChannel(name string) error {
	id, ok := sc.channels[name]
	if !ok {
		return fmt.Errorf("channel %q not set up", name)
	}
	msgs, err := ops.ReadMessages(sc.client, id, ops.ReadOptions{})
	sc.lastErr = err
	sc.messages = msgs
	return nil
}

func (sc *scenarioCtx) iReadMessagesFromChannelWithLimit(name string, limit int) error {
	id, ok := sc.channels[name]
	if !ok {
		return fmt.Errorf("channel %q not set up", name)
	}
	msgs, err := ops.ReadMessages(sc.client, id, ops.ReadOptions{Limit: limit})
	sc.lastErr = err
	sc.messages = msgs
	return nil
}

func (sc *scenarioCtx) iReadMessagesFromChannelSince(name, since string) error {
	id, ok := sc.channels[name]
	if !ok {
		return fmt.Errorf("channel %q not set up", name)
	}

	// Handle "yesterday" as a special case
	sinceVal := since
	if since == "yesterday" {
		sinceVal = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}

	msgs, err := ops.ReadMessages(sc.client, id, ops.ReadOptions{Since: sinceVal})
	sc.lastErr = err
	sc.messages = msgs
	return nil
}

func (sc *scenarioCtx) iReadMessagesFromChannelNamed(name string) error {
	msgs, err := ops.ReadMessagesByName(sc.client, name, ops.ReadOptions{})
	sc.lastErr = err
	sc.messages = msgs
	return nil
}

func (sc *scenarioCtx) iShouldReceiveNMessages(n int) error {
	msgs, ok := sc.messages.([]ops.Message)
	if !ok {
		if n == 0 && sc.messages == nil {
			return nil
		}
		return fmt.Errorf("no messages available")
	}
	if len(msgs) != n {
		return fmt.Errorf("expected %d messages, got %d", n, len(msgs))
	}
	return nil
}

func (sc *scenarioCtx) iShouldReceiveAtMostNMessages(n int) error {
	msgs, ok := sc.messages.([]ops.Message)
	if !ok {
		return nil // no messages = 0 <= n
	}
	if len(msgs) > n {
		return fmt.Errorf("expected at most %d messages, got %d", n, len(msgs))
	}
	return nil
}

func (sc *scenarioCtx) messagesShouldContain(text string) error {
	msgs, ok := sc.messages.([]ops.Message)
	if !ok {
		return fmt.Errorf("no messages available")
	}
	for _, m := range msgs {
		if strings.Contains(m.Text, text) {
			return nil
		}
	}
	return fmt.Errorf("no message contains %q", text)
}

func (sc *scenarioCtx) messagesShouldNotContain(text string) error {
	msgs, ok := sc.messages.([]ops.Message)
	if !ok {
		return nil // no messages, so doesn't contain
	}
	for _, m := range msgs {
		if strings.Contains(m.Text, text) {
			return fmt.Errorf("message should not contain %q but does: %q", text, m.Text)
		}
	}
	return nil
}

// --- Channel send steps ---

func (sc *scenarioCtx) iSendToChannel(message, name string) error {
	id, ok := sc.channels[name]
	if !ok {
		return fmt.Errorf("channel %q not set up", name)
	}
	result, err := ops.SendMessage(sc.client, id, message)
	sc.lastErr = err
	sc.sendResult = result
	return nil
}

func (sc *scenarioCtx) iSendToChannelNamed(message, name string) error {
	result, err := ops.SendMessageByName(sc.client, name, message)
	sc.lastErr = err
	sc.sendResult = result
	return nil
}

func (sc *scenarioCtx) messageShouldBePosted() error {
	if sc.lastErr != nil {
		return fmt.Errorf("expected message to be posted but got error: %v", sc.lastErr)
	}
	if sc.sendResult == nil {
		return fmt.Errorf("no send result available")
	}
	return nil
}

func (sc *scenarioCtx) postedMessageShouldHaveAnID() error {
	result, ok := sc.sendResult.(*ops.SendResult)
	if !ok || result == nil {
		return fmt.Errorf("no send result available")
	}
	if result.ID == "" {
		return fmt.Errorf("posted message has empty ID")
	}
	return nil
}

// --- helpers ---

func parseAge(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid age %q", s)
	}

	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	var n int
	if _, err := fmt.Sscanf(numStr, "%d", &n); err != nil {
		return 0, fmt.Errorf("invalid age %q: %w", s, err)
	}

	switch unit {
	case 's':
		return time.Duration(n) * time.Second, nil
	case 'm':
		return time.Duration(n) * time.Minute, nil
	case 'h':
		return time.Duration(n) * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown age unit %q", string(unit))
	}
}
