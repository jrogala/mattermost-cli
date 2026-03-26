package tui

import "github.com/charmbracelet/lipgloss"

const sidebarWidth = 22

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#999999", Dark: "#666666"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	sidebarStyle = lipgloss.NewStyle().
			Width(sidebarWidth).
			BorderRight(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(subtle)

	sidebarActiveStyle = lipgloss.NewStyle().
				Width(sidebarWidth).
				BorderRight(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(highlight)

	channelStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(subtle)

	channelSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Bold(true).
				Foreground(highlight)

	channelUnreadStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Bold(true)

	channelMutedStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Foreground(subtle).
				Faint(true)

	inputStyle = lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(subtle)

	inputActiveStyle = lipgloss.NewStyle().
				BorderTop(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(highlight)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(special).
			PaddingLeft(1)

	msgTimeStyle = lipgloss.NewStyle().
			Foreground(subtle)

	msgUserStyle = lipgloss.NewStyle().
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(subtle).
			PaddingLeft(1)
)
