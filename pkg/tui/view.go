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
	sbStyle := sidebarStyle
	if m.pane == paneSidebar {
		sbStyle = sidebarActiveStyle
	}
	return sbStyle.Height(m.height).Render(m.list.View())
}

func (m Model) renderContent() string {
	// Header
	chName := ""
	if ch := m.selectedChannel(); ch != nil {
		chName = ch.DisplayName
	}
	header := headerStyle.Render("#" + chName)

	// Input
	inStyle := inputStyle
	if m.pane == paneInput {
		inStyle = inputActiveStyle
	}
	contentWidth := m.width - sidebarWidth - 2
	input := inStyle.Width(contentWidth).Render(m.input.View())

	// Ensure viewport never exceeds available space
	vpH := m.height - 4
	if vpH < 1 {
		vpH = 1
	}
	m.viewport.Width = contentWidth
	m.viewport.Height = vpH

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
		fmt.Fprintf(&sb, "%s %s  %s\n", ts, user, msg.Text)
	}
	return sb.String()
}
