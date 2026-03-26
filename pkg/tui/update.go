package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jrogala/mattermost-cli/pkg/ops"
)

// Update handles messages and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.list.SetHeight(msg.Height)
		m.viewport = viewport.New(msg.Width-sidebarWidth-2, msg.Height-4)
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case meLoadedMsg:
		m.me = msg.me
		return m, nil

	case channelsLoadedMsg:
		m.channels = msg.channels
		items := make([]list.Item, len(msg.channels))
		for i, ch := range msg.channels {
			items[i] = channelItem{ch}
		}
		cmd := m.list.SetItems(items)
		cmds = append(cmds, cmd)
		if len(m.channels) > 0 {
			m.list.Select(0)
			cmds = append(cmds, loadMessages(m.client, m.channels[0].ID))
			// Start WebSocket
			ctx, cancel := context.WithCancel(context.Background())
			m.wsCancel = cancel
			events, errs, err := ops.Listen(ctx, m.client, ops.ListenOptions{})
			if err == nil {
				m.wsEvents = events
				m.wsErrors = errs
				cmds = append(cmds, waitForWS(events))
			}
		}
		return m, tea.Batch(cmds...)

	case messagesLoadedMsg:
		m.messages = msg.messages
		if m.ready {
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
		}
		return m, nil

	case wsEventMsg:
		m.handleWSEvent(msg.msg)
		cmds = append(cmds, waitForWS(m.wsEvents))
		return m, tea.Batch(cmds...)

	case wsClosedMsg:
		return m, nil

	case messageSentMsg:
		m.input.Reset()
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	// Pass to input if focused
	if m.pane == paneInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		if m.wsCancel != nil {
			m.wsCancel()
		}
		return m, tea.Quit

	case "tab":
		if m.pane == paneSidebar {
			m.pane = paneInput
			m.input.Focus()
		} else {
			m.pane = paneSidebar
			m.input.Blur()
		}
		return m, nil

	case "esc":
		if m.pane == paneInput {
			m.pane = paneSidebar
			m.input.Blur()
		}
		return m, nil
	}

	if m.pane == paneSidebar {
		return m.handleSidebarKey(msg)
	}

	if m.pane == paneInput {
		return m.handleInputKey(msg)
	}

	return m, nil
}

func (m Model) handleSidebarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	prevIndex := m.list.Index()

	switch msg.String() {
	case "q":
		if m.wsCancel != nil {
			m.wsCancel()
		}
		return m, tea.Quit

	case "enter":
		if ch := m.selectedChannel(); ch != nil {
			// Reset unread in the channel list
			idx := m.list.Index()
			if idx >= 0 && idx < len(m.channels) {
				m.channels[idx].Unread = 0
				m.channels[idx].Mentions = 0
				m.updateListItem(idx)
			}
			return m, loadMessages(m.client, ch.ID)
		}
		return m, nil

	case "pgup":
		m.viewport.HalfPageUp()
		return m, nil
	case "pgdown":
		m.viewport.HalfPageDown()
		return m, nil
	}

	// Delegate j/k/up/down to the list component
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	// If selection changed, load messages for new channel
	if m.list.Index() != prevIndex {
		if ch := m.selectedChannel(); ch != nil {
			return m, tea.Batch(cmd, loadMessages(m.client, ch.ID))
		}
	}

	return m, cmd
}

func (m Model) handleInputKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		text := m.input.Value()
		if text == "" {
			return m, nil
		}
		chID := m.selectedChannelID()
		if chID == "" {
			return m, nil
		}
		return m, sendMessage(m.client, chID, text)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) handleWSEvent(msg ops.Message) {
	chID := m.selectedChannelID()
	if msg.ChannelID == chID {
		m.messages = append(m.messages, msg)
		if m.ready {
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
		}
	} else {
		// Increment unread in sidebar
		for i := range m.channels {
			if m.channels[i].ID == msg.ChannelID {
				m.channels[i].Unread++
				m.updateListItem(i)
				break
			}
		}
	}
}

func (m *Model) updateListItem(idx int) {
	if idx >= 0 && idx < len(m.channels) {
		m.list.SetItem(idx, channelItem{m.channels[idx]})
	}
}
