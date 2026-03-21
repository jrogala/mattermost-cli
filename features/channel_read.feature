Feature: Channel reading
  As a user I can read messages from a channel, filtered by time.

  Background:
    Given a running Mattermost instance
    And an authenticated user

  Scenario: Read messages from a channel
    Given a channel "ReadCh" exists with messages:
      | user  | message     |
      | alice | Hello world |
      | bob   | Hi there    |
    When I read messages from channel "ReadCh"
    Then I should receive 2 messages
    And the messages should contain "Hello world"
    And the messages should contain "Hi there"

  Scenario: Read messages with limit
    Given a channel "LimitCh" exists with 10 messages
    When I read messages from channel "LimitCh" with limit 3
    Then I should receive at most 3 messages

  Scenario: Read messages since a duration
    Given a channel "SinceCh" exists with messages:
      | user  | message     | age |
      | alice | Old message | 48h |
      | bob   | New message | 1m  |
    When I read messages from channel "SinceCh" since "1h"
    Then the messages should contain "New message"
    And the messages should not contain "Old message"

  Scenario: Read messages since a date
    Given a channel "DateCh" exists with messages:
      | user  | message     | age |
      | alice | Old message | 48h |
      | bob   | New message | 1m  |
    When I read messages from channel "DateCh" since "yesterday"
    Then the messages should contain "New message"
    And the messages should not contain "Old message"

  Scenario: Read empty channel
    Given a channel "EmptyCh" exists with no messages
    When I read messages from channel "EmptyCh"
    Then I should receive 0 messages
