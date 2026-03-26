package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jrogala/mattermost-cli/pkg/tui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tuiCmd)
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive chat TUI.",
	RunE: func(_ *cobra.Command, _ []string) error {
		c := newClient()
		m := tui.New(c)
		p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
		_, err := p.Run()
		return err
	},
}
