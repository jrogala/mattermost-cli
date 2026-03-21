Feature: Channel finding
  As a user I can find channels by name or find a user's DM channel.

  Background:
    Given a running Mattermost instance
    And an authenticated user

  Scenario: Find channel by name
    Given a channel "Find Target" exists
    When I search for channel "find target"
    Then I should find channel "Find Target"
    And the found channel should have an ID

  Scenario: Find with multiple matches lists all
    Given a channel "Multi Find Alpha" exists
    And a channel "Multi Find Beta" exists
    When I search for channel "multi find"
    Then I should find channel "Multi Find Alpha"
    And I should find channel "Multi Find Beta"

  Scenario: Find with no matches
    When I search for channel "nonexistentxyz"
    Then I should find 0 channels

  Scenario: Find user DM channel
    Given a user "testbot" exists
    When I search for DM channel with user "testbot"
    Then I should get a DM channel
    And the DM channel should have an ID

  Scenario: Find user DM with unknown user
    When I search for DM channel with user "unknownuserxyz"
    Then it should fail with "not found"
