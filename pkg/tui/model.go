package tui

import (
	"context"
	"fmt"

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
	if c.channel.Muted {
		name += " (muted)"
	}
	if c.channel.Unread > 0 {
		name = fmt.Sprintf("%s [%d]", name, c.channel.Unread)
	}
	return name
}
func (c channelItem) Description() string { return "" }

// sectionItem is a non-selectable header in the list.
type sectionItem struct{ title string }

func (s sectionItem) FilterValue() string { return "" }
func (s sectionItem) Title() string       { return s.title }
func (s sectionItem) Description() string { return "" }

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

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.SetSpacing(0)
	delegate.Styles.NormalTitle = channelStyle
	delegate.Styles.SelectedTitle = channelSelectedStyle
	delegate.Styles.DimmedTitle = channelMutedStyle

	l := list.New(nil, delegate, sidebarWidth, 10)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

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
	ci, ok := item.(channelItem)
	if !ok {
		return nil // section header
	}
	return &ci.channel
}

// selectedChannelID returns the current channel ID or empty.
func (m Model) selectedChannelID() string {
	if ch := m.selectedChannel(); ch != nil {
		return ch.ID
	}
	return ""
}

// buildListItems creates the sidebar items with section headers.
// Channels first, then DMs. DMs without unread are hidden.
func buildListItems(channels []ops.Channel) []list.Item {
	var chItems, dmItems []list.Item

	for _, ch := range channels {
		item := channelItem{ch}
		if ch.Type == "dm" || ch.Type == "group" {
			// Only show DMs with unread messages
			if ch.Unread > 0 {
				dmItems = append(dmItems, item)
			}
		} else {
			chItems = append(chItems, item)
		}
	}

	var items []list.Item
	if len(chItems) > 0 {
		items = append(items, sectionItem{"── Channels ──"})
		items = append(items, chItems...)
	}
	if len(dmItems) > 0 {
		items = append(items, sectionItem{"── Direct Messages ──"})
		items = append(items, dmItems...)
	}

	return items
}
