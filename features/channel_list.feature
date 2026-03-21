Feature: Channel listing
  As a user I can list channels I belong to, excluding muted by default.

  Background:
    Given a running Mattermost instance
    And an authenticated user

  Scenario: List channels excludes muted by default
    Given a channel "Visible" exists
    And a channel "Noisy" exists and is muted
    When I list my channels
    Then the result should include channel "Visible"
    And the result should not include channel "Noisy"

  Scenario: List channels with muted included
    Given a channel "Visible" exists
    And a channel "Noisy" exists and is muted
    When I list my channels including muted
    Then the result should include channel "Visible"
    And the result should include channel "Noisy"

  Scenario: List channels filtered by type
    Given a public channel "PubChan" exists
    And a private channel "PrivChan" exists
    When I list my channels filtered by type "public"
    Then the result should include channel "PubChan"
    And the result should not include channel "PrivChan"

  Scenario: List shows DM channels with resolved usernames
    Given I have a DM channel with user "alice"
    When I list my channels filtered by type "dm"
    Then the result should include a DM with "alice"
