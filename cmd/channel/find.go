package channel

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"os"

	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/internal/cmdutil"
	"github.com/spf13/cobra"
)

var findType string

func init() {
	findCmd.Flags().StringVar(&findType, "type", "", "filter by type: public, private, dm, group, user")
	Cmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:   "find <name>",
	Short: "Find a channel or user DM by name. Use --type user to find a user's DM channel.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := cmdutil.NewClient()
		term := args[0]

		// --type user: look up a user by username and return their DM channel
		if findType == "user" {
			u, err := c.GetUserByUsername(term)
			if err != nil {
				return fmt.Errorf("user %q not found: %w", term, err)
			}
			dm, err := c.GetOrCreateDM(u.ID)
			if err != nil {
				return err
			}
			if cmdutil.IsJSON(cmd) {
				return cmdutil.PrintJSON(map[string]string{
					"channel_id": dm.ID,
					"username":   u.Username,
					"type":       "dm",
				})
			}
			fmt.Printf("%s\t%s (DM)\n", dm.ID, u.Username)
			return nil
		}

		me, err := c.Me()
		if err != nil {
			return err
		}

		teamID, err := cmdutil.ResolveTeam(c, "")
		if err != nil {
			return err
		}

		channels, err := c.GetChannels(teamID)
		if err != nil {
			return err
		}

		if findType != "" {
			typeCode := channelTypeCode(findType)
			var filtered []client.Channel
			for _, ch := range channels {
				if ch.Type == typeCode {
					filtered = append(filtered, ch)
				}
			}
			channels = filtered
		}

		termLower := strings.ToLower(term)
		var matches []client.Channel
		for _, ch := range channels {
			name := cmdutil.ResolveChannelName(c, ch, me.ID)
			if strings.Contains(strings.ToLower(name), termLower) ||
				strings.Contains(strings.ToLower(ch.Name), termLower) {
				matches = append(matches, ch)
			}
		}

		if cmdutil.IsJSON(cmd) {
			type result struct {
				ID   string `json:"channel_id"`
				Type string `json:"type"`
				Name string `json:"name"`
			}
			var out []result
			for _, ch := range matches {
				out = append(out, result{
					ID:   ch.ID,
					Type: channelTypeName(ch.Type),
					Name: cmdutil.ResolveChannelName(c, ch, me.ID),
				})
			}
			return cmdutil.PrintJSON(out)
		}

		if len(matches) == 0 {
			fmt.Printf("No channels matching %q\n", term)
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "ID\tTYPE\tNAME")
		for _, ch := range matches {
			name := cmdutil.ResolveChannelName(c, ch, me.ID)
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", ch.ID, channelTypeName(ch.Type), name)
		}
		return w.Flush()
	},
}
