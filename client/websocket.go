package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// WSEvent represents a parsed Mattermost WebSocket event.
type WSEvent struct {
	Event     string          `json:"event"`
	Data      json.RawMessage `json:"data"`
	Broadcast WSBroadcast     `json:"broadcast"`
	Seq       int             `json:"seq"`
}

// WSBroadcast holds the broadcast routing info from a WS event.
type WSBroadcast struct {
	ChannelID string `json:"channel_id"`
	TeamID    string `json:"team_id"`
	UserID    string `json:"user_id"`
}

// WSConn wraps a gorilla/websocket connection to Mattermost.
type WSConn struct {
	conn *websocket.Conn
}

// Connect establishes a WebSocket connection and authenticates.
func (c *Client) Connect(ctx context.Context) (*WSConn, error) {
	wsURL := c.wsURL()

	dialer := websocket.Dialer{
		TLSClientConfig: c.tlsCfg,
	}

	conn, _, err := dialer.DialContext(ctx, wsURL, http.Header{})
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", wsURL, err)
	}

	// Send authentication challenge
	auth := struct {
		Seq    int `json:"seq"`
		Action string `json:"action"`
		Data   struct {
			Token string `json:"token"`
		} `json:"data"`
	}{
		Seq:    1,
		Action: "authentication_challenge",
	}
	auth.Data.Token = c.token

	if err := conn.WriteJSON(auth); err != nil {
		conn.Close()
		return nil, fmt.Errorf("send auth: %w", err)
	}

	// Read auth reply — skip any event frames (e.g. "hello") until we get the response
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("read auth reply: %w", err)
		}
		var reply struct {
			Status   string `json:"status"`
			SeqReply int    `json:"seq_reply"`
			Event    string `json:"event"`
			Error    *struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(raw, &reply); err != nil {
			conn.Close()
			return nil, fmt.Errorf("parse auth reply: %w", err)
		}
		// Skip event frames (like "hello") until we get the auth response
		if reply.Event != "" {
			continue
		}
		if reply.Status != "OK" {
			conn.Close()
			msg := "unknown error"
			if reply.Error != nil {
				msg = reply.Error.Message
			}
			return nil, fmt.Errorf("auth failed: %s (raw: %s)", msg, string(raw))
		}
		break
	}

	return &WSConn{conn: conn}, nil
}

// ReadEvent blocks until the next event is available or the connection closes.
func (ws *WSConn) ReadEvent() (*WSEvent, error) {
	var evt WSEvent
	if err := ws.conn.ReadJSON(&evt); err != nil {
		return nil, err
	}
	return &evt, nil
}

// Close closes the underlying WebSocket connection.
func (ws *WSConn) Close() error {
	return ws.conn.Close()
}

// wsURL derives the WebSocket URL from the REST base URL.
// http://host:port/api/v4 → ws://host:port/api/v4/websocket
// https://host/api/v4     → wss://host/api/v4/websocket
func (c *Client) wsURL() string {
	u := c.baseURL
	u = strings.Replace(u, "https://", "wss://", 1)
	u = strings.Replace(u, "http://", "ws://", 1)
	return strings.TrimSuffix(u, "/api/v4") + "/api/v4/websocket"
}
