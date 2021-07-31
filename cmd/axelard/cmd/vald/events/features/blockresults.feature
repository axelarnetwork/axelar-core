Feature: provide a robust source for block results

  Background:
    Given a block notifier
    And a client to get blockchain events
    And a cancellable context

  Scenario Outline: block results are provided
    Given a block result source
    And blocks <start> to <latest> are available
    When I try to receive blockchain results
    Then I receive all results from <start> to <latest>
    Examples:
      | start  | latest |
      | 1      | 1      |
      | 1      | 2      |
      | 100000 | 100018 |

  Scenario: block notifier fails
    Given a block result source
    When I try to receive blockchain results
    And the block notifier fails
    Then the block result source fails

  Scenario: client fails
    Given a block result source
    When I try to receive blockchain results
    And the client fails
    Then the block result source fails

  Scenario Outline: canceled context
    Given a block result source
    And blocks <start> to <latest> are available
    When I try to receive blockchain results
    And the context is canceled
    Then the result channel gets closed
    Examples:
      | start  | latest |
      | 1      | 1      |
      | 1      | 2      |
      | 100000 | 100018 |