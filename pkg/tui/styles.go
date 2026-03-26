package tui

import "github.com/charmbracelet/lipgloss"

const sidebarWidth = 28

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
				BorderForeground(special)

	sectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(special).
				PaddingLeft(1)

	channelStyle = lipgloss.NewStyle().
			Foreground(subtle)

	channelSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(highlight)

	channelUnreadStyle = lipgloss.NewStyle().
				Bold(true)

	channelMutedStyle = lipgloss.NewStyle().
				Foreground(subtle).
				Faint(true)

	inputStyle = lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(subtle)

	inputActiveStyle = lipgloss.NewStyle().
				BorderTop(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(special)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(special).
			PaddingLeft(1)

	headerActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(special).
				PaddingLeft(1).
				Underline(true)

	msgTimeStyle = lipgloss.NewStyle().
			Foreground(subtle)

	msgUserStyle = lipgloss.NewStyle().
			Bold(true)

	msgSelectedStyle = lipgloss.NewStyle().
				Foreground(highlight).
				Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(subtle).
			PaddingLeft(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(subtle)

	notificationStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.AdaptiveColor{Light: "#E85D4A", Dark: "#FF6B6B"})

	detailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(special).
			Padding(0, 1)
)
