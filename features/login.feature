Feature: Authentication
  As a user I can authenticate with Mattermost and verify my credentials.

  Background:
    Given a running Mattermost instance

  Scenario: Valid token authenticates successfully
    When I authenticate with a valid token
    Then authentication should succeed
    And I should be able to retrieve my user info

  Scenario: Invalid token fails authentication
    When I authenticate with an invalid token
    Then authentication should fail

  Scenario: Token can access the API
    Given an authenticated user
    When I request my user info
    Then I should get a username
