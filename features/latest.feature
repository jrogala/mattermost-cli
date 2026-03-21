Feature: Latest messages
  As a user I can see the latest messages across my non-muted channels.

  Background:
    Given a running Mattermost instance
    And an authenticated user

  Scenario: Get latest messages
    Given a channel "LatestA" exists with recent messages
    And a channel "LatestB" exists with recent messages
    When I get latest messages
    Then the latest results should include channel "LatestA"
    And the latest results should include channel "LatestB"

  Scenario: Latest excludes muted channels
    Given a channel "VisibleCh" exists with recent messages
    And a channel "MutedCh" exists with recent messages and is muted
    When I get latest messages
    Then the latest results should include channel "VisibleCh"
    And the latest results should not include channel "MutedCh"

  Scenario: Latest with channel limit
    Given 5 channels exist with recent messages
    When I get latest messages with channel limit 2
    Then the latest results should have at most 2 channels

  Scenario: Latest with per-channel limit
    Given a channel "ManyMsgs" exists with 10 recent messages
    When I get latest messages with per-channel limit 2
    Then channel "ManyMsgs" should have at most 2 messages in the results

  Scenario: No recent messages
    Given a channel "EmptyLatest" exists with no messages
    When I get latest messages
    Then the latest results should be empty
