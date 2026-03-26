package tui

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
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

// channelItem implements list.Item for the sidebar.
type channelItem struct {
	channel ops.Channel
}

func (c channelItem) FilterValue() string { return c.channel.DisplayName }
func (c channelItem) Title() string {
	name := c.channel.DisplayName
	if c.channel.Unread > 0 {
		return fmt.Sprintf("%s (%d)", name, c.channel.Unread)
	}
	return name
}
func (c channelItem) Description() string { return "" }

// channelDelegate renders a channel item in the list.
type channelDelegate struct{}

func (d channelDelegate) Height() int                             { return 1 }
func (d channelDelegate) Spacing() int                            { return 0 }
func (d channelDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d channelDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ci, ok := item.(channelItem)
	if !ok {
		return
	}

	name := ci.channel.DisplayName
	if len(name) > sidebarWidth-6 {
		name = name[:sidebarWidth-9] + "..."
	}

	badge := ""
	if ci.channel.Unread > 0 {
		badge = fmt.Sprintf(" (%d)", ci.channel.Unread)
	}

	var line string
	if index == m.Index() {
		line = channelSelectedStyle.Render("▸ " + name + badge)
	} else if ci.channel.Unread > 0 {
		line = channelUnreadStyle.Render("  " + name + badge)
	} else {
		line = channelStyle.Render("  " + name)
	}

	fmt.Fprint(w, line)
}

// Model is the Bubble Tea model for the TUI.
type Model struct {
	client *client.Client
	me     *ops.UserInfo
	err    error

	// Sidebar
	channels []ops.Channel
	list     list.Model

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

	l := list.New(nil, channelDelegate{}, sidebarWidth, 10)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	return Model{
		client: c,
		input:  ti,
		list:   l,
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

// selectedChannel returns the currently selected channel or nil.
func (m Model) selectedChannel() *ops.Channel {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	ci := item.(channelItem)
	return &ci.channel
}

// selectedChannelID returns the current channel ID or empty.
func (m Model) selectedChannelID() string {
	if ch := m.selectedChannel(); ch != nil {
		return ch.ID
	}
	return ""
}
