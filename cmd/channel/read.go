package channel

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var (
	readLimit int
	readTrunc int
	readSince string
)

func init() {
	readCmd.Flags().IntVarP(&readLimit, "limit", "n", 50, "max number of messages to show")
	readCmd.Flags().IntVar(&readTrunc, "truncate", 120, "max chars per message (0 = no truncation)")
	readCmd.Flags().StringVar(&readSince, "since", "", "show messages since (e.g. 1h, 24h, 2d, 1w, 2026-03-20)")
	Cmd.AddCommand(readCmd)
}

var readCmd = &cobra.Command{
	Use:   "read <channel_id_or_name>",
	Short: "Read recent messages from a channel. Use --since to filter by time.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdutil.NewClient()
		channelID, err := resolveChannelArg(c, args[0])
		if err != nil {
			return err
		}

		params := map[string]string{
			"per_page": strconv.Itoa(readLimit),
		}

		if readSince != "" {
			sinceMs, err := parseSince(readSince)
			if err != nil {
				return fmt.Errorf("invalid --since value %q: %w", readSince, err)
			}
			params["since"] = strconv.FormatInt(sinceMs, 10)
		}

		pl, err := c.GetChannelPosts(channelID, params)
		if err != nil {
			return err
		}

		if cmdutil.IsJSON(cmd) {
			return cmdutil.PrintJSON(pl)
		}

		if len(pl.Order) == 0 {
			fmt.Println("No messages.")
			return nil
		}

		userCache := map[string]string{}
		resolveUser := func(userID string) string {
			if name, ok := userCache[userID]; ok {
				return name
			}
			u, err := c.GetUser(userID)
			if err != nil {
				userCache[userID] = userID[:8]
				return userCache[userID]
			}
			userCache[userID] = u.Username
			return u.Username
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for i := len(pl.Order) - 1; i >= 0; i-- {
			post := pl.Posts[pl.Order[i]]
			ts := time.UnixMilli(post.CreateAt).Format("Jan 02 15:04")
			user := resolveUser(post.UserID)
			msg := cmdutil.TruncateMsg(post.Message, readTrunc)
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", ts, user, msg)
		}
		return w.Flush()
	},
}

// parseSince converts a duration or date string to a Unix millisecond timestamp.
// Supports: "1h", "24h", "2d", "1w", "YYYY-MM-DD".
func parseSince(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToLower(s))

	// Try YYYY-MM-DD format
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UnixMilli(), nil
	}

	if len(s) < 2 {
		return 0, fmt.Errorf("use: 1h, 24h, 2d, 1w, or YYYY-MM-DD")
	}

	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("cannot parse number from %q", s)
	}

	var d time.Duration
	switch unit {
	case 'h':
		d = time.Duration(num) * time.Hour
	case 'd':
		d = time.Duration(num) * 24 * time.Hour
	case 'w':
		d = time.Duration(num) * 7 * 24 * time.Hour
	default:
		return 0, fmt.Errorf("unknown unit %q, use h/d/w", string(unit))
	}

	return time.Now().Add(-d).UnixMilli(), nil
}
