package tests

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jrogala/mattermost-cli/pkg/ops"
)

func (sc *scenarioCtx) iAmListeningOnAllEvents() error {
	ctx, cancel := context.WithCancel(context.Background())
	sc.listenCancel = cancel

	ch, errCh, err := ops.Listen(ctx, sc.client, ops.ListenOptions{})
	if err != nil {
		cancel()
		return err
	}
	sc.listenEvents = ch
	sc.listenErrors = errCh
	return nil
}

func (sc *scenarioCtx) iAmListeningForEventsOnly(eventType string) error {
	ctx, cancel := context.WithCancel(context.Background())
	sc.listenCancel = cancel

	ch, errCh, err := ops.Listen(ctx, sc.client, ops.ListenOptions{
		EventTypes: []string{eventType},
	})
	if err != nil {
		cancel()
		return err
	}
	sc.listenEvents = ch
	sc.listenErrors = errCh
	return nil
}

func (sc *scenarioCtx) iAmListeningOnChannelOnly(channelName string) error {
	channelID, ok := sc.channels[channelName]
	if !ok {
		return fmt.Errorf("channel %q not set up", channelName)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sc.listenCancel = cancel

	ch, errCh, err := ops.Listen(ctx, sc.client, ops.ListenOptions{
		ChannelIDs: []string{channelID},
	})
	if err != nil {
		cancel()
		return err
	}
	sc.listenEvents = ch
	sc.listenErrors = errCh
	return nil
}

func (sc *scenarioCtx) alicePostsToChannel(message, channelName string) error {
	channelID, ok := sc.channels[channelName]
	if !ok {
		return fmt.Errorf("channel %q not set up", channelName)
	}
	alice := sc.env.Users["alice"]
	postID, err := sc.env.PostMessage(alice.Token, channelID, message)
	if err != nil {
		return err
	}
	sc.lastPostID = postID
	return nil
}

func (sc *scenarioCtx) thatPostIsDeleted() error {
	if sc.lastPostID == "" {
		return fmt.Errorf("no post to delete")
	}
	return sc.env.DeletePost(sc.env.AdminToken, sc.lastPostID)
}

func (sc *scenarioCtx) iShouldReceiveEventWithinNSeconds(eventType string, seconds int) error {
	timeout := time.Duration(seconds) * time.Second
	deadline := time.After(timeout)

	for {
		select {
		case evt, ok := <-sc.listenEvents:
			if !ok {
				return fmt.Errorf("event channel closed before receiving %q", eventType)
			}
			sc.receivedEvents = append(sc.receivedEvents, evt)
			if evt.Event == eventType {
				return nil
			}
		case <-deadline:
			return fmt.Errorf("timed out after %ds waiting for %q event (received %d events: %s)",
				seconds, eventType, len(sc.receivedEvents), receivedSummary(sc.receivedEvents))
		}
	}
}

func (sc *scenarioCtx) eventMessageShouldContain(text string) error {
	if len(sc.receivedEvents) == 0 {
		return fmt.Errorf("no events received")
	}
	last := sc.receivedEvents[len(sc.receivedEvents)-1]
	if !strings.Contains(last.Message, text) {
		return fmt.Errorf("event message %q does not contain %q", last.Message, text)
	}
	return nil
}

func (sc *scenarioCtx) eventSenderShouldBe(username string) error {
	if len(sc.receivedEvents) == 0 {
		return fmt.Errorf("no events received")
	}
	last := sc.receivedEvents[len(sc.receivedEvents)-1]
	if last.Sender != username {
		return fmt.Errorf("expected sender %q, got %q", username, last.Sender)
	}
	return nil
}

func (sc *scenarioCtx) iShouldNotHaveReceivedEventForChannel(channelName string) error {
	channelID, ok := sc.channels[channelName]
	if !ok {
		return fmt.Errorf("channel %q not set up", channelName)
	}
	// Drain any remaining events briefly
	drainTimeout := time.After(500 * time.Millisecond)
drain:
	for {
		select {
		case evt, ok := <-sc.listenEvents:
			if !ok {
				break drain
			}
			sc.receivedEvents = append(sc.receivedEvents, evt)
		case <-drainTimeout:
			break drain
		}
	}

	for _, evt := range sc.receivedEvents {
		if evt.ChannelID == channelID {
			return fmt.Errorf("received unexpected event for channel %q: %s", channelName, evt.Event)
		}
	}
	return nil
}

func (sc *scenarioCtx) eventShouldHaveNonEmptyChannelID() error {
	if len(sc.receivedEvents) == 0 {
		return fmt.Errorf("no events received")
	}
	last := sc.receivedEvents[len(sc.receivedEvents)-1]
	if last.ChannelID == "" {
		return fmt.Errorf("event has empty channel ID")
	}
	return nil
}

func receivedSummary(events []ops.ListenEvent) string {
	if len(events) == 0 {
		return "none"
	}
	var types []string
	for _, e := range events {
		types = append(types, e.Event)
	}
	return strings.Join(types, ", ")
}
