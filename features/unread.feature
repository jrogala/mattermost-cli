Feature: Unread channels
  As a user I can see which channels have unread messages.

  Background:
    Given a running Mattermost instance
    And an authenticated user

  Scenario: Get unread channels
    Given a channel "Active" exists with 5 unread messages
    And a channel "Quiet" exists with 0 unread messages
    When I get unread channels
    Then the unread list should include channel "Active"
    And the unread list should not include channel "Quiet"

  Scenario: Unread excludes muted channels
    Given a channel "Visible" exists with 5 unread messages
    And a channel "MutedCh" exists with 10 unread messages and is muted
    When I get unread channels
    Then the unread list should include channel "Visible"
    And the unread list should not include channel "MutedCh"

  Scenario: Unread with muted included
    Given a channel "Visible" exists with 5 unread messages
    And a channel "MutedCh" exists with 10 unread messages and is muted
    When I get unread channels including muted
    Then the unread list should include channel "Visible"
    And the unread list should include channel "MutedCh"

  Scenario: Unread sorts mentions first
    Given a channel "NoMention" exists with 5 unread messages and 0 mentions
    And a channel "WithMention" exists with 2 unread messages and 1 mention
    When I get unread channels
    Then "WithMention" should appear before "NoMention" in the unread list

  Scenario: No unread messages
    Given no channels have unread messages
    When I get unread channels
    Then the unread list should be empty
