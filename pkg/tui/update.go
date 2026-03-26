package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

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

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.list.SetHeight(msg.Height)
		contentW := msg.Width - sidebarWidth - 2
		vpH := msg.Height - 4
		if m.showHelp {
			vpH--
		}
		if vpH < 1 {
			vpH = 1
		}
		m.viewport.Width = contentW
		m.viewport.Height = vpH
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
		return m, nil

	case meLoadedMsg:
		m.me = msg.me
		return m, nil

	case channelsLoadedMsg:
		m.channels = msg.channels
		items := buildListItems(m.channels, m.sidebarMode)
		cmd := m.list.SetItems(items)
		cmds = append(cmds, cmd)
		// Select first actual channel (skip section header)
		m.list.Select(1)
		if ch := m.selectedChannel(); ch != nil {
			cmds = append(cmds, loadMessages(m.client, ch.ID))
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

	case dismissNotificationMsg:
		m.notification = ""
		return m, nil

	case debounceLoadMsg:
		if msg.channelID == m.selectedChannelID() {
			return m, loadMessages(m.client, msg.channelID)
		}
		return m, nil

	case messagesLoadedMsg:
		// Discard stale responses from fast channel switching
		if msg.channelID != m.selectedChannelID() {
			return m, nil
		}
		m.messages = msg.messages
		if len(m.messages) > 0 {
			m.selectedMsg = len(m.messages) - 1
		} else {
			m.selectedMsg = 0
		}
		// Clear unread in sidebar
		for i := range m.channels {
			if m.channels[i].ID == msg.channelID {
				m.channels[i].Unread = 0
				m.channels[i].Mentions = 0
				m.updateChannelInList(m.channels[i])
				break
			}
		}
		if m.ready {
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
		}
		// Mark as viewed on server
		return m, viewChannel(m.client, msg.channelID)

	case wsEventMsg:
		if cmd := m.handleWSEvent(msg.msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
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
	// Detail view captures all keys except ctrl+c
	if m.showDetail {
		if msg.String() == "ctrl+c" {
			if m.wsCancel != nil {
				m.wsCancel()
			}
			return m, tea.Quit
		}
		return m.handleDetailKey(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		if m.wsCancel != nil {
			m.wsCancel()
		}
		return m, tea.Quit

	case "ctrl+@": // Ctrl+Space: jump to next unread channel
		nextID := m.nextUnreadChannelID()
		if nextID == "" {
			return m, nil
		}
		items := buildListItems(m.channels, m.sidebarMode)
		m.list.SetItems(items)
		for i, item := range m.list.Items() {
			if ci, ok := item.(channelItem); ok && ci.channel.ID == nextID {
				m.list.Select(i)
				break
			}
		}
		m.pane = paneSidebar
		m.input.Blur()
		m.messages = nil
		m.selectedMsg = 0
		m.viewport.SetContent(statusStyle.Render("Loading..."))
		return m, loadMessages(m.client, nextID)

	case "tab":
		switch m.pane {
		case paneSidebar:
			m.pane = paneMessages
			if m.ready && len(m.messages) > 0 {
				m.syncMessagesViewport()
			}
		case paneMessages:
			m.pane = paneInput
			m.input.Focus()
			if m.ready {
				m.viewport.SetContent(m.renderMessages())
			}
		case paneInput:
			m.pane = paneSidebar
			m.input.Blur()
		}
		return m, nil

	case "f1":
		m.showHelp = !m.showHelp
		vpH := m.height - 4
		if m.showHelp {
			vpH--
		}
		if vpH < 1 {
			vpH = 1
		}
		m.viewport.Height = vpH
		m.viewport.SetContent(m.renderMessages())
		return m, nil

	case "esc":
		if m.pane != paneSidebar {
			m.pane = paneSidebar
			m.input.Blur()
			if m.ready {
				m.viewport.SetContent(m.renderMessages())
			}
		}
		return m, nil
	}

	if m.pane == paneSidebar {
		return m.handleSidebarKey(msg)
	}

	if m.pane == paneMessages {
		return m.handleMessagesKey(msg)
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
			// Reset unread in the source data and list
			for i := range m.channels {
				if m.channels[i].ID == ch.ID {
					m.channels[i].Unread = 0
					m.channels[i].Mentions = 0
					m.updateChannelInList(m.channels[i])
					break
				}
			}
			return m, loadMessages(m.client, ch.ID)
		}
		return m, nil

	case "d":
		if m.sidebarMode == sidebarModeChannels {
			m.sidebarMode = sidebarModeDMs
		} else {
			m.sidebarMode = sidebarModeChannels
		}
		items := buildListItems(m.channels, m.sidebarMode)
		m.list.SetItems(items)
		m.list.Select(1)
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

	// Skip section headers
	if _, ok := m.list.SelectedItem().(sectionItem); ok {
		dir := 1
		if msg.String() == "k" || msg.String() == "up" {
			dir = -1
		}
		next := m.list.Index() + dir
		if next >= 0 && next < len(m.list.Items()) {
			m.list.Select(next)
		} else {
			m.list.Select(prevIndex) // stay put
		}
	}

	// If selection changed, clear messages and debounce-load new channel
	if m.list.Index() != prevIndex {
		if ch := m.selectedChannel(); ch != nil {
			m.messages = nil
			m.selectedMsg = 0
			m.viewport.SetContent(statusStyle.Render("Loading..."))
			chID := ch.ID
			debounce := tea.Tick(time.Second, func(time.Time) tea.Msg {
				return debounceLoadMsg{channelID: chID}
			})
			return m, tea.Batch(cmd, debounce)
		}
	}

	return m, cmd
}

func (m Model) handleMessagesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		if m.wsCancel != nil {
			m.wsCancel()
		}
		return m, tea.Quit
	case " ":
		if len(m.messages) > 0 && m.selectedMsg < len(m.messages) {
			selected := m.messages[m.selectedMsg]
			m.showDetail = true
			content := fmt.Sprintf("%s  %s\n\n%s",
				selected.Time.Format("15:04"),
				msgUserStyle.Render(selected.User),
				selected.Text)
			contentW := m.width - sidebarWidth - 6
			vpH := m.height - 4
			if vpH < 1 {
				vpH = 1
			}
			m.detailVP = viewport.New(contentW, vpH)
			m.detailVP.SetContent(content)
		}
		return m, nil
	case "j", "down":
		if m.selectedMsg < len(m.messages)-1 {
			m.selectedMsg++
			m.syncMessagesViewport()
		}
		return m, nil
	case "k", "up":
		if m.selectedMsg > 0 {
			m.selectedMsg--
			m.syncMessagesViewport()
		}
		return m, nil
	case "pgup":
		m.viewport.HalfPageUp()
		return m, nil
	case "pgdown":
		m.viewport.HalfPageDown()
		return m, nil
	}
	return m, nil
}

