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
	if m.showDetail {
		return m.renderDetail()
	}

	// Header
	chName := ""
	if ch := m.selectedChannel(); ch != nil {
		chName = ch.DisplayName
	}
	hStyle := headerStyle
	if m.pane == paneMessages {
		hStyle = headerActiveStyle
	}
	header := hStyle.Render("#" + chName)
	if m.notification != "" {
		header += "  " + notificationStyle.Render(m.notification)
	}

	// Input
	inStyle := inputStyle
	if m.pane == paneInput {
		inStyle = inputActiveStyle
	}
	contentWidth := m.width - sidebarWidth - 2
	input := inStyle.Width(contentWidth).Render(m.input.View())

	// Ensure viewport never exceeds available space
	vpH := m.height - 4
	if m.showHelp {
		vpH--
	}
	if vpH < 1 {
		vpH = 1
	}
	m.viewport.Width = contentWidth
	m.viewport.Height = vpH

	parts := []string{header, m.viewport.View(), input}
	if m.showHelp {
		parts = append(parts, m.renderHelp())
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) renderHelp() string {
	var help string
	switch m.pane {
	case paneSidebar:
		help = "tab: messages │ enter: select │ d: DMs │ C-spc: next unread │ j/k: navigate │ q: quit"
	case paneMessages:
		help = "tab: input │ j/k: navigate │ pgup/pgdn: scroll │ esc: sidebar │ q: quit"
	case paneInput:
		help = "tab: sidebar │ enter: send │ esc: sidebar"
	}
	return helpStyle.Render("F1: toggle help │ " + help)
}

func (m Model) renderDetail() string {
	contentWidth := m.width - sidebarWidth - 2
	boxW := contentWidth - 4
	if boxW < 10 {
		boxW = 10
	}
	box := detailStyle.Width(boxW).Render(m.detailVP.View())
	help := helpStyle.Render("esc/space: close │ j/k: scroll")
	return lipgloss.JoinVertical(lipgloss.Left, box, help)
}

// displayText truncates multi-line messages (>4 lines) to just the first line.
func displayText(text string) string {
	lines := strings.SplitN(text, "\n", 5)
	if len(lines) > 4 {
		return lines[0] + " [...]"
	}
	return text
}

func (m Model) renderMessages() string {
	if len(m.messages) == 0 {
		return statusStyle.Render("No messages.")
	}

	var sb strings.Builder
	for i, msg := range m.messages {
		ts := msgTimeStyle.Render(msg.Time.Format("15:04"))
		user := msgUserStyle.Render(msg.User)
		text := displayText(msg.Text)
		if m.pane == paneMessages && i == m.selectedMsg {
			cursor := msgSelectedStyle.Render("▸")
			fmt.Fprintf(&sb, "%s %s %s  %s\n", cursor, ts, user, text)
		} else {
			fmt.Fprintf(&sb, "  %s %s  %s\n", ts, user, text)
		}
	}
	return sb.String()
}
