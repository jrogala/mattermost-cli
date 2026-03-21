package tests

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

func initializeScenario(ctx *godog.ScenarioContext) {
	sc := newScenarioCtx(globalEnv)

	ctx.Before(func(ctx context.Context, sc2 *godog.Scenario) (context.Context, error) {
		sc.lastErr = nil
		sc.channels = make(map[string]string)
		sc.userInfo = nil
		sc.channelList = nil
		sc.findResults = nil
		sc.dmResult = nil
		sc.sendResult = nil
		sc.messages = nil
		sc.unreadList = nil
		sc.latestList = nil
		sc.mentionList = nil
		sc.client = sc.newClient()
		sc.ensureEnvVars()
		return ctx, nil
	})

	// --- Background ---
	ctx.Step(`^a running Mattermost instance$`, sc.aRunningMattermostInstance)
	ctx.Step(`^an authenticated user$`, sc.anAuthenticatedUser)

	// --- Me ---
	ctx.Step(`^I request my user info$`, sc.iRequestMyUserInfo)
	ctx.Step(`^I should get a username$`, sc.iShouldGetAUsername)
	ctx.Step(`^I should get an email$`, sc.iShouldGetAnEmail)
	ctx.Step(`^I should get a user ID$`, sc.iShouldGetAUserID)

	// --- Auth ---
	ctx.Step(`^I authenticate with a valid token$`, sc.iAuthenticateWithValidToken)
	ctx.Step(`^I authenticate with an invalid token$`, sc.iAuthenticateWithInvalidToken)
	ctx.Step(`^authentication should succeed$`, sc.authShouldSucceed)
	ctx.Step(`^authentication should fail$`, sc.authShouldFail)
	ctx.Step(`^I should be able to retrieve my user info$`, sc.iShouldBeAbleToRetrieveMyUserInfo)

	// --- Channel list ---
	ctx.Step(`^I belong to channel "([^"]*)"$`, sc.iBelongToChannel)
	ctx.Step(`^I belong to channel "([^"]*)" which is muted$`, sc.iBelongToChannelMuted)
	ctx.Step(`^I belong to a public channel "([^"]*)"$`, sc.iBelongToPublicChannel)
	ctx.Step(`^I belong to a private channel "([^"]*)"$`, sc.iBelongToPrivateChannel)
	ctx.Step(`^I have a DM channel with user "([^"]*)"$`, sc.iHaveADMChannelWithUser)
	ctx.Step(`^I list my channels$`, sc.iListMyChannels)
	ctx.Step(`^I list my channels including muted$`, sc.iListMyChannelsIncludingMuted)
	ctx.Step(`^I list my channels filtered by type "([^"]*)"$`, sc.iListMyChannelsFilteredByType)
	ctx.Step(`^the result should include channel "([^"]*)"$`, sc.resultShouldIncludeChannel)
	ctx.Step(`^the result should not include channel "([^"]*)"$`, sc.resultShouldNotIncludeChannel)
	ctx.Step(`^the result should include a DM with "([^"]*)"$`, sc.resultShouldIncludeDMWith)

	// --- Channel find ---
	ctx.Step(`^a channel "([^"]*)" exists$`, sc.aChannelExists)
	ctx.Step(`^a user "([^"]*)" exists$`, sc.aUserExists)
	ctx.Step(`^I search for channel "([^"]*)"$`, sc.iSearchForChannel)
	ctx.Step(`^I search for DM channel with user "([^"]*)"$`, sc.iSearchForDMChannelWithUser)
	ctx.Step(`^I should find (\d+) channels$`, sc.iShouldFindExactlyNChannels)
	ctx.Step(`^I should find channel "([^"]*)"$`, sc.iShouldFindChannel)
	ctx.Step(`^the found channel should have an ID$`, sc.foundChannelShouldHaveAnID)
	ctx.Step(`^I should get a DM channel$`, sc.iShouldGetADMChannel)
	ctx.Step(`^the DM channel should have an ID$`, sc.dmChannelShouldHaveAnID)
	ctx.Step(`^it should fail with "([^"]*)"$`, sc.itShouldFailWith)

	// --- Channel read ---
	ctx.Step(`^a channel "([^"]*)" with messages:$`, sc.aChannelWithMessages)
	ctx.Step(`^a channel "([^"]*)" with (\d+) messages$`, sc.aChannelWithNMessages)
	ctx.Step(`^a channel "([^"]*)" with no messages$`, sc.aChannelWithNoMessages)
	ctx.Step(`^I read messages from channel "([^"]*)"$`, sc.iReadMessagesFromChannel)
	ctx.Step(`^I read messages from channel "([^"]*)" with limit (\d+)$`, sc.iReadMessagesFromChannelWithLimit)
	ctx.Step(`^I read messages from channel "([^"]*)" since "([^"]*)"$`, sc.iReadMessagesFromChannelSince)
	ctx.Step(`^I read messages from channel named "([^"]*)"$`, sc.iReadMessagesFromChannelNamed)
	ctx.Step(`^I should receive (\d+) messages$`, sc.iShouldReceiveNMessages)
	ctx.Step(`^I should receive at most (\d+) messages$`, sc.iShouldReceiveAtMostNMessages)
	ctx.Step(`^the messages should contain "([^"]*)"$`, sc.messagesShouldContain)
	ctx.Step(`^the messages should not contain "([^"]*)"$`, sc.messagesShouldNotContain)

	// --- Channel send ---
	ctx.Step(`^I send "([^"]*)" to channel "([^"]*)"$`, sc.iSendToChannel)
	ctx.Step(`^I send "([^"]*)" to channel named "([^"]*)"$`, sc.iSendToChannelNamed)
	ctx.Step(`^the message should be posted successfully$`, sc.messageShouldBePosted)
	ctx.Step(`^the posted message should have an ID$`, sc.postedMessageShouldHaveAnID)

	// --- Unread ---
	ctx.Step(`^channel "([^"]*)" has (\d+) unread messages$`, sc.channelHasUnreadMessages)
	ctx.Step(`^channel "([^"]*)" has (\d+) unread messages and is muted$`, sc.channelHasUnreadMessagesAndIsMuted)
	ctx.Step(`^channel "([^"]*)" has (\d+) unread messages and (\d+) mentions?$`, sc.channelHasUnreadMessagesAndMentions)
	ctx.Step(`^no channels have unread messages$`, sc.noChannelsHaveUnreadMessages)
	ctx.Step(`^I get unread channels$`, sc.iGetUnreadChannels)
	ctx.Step(`^I get unread channels including muted$`, sc.iGetUnreadChannelsIncludingMuted)
	ctx.Step(`^the unread list should include "([^"]*)"$`, sc.unreadListShouldInclude)
	ctx.Step(`^the unread list should not include "([^"]*)"$`, sc.unreadListShouldNotInclude)
	ctx.Step(`^"([^"]*)" should appear before "([^"]*)" in the unread list$`, sc.shouldAppearBeforeInUnreadList)
	ctx.Step(`^the unread list should be empty$`, sc.unreadListShouldBeEmpty)

	// --- Latest ---
	ctx.Step(`^channel "([^"]*)" has recent messages$`, sc.channelHasRecentMessages)
	ctx.Step(`^channel "([^"]*)" is muted and has recent messages$`, sc.channelIsMutedAndHasRecentMessages)
	ctx.Step(`^(\d+) channels have recent messages$`, sc.nChannelsHaveRecentMessages)
	ctx.Step(`^channel "([^"]*)" has (\d+) recent messages$`, sc.channelHasNRecentMessages)
	ctx.Step(`^no channels have recent messages$`, sc.noChannelsHaveRecentMessages)
	ctx.Step(`^I get latest messages$`, sc.iGetLatestMessages)
	ctx.Step(`^I get latest messages with channel limit (\d+)$`, sc.iGetLatestMessagesWithChannelLimit)
	ctx.Step(`^I get latest messages with per-channel limit (\d+)$`, sc.iGetLatestMessagesWithPerChannelLimit)
	ctx.Step(`^the latest results should include channel "([^"]*)"$`, sc.latestResultsShouldIncludeChannel)
	ctx.Step(`^the latest results should not include channel "([^"]*)"$`, sc.latestResultsShouldNotIncludeChannel)
	ctx.Step(`^the latest results should have at most (\d+) channels$`, sc.latestResultsShouldHaveAtMostNChannels)
	ctx.Step(`^the "([^"]*)" section should have at most (\d+) messages$`, sc.sectionShouldHaveAtMostNMessages)
	ctx.Step(`^the latest results should be empty$`, sc.latestResultsShouldBeEmpty)

	// --- Mentions ---
	ctx.Step(`^channel "([^"]*)" has a message mentioning me from "([^"]*)"$`, sc.channelHasMentionFrom)
	ctx.Step(`^channel "([^"]*)" is muted and has a message mentioning me from "([^"]*)"$`, sc.channelIsMutedAndHasMentionFrom)
	ctx.Step(`^no channels have mentions$`, sc.noChannelsHaveMentions)
	ctx.Step(`^I get my mentions$`, sc.iGetMyMentions)
	ctx.Step(`^the mentions should include a message from "([^"]*)"$`, sc.mentionsShouldIncludeMessageFrom)
	ctx.Step(`^the mentions should reference channel "([^"]*)"$`, sc.mentionsShouldReferenceChannel)
	ctx.Step(`^the mentions list should be empty$`, sc.mentionsListShouldBeEmpty)
}

// --- Background steps ---

func (sc *scenarioCtx) aRunningMattermostInstance() error {
	if sc.env == nil || sc.env.MattermostURL == "" {
		return fmt.Errorf("Mattermost is not running")
	}
	return nil
}

func (sc *scenarioCtx) anAuthenticatedUser() error {
	if sc.client == nil {
		return fmt.Errorf("no authenticated client")
	}
	return nil
}

// --- Common assertion ---

func (sc *scenarioCtx) itShouldFailWith(expected string) error {
	if sc.lastErr == nil {
		return fmt.Errorf("expected an error containing %q but got none", expected)
	}
	if !strings.Contains(sc.lastErr.Error(), expected) {
		return fmt.Errorf("expected error containing %q, got: %s", expected, sc.lastErr.Error())
	}
	return nil
}
