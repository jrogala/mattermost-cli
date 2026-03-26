# Mattermost CLI — Websocket Integration Plan

## Goal

Add a `listen` command to mattermost-cli that connects to the Mattermost websocket and streams events in real-time. This is step 1 toward a Bubble Tea TUI.

## Websocket API

**Endpoint**: `wss://mattermost.gatewatcher.fr/api/v4/websocket`

**Auth**: Bearer token (same token from keyring used for REST API)

**Official Go client**: `github.com/mattermost/mattermost/server/public/model` provides `WebSocketClient`:

```go
import "github.com/mattermost/mattermost/server/public/model"

client, err := model.NewWebSocketClient4("wss://mattermost.gatewatcher.fr", token)
client.Listen()

for event := range client.EventChannel {
    // handle event
}
```

The client exposes 3 channels:
- `EventChannel chan *WebSocketEvent` — all server-pushed events
- `ResponseChannel chan *WebSocketResponse` — responses to client requests
- `PingTimeoutChannel chan bool` — connection health

## Key Event Types

| Event | Description |
|-------|-------------|
| `posted` | New message posted (contains full post JSON) |
| `typing` | User is typing in a channel |
| `post_edited` | Message was edited |
| `post_deleted` | Message was deleted |
| `channel_updated` | Channel metadata changed |
| `status_change` | User online/offline/away |
| `reaction_added` / `reaction_removed` | Emoji reactions |
| `channel_viewed` | Someone viewed a channel |
| `direct_added` | Added to a DM |

## Implementation

### Step 1: Add websocket to client package

In `client/client.go`, add a method to create a websocket connection:

```go
func (c *Client) NewWebSocket() (*model.WebSocketClient, error) {
    wsURL := "wss://" + c.URL // or derive from existing config
    ws, err := model.NewWebSocketClient4(wsURL, c.Token)
    if err != nil {
        return nil, err
    }
    ws.Listen()
    return ws, nil
}
```

### Step 2: Add `listen` command

New file: `cmd/listen.go`

```go
// cmd/listen.go
// Usage: mattermost-cli listen [--channel <id>] [--events posted,typing]

func runListen(cmd *cobra.Command, args []string) error {
    client := getClient()
    ws, err := client.NewWebSocket()
    if err != nil {
        return err
    }
    defer ws.Close()

    filterEvents := getEventFilter(cmd) // optional: only show certain events
    filterChannel := getChannelFilter(cmd) // optional: only show certain channels

    for {
        select {
        case event := <-ws.EventChannel:
            if shouldDisplay(event, filterEvents, filterChannel) {
                printEvent(event)
            }
        case <-ws.PingTimeoutChannel:
            // reconnect logic
        }
    }
}
```

### Step 3: Format output

For CLI mode, output events as structured lines:

```
[16:42:03] #general | jimmy.rogala: hello world
[16:42:05] #general | moran.abadie is typing...
[16:42:08] #general | moran.abadie: salut !
```

With `--json` flag, output raw JSON for piping.

### Step 4 (future): Use as foundation for TUI

The websocket goroutine feeds a `chan Event` that Bubble Tea's `tea.Program` can consume via `tea.Cmd`. The same `listen` logic becomes the backbone of the real-time TUI.

## Dependencies to add

```
go get github.com/mattermost/mattermost/server/public/model
```

Note: this is a large dependency (mattermost server model). Alternative: use `github.com/gorilla/websocket` directly and handle the Mattermost websocket protocol manually (simpler, lighter):

```go
conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{
    "Authorization": {"Bearer " + token},
})
// Read JSON frames, parse event_type field
```

The raw approach is ~50 lines and avoids pulling in the entire mattermost server model package.

## Recommended approach

Use raw `gorilla/websocket` — your CLI is lightweight (3.8k lines), no need to pull the full mattermost model package. The websocket protocol is simple JSON frames:

```json
{"event":"posted","data":{"channel_display_name":"General","channel_name":"general","channel_type":"O","post":"{...json...}","sender_name":"@jimmy.rogala"},"broadcast":{"channel_id":"abc123"},"seq":42}
```

Parse `event` field, extract `data.post` for messages, `data.sender_name` for who.
