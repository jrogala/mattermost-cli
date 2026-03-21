// Package channel implements the channel subcommands.
package channel

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent channel command.
var Cmd = &cobra.Command{
	Use:     "channel",
	Aliases: []string{"ch"},
	Short:   "Manage channels: find, read, send",
}
