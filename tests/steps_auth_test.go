package tests

import (
	"fmt"

	"github.com/jrogala/mattermost-cli/client"
	"github.com/jrogala/mattermost-cli/config"
	"github.com/jrogala/mattermost-cli/pkg/ops"
)

func (sc *scenarioCtx) iAuthenticateWithValidToken() error {
	cfg := &config.Config{
		URL:   sc.env.MattermostURL,
		Token: sc.env.AdminToken,
	}
	c := client.New(cfg)
	_, err := ops.GetMe(c)
	sc.lastErr = err
	if err == nil {
		sc.client = c
	}
	return nil
}

func (sc *scenarioCtx) iAuthenticateWithInvalidToken() error {
	cfg := &config.Config{
		URL:   sc.env.MattermostURL,
		Token: "invalid-token-that-does-not-exist",
	}
	c := client.New(cfg)
	_, err := ops.GetMe(c)
	sc.lastErr = err
	return nil
}

func (sc *scenarioCtx) authShouldSucceed() error {
	if sc.lastErr != nil {
		return fmt.Errorf("expected auth to succeed, got: %v", sc.lastErr)
	}
	return nil
}

func (sc *scenarioCtx) authShouldFail() error {
	if sc.lastErr == nil {
		return fmt.Errorf("expected auth to fail, but it succeeded")
	}
	return nil
}

func (sc *scenarioCtx) iShouldBeAbleToRetrieveMyUserInfo() error {
	if sc.client == nil {
		return fmt.Errorf("no authenticated client")
	}
	info, err := ops.GetMe(sc.client)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}
	if info.Username == "" {
		return fmt.Errorf("username is empty")
	}
	return nil
}
