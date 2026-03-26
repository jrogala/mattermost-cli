package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/jrogala/mattermost-cli/pkg/ops"
	"github.com/spf13/cobra"
)

var (
	listenEvents  string
	listenChannel string
)

func init() {
	rootCmd.AddCommand(listenCmd)
	listenCmd.Flags().StringVar(&listenEvents, "events", "", "comma-separated event types (e.g. posted,typing)")
	listenCmd.Flags().StringVar(&listenChannel, "channel", "", "channel ID or name to filter")
}

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Stream real-time events from Mattermost WebSocket.",
	RunE: func(cmd *cobra.Command, _ []string) error {
		c := newClient()

		var eventTypes []string
		if listenEvents != "" {
			eventTypes = strings.Split(listenEvents, ",")
		}

		var channelIDs []string
		if listenChannel != "" {
			chID, err := ops.ResolveChannelArg(c, listenChannel)
			if err != nil {
				return err
			}
			channelIDs = []string{chID}
		}

		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		events, errs, err := ops.Listen(ctx, c, ops.ListenOptions{
			EventTypes: eventTypes,
			ChannelIDs: channelIDs,
		})
		if err != nil {
			return err
		}

		jsonMode := cmdutil.IsJSON(cmd)

		for {
			select {
			case evt, ok := <-events:
				if !ok {
					return nil
				}
				if jsonMode {
					_ = cmdutil.PrintJSON(evt)
				} else {
					printEvent(evt)
				}
			case err, ok := <-errs:
				if !ok {
					return nil
				}
				fmt.Fprintf(os.Stderr, "websocket error: %v\n", err)
				return err
			case <-ctx.Done():
				return nil
			}
		}
	},
}

func printEvent(evt ops.Message) {
	ts := evt.Time.Format("15:04:05")
	ch := evt.Channel
	if ch == "" && len(evt.ChannelID) >= 8 {
		ch = evt.ChannelID[:8]
	}

	switch evt.Event {
	case "posted", "post_edited":
		fmt.Printf("[%s] %s | #%s | %s: %s\n", ts, evt.Event, ch, evt.User, truncateMsg(evt.Text, 120))
	case "typing":
		fmt.Printf("[%s] typing | #%s | %s\n", ts, ch, evt.User)
	default:
		fmt.Printf("[%s] %s | #%s\n", ts, evt.Event, ch)
	}
}
