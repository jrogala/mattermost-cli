package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the TUI.
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress q to quit.", m.err)
	}

	if !m.ready {
		return "Loading..."
	}

	sidebar := m.renderSidebar()
	content := m.renderContent()

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
}

func (m Model) renderSidebar() string {
	var sb strings.Builder

	sbStyle := sidebarStyle
	if m.pane == paneSidebar {
		sbStyle = sidebarActiveStyle
	}

	for i, ch := range m.channels {
		name := ch.DisplayName
		if len(name) > sidebarWidth-4 {
			name = name[:sidebarWidth-7] + "..."
		}

		var line string
		switch {
		case i == m.selected:
			badge := ""
			if ch.Unread > 0 {
				badge = fmt.Sprintf(" (%d)", ch.Unread)
			}
			line = channelSelectedStyle.Render("▸ " + name + badge)
		case ch.Unread > 0:
			badge := fmt.Sprintf(" (%d)", ch.Unread)
			line = channelUnreadStyle.Render("  " + name + badge)
		default:
			line = channelStyle.Render("  " + name)
		}

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Pad to fill height
	lines := len(m.channels)
	for i := lines; i < m.height-1; i++ {
		sb.WriteString("\n")
	}

	return sbStyle.Height(m.height - 1).Render(sb.String())
}

func (m Model) renderContent() string {
	// Header
	chName := ""
	if m.selected >= 0 && m.selected < len(m.channels) {
		chName = m.channels[m.selected].DisplayName
	}
	header := headerStyle.Render("#" + chName)

	// Input
	inStyle := inputStyle
	if m.pane == paneInput {
		inStyle = inputActiveStyle
	}
	contentWidth := m.width - sidebarWidth - 2
	input := inStyle.Width(contentWidth).Render(m.input.View())

	// Viewport fills remaining space
	vpHeight := m.height - 3 // header(1) + border(1) + input(1)
	if vpHeight < 1 {
		vpHeight = 1
	}
	m.viewport.Width = contentWidth
	m.viewport.Height = vpHeight

	return lipgloss.JoinVertical(lipgloss.Left, header, m.viewport.View(), input)
}

func (m Model) renderMessages() string {
	if len(m.messages) == 0 {
		return statusStyle.Render("No messages.")
	}

	var sb strings.Builder
	for _, msg := range m.messages {
		ts := msgTimeStyle.Render(msg.Time.Format("15:04"))
		user := msgUserStyle.Render(msg.User)
		sb.WriteString(fmt.Sprintf("%s %s  %s\n", ts, user, msg.Text))
	}
	return sb.String()
}
