package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/pkg/ops"
)

const (
	paneSidebar = iota
	paneInput
)

// Model is the Bubble Tea model for the TUI.
type Model struct {
	client *client.Client
	me     *ops.UserInfo
	err    error

	// Sidebar
	channels []ops.Channel
	selected int

	// Messages
	messages []ops.Message
	viewport viewport.Model

	// Input
	input textinput.Model

	// WebSocket
	wsEvents <-chan ops.Message
	wsErrors <-chan error
	wsCancel context.CancelFunc

	// Layout
	width  int
	height int
	pane   int
	ready  bool
}

// tea.Msg types
type channelsLoadedMsg struct{ channels []ops.Channel }
type messagesLoadedMsg struct{ messages []ops.Message }
type wsEventMsg struct{ msg ops.Message }
type wsClosedMsg struct{}
type messageSentMsg struct{}
type errMsg struct{ err error }

// New creates a new TUI model.
func New(c *client.Client) Model {
	ti := textinput.New()
	ti.Placeholder = "type a message..."
	ti.CharLimit = 4000

	return Model{
		client: c,
		input:  ti,
		pane:   paneSidebar,
	}
}

// Init starts loading channels and user info.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		loadChannels(m.client),
		loadMe(m.client),
	)
}

// selectedChannelID returns the current channel ID or empty.
func (m Model) selectedChannelID() string {
	if m.selected >= 0 && m.selected < len(m.channels) {
		return m.channels[m.selected].ID
	}
	return ""
}
