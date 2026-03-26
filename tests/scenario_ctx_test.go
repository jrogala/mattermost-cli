package tests

import (
	"context"
	"fmt"
	"os"

	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/config"
	"github.com/jrogala/mattermost-cli/pkg/ops"
)

// scenarioCtx holds per-scenario state.
type scenarioCtx struct {
	env *TestEnvironment

	// client for the "authenticated user" (admin)
	client *client.Client

	// per-scenario data
	channels map[string]string // display_name -> channel_id
	lastErr  error

	// results from the latest "When" step
	userInfo    any
	channelList any
	findResults any
	dmResult    any
	sendResult  any
	messages    any
	unreadList  any
	latestList  any
	mentionList any

	// WebSocket listen state
	listenEvents   <-chan ops.Message
	listenErrors   <-chan error
	listenCancel   context.CancelFunc
	receivedEvents []ops.Message
	lastPostID     string
}

func newScenarioCtx(env *TestEnvironment) *scenarioCtx {
	return &scenarioCtx{
		env:      env,
		channels: make(map[string]string),
	}
}

// newClient creates a client pointing at the test Mattermost instance.
func (sc *scenarioCtx) newClient() *client.Client {
	cfg := &config.Config{
		URL:   sc.env.MattermostURL,
		Token: sc.env.AdminToken,
	}
	return client.New(cfg)
}

// newClientForUser creates a client for a specific test user.
func (sc *scenarioCtx) newClientForUser(username string) (*client.Client, error) {
	u, ok := sc.env.Users[username]
	if !ok {
		return nil, fmt.Errorf("unknown test user %q", username)
	}
	cfg := &config.Config{
		URL:   sc.env.MattermostURL,
		Token: u.Token,
	}
	return client.New(cfg), nil
}

// ensureEnvVars sets MATTERMOST_URL and MATTERMOST_TOKEN for any code
// that reads config from env (e.g. cmdutil.NewClient).
func (sc *scenarioCtx) ensureEnvVars() {
	os.Setenv("MATTERMOST_URL", sc.env.MattermostURL)
	os.Setenv("MATTERMOST_TOKEN", sc.env.AdminToken)
}
