package tests

import (
	"fmt"

	"github.com/jrogala/mattermost-cli/pkg/ops"
)

func (sc *scenarioCtx) iRequestMyUserInfo() error {
	info, err := ops.GetMe(sc.client)
	sc.lastErr = err
	sc.userInfo = info
	return nil
}

func (sc *scenarioCtx) iShouldGetAUsername() error {
	info, ok := sc.userInfo.(*ops.UserInfo)
	if !ok || info == nil {
		return fmt.Errorf("no user info available")
	}
	if info.Username == "" {
		return fmt.Errorf("username is empty")
	}
	return nil
}

func (sc *scenarioCtx) iShouldGetAnEmail() error {
	info, ok := sc.userInfo.(*ops.UserInfo)
	if !ok || info == nil {
		return fmt.Errorf("no user info available")
	}
	if info.Email == "" {
		return fmt.Errorf("email is empty")
	}
	return nil
}

func (sc *scenarioCtx) iShouldGetAUserID() error {
	info, ok := sc.userInfo.(*ops.UserInfo)
	if !ok || info == nil {
		return fmt.Errorf("no user info available")
	}
	if info.ID == "" {
		return fmt.Errorf("user ID is empty")
	}
	return nil
}
