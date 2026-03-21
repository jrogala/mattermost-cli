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

func channelTypeName(t string) string {
	switch t {
	case "O":
		return "public"
	case "P":
		return "private"
	case "D":
		return "dm"
	case "G":
		return "group"
	default:
		return t
	}
}

func channelTypeCode(name string) string {
	switch name {
	case "public":
		return "O"
	case "private":
		return "P"
	case "dm":
		return "D"
	case "group":
		return "G"
	default:
		return name
	}
}
