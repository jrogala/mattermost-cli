// Package cmd implements the mattermost-cli commands.
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/jrogala/mattermost-cli/cmd/channel"
	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:   "mattermost-cli",
	Short: "CLI client for Mattermost",
}

func init() {
	rootCmd.PersistentFlags().Bool("json", false, "output raw JSON")
	rootCmd.SetHelpFunc(customHelp)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.AddCommand(channel.Cmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Re-export helpers for top-level commands (unread, latest, mentions, me, login).
var newClient = cmdutil.NewClient
var truncateMsg = cmdutil.TruncateMsg

func customHelp(cmd *cobra.Command, _ []string) {
	if cmd == rootCmd {
		printTree()
		return
	}
	if !cmd.HasSubCommands() {
		printLeafHelp(cmd)
		return
	}
	printSubtree(cmd)
}

func printTree() {
	fmt.Println("mattermost-cli - Mattermost CLI client")
	fmt.Println("")
	fmt.Println("Global: --json (raw JSON output)")
	fmt.Println("")
	fmt.Println("Commands:")

	for _, cmd := range rootCmd.Commands() {
		if cmd.Hidden || cmd.Name() == "help" || cmd.Name() == "completion" {
			continue
		}
		if cmd.HasSubCommands() {
			fmt.Printf("  %s\n", cmd.Name())
			for _, sub := range cmd.Commands() {
				if sub.Hidden {
					continue
				}
				aliases := ""
				if len(sub.Aliases) > 0 {
					aliases = " (" + strings.Join(sub.Aliases, ", ") + ")"
				}
				fmt.Printf("    %-10s %s%s\n", sub.Name(), sub.Short, aliases)
			}
		} else {
			fmt.Printf("  %-12s %s\n", cmd.Name(), cmd.Short)
		}
	}

	fmt.Println("")
	fmt.Println("Run 'mattermost-cli <command> <subcommand> --help' for full details.")
}

func printSubtree(cmd *cobra.Command) {
	fmt.Printf("%s\n\n", cmd.Short)

	for _, sub := range cmd.Commands() {
		if sub.Hidden {
			continue
		}
		aliases := ""
		if len(sub.Aliases) > 0 {
			aliases = " (" + strings.Join(sub.Aliases, ", ") + ")"
		}
		fmt.Printf("  %-10s %s%s\n", sub.Name(), sub.Short, aliases)
	}

	fmt.Println("")
	fmt.Printf("Run 'mattermost-cli %s <subcommand> --help' for full details.\n", cmd.Name())
}

func printLeafHelp(cmd *cobra.Command) {
	fmt.Printf("%s %s\n", cmd.UseLine(), "")
	fmt.Println(cmd.Short)

	if cmd.HasLocalFlags() {
		fmt.Println("")
		fmt.Println("Flags:")
		cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
			shorthand := ""
			if f.Shorthand != "" {
				shorthand = "-" + f.Shorthand + ", "
			}
			def := ""
			if f.DefValue != "" && f.DefValue != "false" && f.DefValue != "0" {
				def = " (default: " + f.DefValue + ")"
			}
			fmt.Printf("  %s--%s %s%s\n", shorthand, f.Name, f.Usage, def)
		})
	}
}
