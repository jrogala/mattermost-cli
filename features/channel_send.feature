Feature: Channel sending
  As a user I can send messages to a channel.

  Background:
    Given a running Mattermost instance
    And an authenticated user

  Scenario: Send a message to a channel
    Given a channel "SendCh" exists
    When I send "Hello from test" to channel "SendCh"
    Then the message should be posted successfully
    And the posted message should have an ID
