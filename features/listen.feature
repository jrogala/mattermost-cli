Feature: WebSocket listen
  As a user I can stream real-time events from Mattermost via WebSocket.

  Background:
    Given a running Mattermost instance
    And an authenticated user

  Scenario: Receive a posted event
    Given a channel "ListenCh" exists
    And I am listening on all events
    When alice posts "ws hello" to channel "ListenCh"
    Then I should receive a "posted" event within 5 seconds
    And the event message should contain "ws hello"
    And the event sender should be "alice"

  Scenario: Filter by event type
    Given a channel "FilterEvtCh" exists
    And I am listening for "posted" events only
    When alice posts "filtered msg" to channel "FilterEvtCh"
    Then I should receive a "posted" event within 5 seconds
    And the event message should contain "filtered msg"

  Scenario: Filter by channel
    Given a channel "WantedCh" exists
    And a channel "UnwantedCh" exists
    And I am listening on channel "WantedCh" only
    When alice posts "wanted" to channel "WantedCh"
    And alice posts "unwanted" to channel "UnwantedCh"
    Then I should receive a "posted" event within 5 seconds
    And the event message should contain "wanted"
    And I should not have received an event for channel "UnwantedCh"

  Scenario: Event contains channel ID
    Given a channel "MetaCh" exists
    And I am listening on all events
    When alice posts "meta test" to channel "MetaCh"
    Then I should receive a "posted" event within 5 seconds
    And the event should have a non-empty channel ID

  Scenario: Receive post_deleted event
    Given a channel "DeleteCh" exists
    And I am listening for "post_deleted" events only
    When alice posts "doomed" to channel "DeleteCh"
    And that post is deleted
    Then I should receive a "post_deleted" event within 5 seconds
