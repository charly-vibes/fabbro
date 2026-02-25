Feature: Web Incremental Search
  As a reviewer using the web interface,
  I want to search for text in the document with /
  so that I can quickly find and navigate to specific content.

  # Implementation Status Legend:
  # @implemented - Working in current build
  # @planned     - Designed but not yet implemented

  # --- Opening / Closing Search ---

  @planned
  Scenario: Open search bar with / key
    Given I am in the editor view with the viewer focused
    When I press "/"
    Then a search bar should appear at the top of the viewer
    And focus should move to the search input

  @planned
  Scenario: Dismiss search with Escape
    Given the search bar is open
    When I press "Escape"
    Then the search bar should close
    And all search highlights should be removed
    And focus should return to the viewer

  @planned
  Scenario: Confirm search with Enter
    Given the search bar is open with matches
    When I press "Enter"
    Then the search bar should close
    But the search highlights should remain visible
    And n/N navigation should still work
    And focus should return to the viewer

  @planned
  Scenario: / key is ignored when typing in textarea or input
    Given I am editing an annotation in a textarea
    When I press "/"
    Then the search bar should not open
    And the "/" character should be typed normally

  # --- Incremental Highlighting ---

  @planned
  Scenario: Matches highlight as the user types
    Given the search bar is open
    When I type "func"
    Then all occurrences of "func" in the document should be highlighted
    And the highlights should use a distinct search-match style

  @planned
  Scenario: Highlights update incrementally
    Given the search bar is open and I have typed "fun"
    When I type "c" to make "func"
    Then the highlights should update to show only "func" matches

  @planned
  Scenario: No matches found
    Given the search bar is open
    When I type "xyznonexistent"
    Then no highlights should appear
    And the match counter should display "0/0"

  @planned
  Scenario: Search is case-insensitive
    Given the document contains "Hello" and "hello"
    When I search for "hello"
    Then both "Hello" and "hello" should be highlighted

  # --- Match Counter ---

  @planned
  Scenario: Match counter shows current position
    Given the search bar is open and there are 12 matches for "the"
    When the first match is active
    Then the match counter should display "1/12"

  @planned
  Scenario: Counter updates on navigation
    Given the match counter shows "1/12"
    When I press "n" or the down-arrow in the search bar
    Then the counter should update to "2/12"

  # --- Match Navigation ---

  @planned
  Scenario: Navigate to next match with n
    Given the search bar is closed but matches were found
    When I press "n" in the viewer
    Then the viewer should scroll to the next match
    And the current match should be visually distinct

  @planned
  Scenario: Navigate to previous match with N
    Given the current match is 5/12
    When I press "N" (Shift+n) in the viewer
    Then the current match should become 4/12
    And the viewer should scroll to it

  @planned
  Scenario: Navigation wraps around at end
    Given the current match is 12/12
    When I press "n"
    Then the current match should wrap to 1/12

  @planned
  Scenario: Navigation wraps around at beginning
    Given the current match is 1/12
    When I press "N"
    Then the current match should wrap to 12/12

  @planned
  Scenario: First match is scrolled to when search begins
    Given the search bar is open
    When I type a query that has matches
    Then the viewer should scroll to the first match
    And the first match should be marked as current