func (m Model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", " ", "q":
		m.showDetail = false
		m.viewport.SetContent(m.renderMessages())
		return m, nil
	case "j", "down":
		m.detailVP.ScrollDown(1)
		return m, nil
	case "k", "up":
		m.detailVP.ScrollUp(1)
		return m, nil
	case "pgup":
		m.detailVP.HalfPageUp()
		return m, nil
	case "pgdown":
		m.detailVP.HalfPageDown()
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

// msgLineOffset returns the rendered line number where the message at idx starts.
func (m *Model) msgLineOffset(idx int) int {
	line := 0
	for i := 0; i < idx && i < len(m.messages); i++ {
		text := displayText(m.messages[i].Text)
		line += 1 + strings.Count(text, "\n")
	}
	return line
}

// syncMessagesViewport re-renders messages and scrolls to keep selectedMsg visible.
func (m *Model) syncMessagesViewport() {
	m.viewport.SetContent(m.renderMessages())
	line := m.msgLineOffset(m.selectedMsg)
	top := m.viewport.YOffset
	bottom := top + m.viewport.Height - 1
	if line < top {
		m.viewport.SetYOffset(line)
	} else if line > bottom {
		m.viewport.SetYOffset(line - m.viewport.Height + 1)
	}
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

// nextUnreadChannelID returns the ID of the next unread channel, prioritizing
// mentions first, then DMs, cycling from the current selection.
func (m Model) nextUnreadChannelID() string {
	currentID := m.selectedChannelID()

	var unread []ops.Channel
	for _, ch := range m.channels {
		if ch.Unread > 0 {
			unread = append(unread, ch)
		}
	}
	if len(unread) == 0 {
		return ""
	}

	sort.SliceStable(unread, func(i, j int) bool {
		a, b := unread[i], unread[j]
		if (a.Mentions > 0) != (b.Mentions > 0) {
			return a.Mentions > 0
		}
		aIsDM := a.Type == "dm" || a.Type == "group"
		bIsDM := b.Type == "dm" || b.Type == "group"
		return aIsDM && !bIsDM
	})

	currentIdx := -1
	for i, ch := range unread {
		if ch.ID == currentID {
			currentIdx = i
			break
		}
	}
	return unread[(currentIdx+1)%len(unread)].ID
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}

	x, y := msg.X, msg.Y

	// Sidebar click
	if x <= sidebarWidth+1 {
		offset := m.list.Index() - m.list.Cursor()
		idx := offset + y
		if idx >= 0 && idx < len(m.list.Items()) {
			if _, ok := m.list.Items()[idx].(sectionItem); ok {
				return m, nil
			}
			m.list.Select(idx)
			m.pane = paneSidebar
			m.input.Blur()
			if ch := m.selectedChannel(); ch != nil {
				m.messages = nil
				m.selectedMsg = 0
				m.viewport.SetContent(statusStyle.Render("Loading..."))
				return m, loadMessages(m.client, ch.ID)
			}
		}
		return m, nil
	}

	// Content area clicks
	vpH := m.height - 4
	if m.showHelp {
		vpH--
	}

	// Input area: below the viewport
	if y > vpH {
		m.pane = paneInput
		m.input.Focus()
		m.viewport.SetContent(m.renderMessages())
		return m, nil
	}

	// Header (Y=0)
	if y == 0 {
		return m, nil
	}

	// Messages area: switch to messages pane and select clicked message
	vpLine := m.viewport.YOffset + (y - 1)
	msgIdx := m.lineToMsgIndex(vpLine)
	if msgIdx >= 0 && msgIdx < len(m.messages) {
		m.pane = paneMessages
		m.input.Blur()
		m.selectedMsg = msgIdx
		m.viewport.SetContent(m.renderMessages())
	}
	return m, nil
}

// lineToMsgIndex converts a viewport line number to a message index.
func (m Model) lineToMsgIndex(line int) int {
	accumulated := 0
	for i, msg := range m.messages {
		text := displayText(msg.Text)
		lines := 1 + strings.Count(text, "\n")
		if accumulated+lines > line {
			return i
		}
		accumulated += lines
	}
	if len(m.messages) > 0 {
		return len(m.messages) - 1
	}
	return 0
}

func (m *Model) handleWSEvent(msg ops.Message) tea.Cmd {
	chID := m.selectedChannelID()
	if msg.ChannelID == chID {
		m.messages = append(m.messages, msg)
		if m.ready {
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()
		}
		return nil
	}

	// Increment unread in sidebar
	for i := range m.channels {
		if m.channels[i].ID == msg.ChannelID {
			m.channels[i].Unread++
			m.updateChannelInList(m.channels[i])

			// Notification for DMs/group messages
			if m.channels[i].Type == "dm" || m.channels[i].Type == "group" {
				text := msg.Text
				if len(text) > 60 {
					text = text[:57] + "..."
				}
				m.notification = "[DM] " + msg.User + ": " + text
				return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
					return dismissNotificationMsg{}
				})
			}
			break
		}
	}
	return nil
}

// updateChannelInList finds the channel in the list items and updates it.
func (m *Model) updateChannelInList(ch ops.Channel) {
	for i, item := range m.list.Items() {
		if ci, ok := item.(channelItem); ok && ci.channel.ID == ch.ID {
			m.list.SetItem(i, channelItem{ch})
			return
		}
	}
}
