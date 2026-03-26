package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/pkg/ops"
)

func loadMe(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		me, err := ops.GetMe(c)
		if err != nil {
			return errMsg{err}
		}
		return meLoadedMsg{me}
	}
}

type meLoadedMsg struct{ me *ops.UserInfo }

func loadChannels(c *client.Client) tea.Cmd {
	return func() tea.Msg {
		channels, err := ops.ListChannels(c, ops.ListOptions{IncludeMuted: true})
		if err != nil {
			return errMsg{err}
		}

		// Merge unread counts
		unread, err := ops.GetUnread(c, ops.UnreadOptions{IncludeMuted: true})
		if err == nil {
			unreadMap := make(map[string]ops.Channel, len(unread))
			for _, u := range unread {
				unreadMap[u.ID] = u
			}
			for i := range channels {
				if u, ok := unreadMap[channels[i].ID]; ok {
					channels[i].Unread = u.Unread
					channels[i].Mentions = u.Mentions
				}
			}
		}

		return channelsLoadedMsg{channels}
	}
}

func loadMessages(c *client.Client, channelID string) tea.Cmd {
	return func() tea.Msg {
		msgs, err := ops.ReadMessages(c, channelID, ops.ReadOptions{Limit: 50})
		if err != nil {
			return errMsg{err}
		}
		return messagesLoadedMsg{channelID, msgs}
	}
}

func waitForWS(events <-chan ops.Message) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-events
		if !ok {
			return wsClosedMsg{}
		}
		return wsEventMsg{msg}
	}
}

func sendMessage(c *client.Client, channelID, text string) tea.Cmd {
	return func() tea.Msg {
		_, err := ops.SendMessage(c, channelID, text)
		if err != nil {
			return errMsg{err}
		}
		return messageSentMsg{}
	}
}
