Feature: Mentions
  As a user I can see messages that mention me, including from muted channels.

  Background:
    Given a running Mattermost instance
    And an authenticated user

  Scenario: Get my mentions
    Given a channel "MentionCh" exists with a message mentioning me from "alice"
    When I get my mentions
    Then the mentions should include a message from "alice"

  Scenario: Mentions includes muted channels
    Given a channel "MutedMention" exists and is muted with a message mentioning me from "bob"
    When I get my mentions
    Then the mentions should include a message from "bob"

  Scenario: No mentions
    Given a channel "NoMentionCh" exists with no messages
    When I get my mentions
    Then the mentions list should be empty
