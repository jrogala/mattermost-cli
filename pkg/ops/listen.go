package ops

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/jrogala/mattermost-cli/client"
)

// ListenEvent is the ops-layer representation of a WebSocket event.
type ListenEvent struct {
	Event     string          `json:"event"`
	ChannelID string          `json:"channel_id"`
	Sender    string          `json:"sender,omitempty"`
	Message   string          `json:"message,omitempty"`
	PostID    string          `json:"post_id,omitempty"`
	PostType  string          `json:"post_type,omitempty"`
	Channel   string          `json:"channel_name,omitempty"`
	Time      time.Time       `json:"time"`
	Raw       json.RawMessage `json:"raw,omitempty"`
}

// ListenOptions configures the listen stream.
type ListenOptions struct {
	EventTypes []string // filter: only these event types (empty = all)
	ChannelIDs []string // filter: only these channel IDs (empty = all)
}

// Listen connects via WebSocket and returns a channel of events.
// The channel closes when ctx is canceled or the connection fails permanently.
func Listen(ctx context.Context, c *client.Client, opts ListenOptions) (<-chan ListenEvent, <-chan error, error) {
	ws, err := c.Connect(ctx)
	if err != nil {
		return nil, nil, err
	}

	events := make(chan ListenEvent)
	errs := make(chan error, 1)

	eventSet := toSet(opts.EventTypes)
	channelSet := toSet(opts.ChannelIDs)

	go func() {
		defer close(events)
		defer close(errs)
		defer ws.Close()

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			raw, err := ws.ReadEvent()
			if err != nil {
				select {
				case <-ctx.Done():
				case errs <- err:
				}
				return
			}

			if raw.Event == "" {
				continue // heartbeat or response frame
			}

			if len(eventSet) > 0 && !eventSet[raw.Event] {
				continue
			}

			if len(channelSet) > 0 && !channelSet[raw.Broadcast.ChannelID] {
				continue
			}

			evt := parseEvent(raw)

			// Skip system posts (join/leave notifications etc.)
			if evt.PostType != "" {
				continue
			}

			select {
			case events <- evt:
			case <-ctx.Done():
				return
			}
		}
	}()

	return events, errs, nil
}

func parseEvent(raw *client.WSEvent) ListenEvent {
	evt := ListenEvent{
		Event:     raw.Event,
		ChannelID: raw.Broadcast.ChannelID,
		Time:      time.Now(),
		Raw:       raw.Data,
	}

	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw.Data, &data); err != nil {
		return evt
	}

	// Extract sender_name (strip leading @)
	if sn, ok := data["sender_name"]; ok {
		var sender string
		if json.Unmarshal(sn, &sender) == nil {
			evt.Sender = strings.TrimPrefix(sender, "@")
		}
	}

	// Extract channel_name
	if cn, ok := data["channel_name"]; ok {
		var name string
		if json.Unmarshal(cn, &name) == nil {
			evt.Channel = name
		}
	}

	// Extract channel_display_name as fallback
	if evt.Channel == "" {
		if cdn, ok := data["channel_display_name"]; ok {
			var name string
			if json.Unmarshal(cdn, &name) == nil {
				evt.Channel = name
			}
		}
	}

	// For posted events, parse the double-encoded post JSON
	if raw.Event == "posted" || raw.Event == "post_edited" {
		if postRaw, ok := data["post"]; ok {
			var postStr string
			if json.Unmarshal(postRaw, &postStr) == nil {
				var post struct {
					ID       string `json:"id"`
					Message  string `json:"message"`
					Type     string `json:"type"`
					CreateAt int64  `json:"create_at"`
				}
				if json.Unmarshal([]byte(postStr), &post) == nil {
					evt.PostID = post.ID
					evt.Message = post.Message
					evt.PostType = post.Type
					if post.CreateAt > 0 {
						evt.Time = time.UnixMilli(post.CreateAt)
					}
				}
			}
		}
	}

	// For post_deleted, extract post ID
	if raw.Event == "post_deleted" {
		if postRaw, ok := data["post"]; ok {
			var postStr string
			if json.Unmarshal(postRaw, &postStr) == nil {
				var post struct {
					ID string `json:"id"`
				}
				if json.Unmarshal([]byte(postStr), &post) == nil {
					evt.PostID = post.ID
				}
			}
		}
	}

	return evt
}

func toSet(items []string) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}
