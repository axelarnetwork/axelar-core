Feature: provide a robust source for existing blocks

  Background:
    Given a client to get block heights
    And a cancellable context

  Scenario Outline: block events don't work
    Given a block notifier starting at block <start>
    And block <latest> is available
    And the client only provides blocks through a query
    When I try to receive block heights
    Then I receive all blocks from <start> to <latest>

    Examples:
      | start  | latest |
      | 0      | 0      |
      | 0      | 1      |
      | 100000 | 100018 |

  Scenario Outline: stale block queries
    Given a block notifier starting at block <start>
    And block <latest> is available
    And the client only provides blocks through events
    When I try to receive block heights
    Then I receive all blocks from <start> to <latest>

    Examples:
      | start  | latest |
      | 0      | 0      |
      | 0      | 1      |
      | 100000 | 100018 |

  Scenario Outline: canceled context
    Given a block notifier starting at block <start>
    And block <latest> is available
    When I try to receive block heights
    And the context is canceled
    Then the block channel gets closed

    Examples:
      | start | latest |
      | 0     | 0      |
      | 0     | 1      |
      | 0     | 10000  |

  Scenario Outline: subscription fails
    Given a block notifier starting at block <start>
    And block <latest> is available
    And the client subscription fails
    When I try to receive block heights
    Then I receive all blocks from <start> to <latest>

    Examples:
      | start  | latest |
      | 0      | 0      |
      | 0      | 1      |
      | 100000 | 100018 |

  Scenario Outline: block query fails
    Given a block notifier starting at block <start>
    And block <latest> is available
    And the client's query fails
    When I try to receive block heights
    Then the block notifier fails

    Examples:
      | start  | latest |
      | 0      | 0      |
      | 0      | 1      |
      | 100000 | 100018 |

  Scenario Outline: negative start block
    Given a block notifier starting at block <start>
    And block <latest> is available
    When I try to receive block heights
    Then I receive all blocks from 0 to <latest>

    Examples:
      | start  | latest |
      | -1     | 0      |
      | -10000 | 1      |
      | -10000 | 10     |
