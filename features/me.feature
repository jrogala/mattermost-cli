Feature: User identity
  As an authenticated user I can retrieve my own profile information.

  Background:
    Given a running Mattermost instance
    And an authenticated user

  Scenario: Get own user info
    When I request my user info
    Then I should get a username
    And I should get an email
    And I should get a user ID
