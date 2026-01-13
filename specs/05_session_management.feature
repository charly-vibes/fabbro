Feature: Session Management
  As a user, I want to manage my review sessions
  so that I can list, resume, and organize my feedback work.

  # Implementation Status Legend:
  # @implemented - Working in current build
  # @planned     - Designed but not yet implemented
  #
  # Note: All scenarios in this spec are currently @planned.
  # Session management commands are a priority for post-tracer development.

  Background:
    Given I am in a directory that has been initialized with `fabbro init`

  # --- Listing Sessions ---

  @planned
  Scenario: Listing all sessions
    Given the following sessions exist:
      | session_id  | created_at          | source      | annotations |
      | review-001  | 2026-01-10 10:00:00 | stdin       | 5           |
      | review-002  | 2026-01-11 14:30:00 | document.md | 12          |
      | review-003  | 2026-01-11 16:00:00 | stdin       | 0           |
    When I run the command `fabbro sessions`
    Then the output should list all 3 sessions
    And each session should show ID, creation time, source, and annotation count
    And sessions should be sorted by creation time (newest first)

  @planned
  Scenario: Listing sessions in JSON format
    Given sessions exist
    When I run the command `fabbro sessions --json`
    Then the output should be valid JSON
    And the JSON should contain an array of session objects

  @planned
  Scenario: No sessions exist
    Given no sessions exist
    When I run the command `fabbro sessions`
    Then the output should indicate no sessions found
    And the output should suggest how to create a session

  # --- Showing Session Details ---

  @planned
  Scenario: Showing session details
    Given a session "review-123" exists with 5 annotations
    When I run the command `fabbro show review-123`
    Then the output should display:
      | field            | value                |
      | Session ID       | review-123           |
      | Created          | <timestamp>          |
      | Source           | stdin                |
      | Annotations      | 5                    |
      | Content lines    | 100                  |
    And the output should list annotation summary by type

  @planned
  Scenario: Showing session with annotation breakdown
    Given a session exists with:
      | type     | count |
      | comment  | 3     |
      | delete   | 1     |
      | question | 2     |
    When I run the command `fabbro show <session-id>`
    Then the annotation breakdown should show:
      """
      Annotations (6 total):
        comment:  3
        question: 2
        delete:   1
      """

  @planned
  Scenario: Showing non-existent session
    Given no session "missing" exists
    When I run the command `fabbro show missing`
    Then an error message should indicate the session was not found
    And the command should exit with code 1

  # --- Resuming Sessions ---

  @planned
  Scenario: Resuming an interrupted review
    Given a session "review-123" exists with annotations
    When I run the command `fabbro resume review-123`
    Then the TUI should open with the session content
    And existing annotations should be visible
    And I should be able to add more annotations

  @planned
  Scenario: Resuming in editor mode
    Given a session "review-123" exists
    When I run the command `fabbro resume review-123 --editor`
    Then the $EDITOR should open with the session file
    And the TUI should NOT be launched

  @planned
  Scenario: Resuming non-existent session
    Given no session "missing" exists
    When I run the command `fabbro resume missing`
    Then an error message should indicate the session was not found
    And the command should exit with code 1

  # --- Deleting Sessions ---

  @planned
  Scenario: Deleting a session
    Given a session "review-123" exists
    When I run the command `fabbro delete review-123`
    Then a confirmation prompt should appear
    When I confirm deletion
    Then the session file should be removed
    And a success message should be displayed

  @planned
  Scenario: Deleting a session with --force
    Given a session "review-123" exists
    When I run the command `fabbro delete review-123 --force`
    Then the session file should be removed without confirmation
    And a success message should be displayed

  @planned
  Scenario: Deleting non-existent session
    Given no session "missing" exists
    When I run the command `fabbro delete missing`
    Then an error message should indicate the session was not found
    And the command should exit with code 1

  # --- Cleaning Old Sessions ---

  @planned
  Scenario: Cleaning sessions older than threshold
    Given sessions exist with various ages:
      | session_id  | age     |
      | old-001     | 30 days |
      | old-002     | 14 days |
      | recent-001  | 2 days  |
    When I run the command `fabbro clean --older-than 7d`
    Then sessions older than 7 days should be listed for deletion
    And a confirmation prompt should appear
    When I confirm
    Then old-001 and old-002 should be deleted
    And recent-001 should remain

  @planned
  Scenario: Dry-run cleaning
    Given old sessions exist
    When I run the command `fabbro clean --older-than 7d --dry-run`
    Then sessions that would be deleted should be listed
    And no sessions should actually be deleted

  # --- Exporting Sessions ---

  @planned
  Scenario: Exporting session as standalone file
    Given a session "review-123" exists
    When I run the command `fabbro export review-123 --output review.fem`
    Then a file "review.fem" should be created
    And the file should contain the complete session with annotations

  @planned
  Scenario: Exporting session to stdout
    Given a session "review-123" exists
    When I run the command `fabbro export review-123`
    Then the session content should be printed to stdout

  # --- Session ID Autocompletion ---

  @planned
  Scenario: Partial session ID matching
    Given a session "review-abc123" exists
    When I run the command `fabbro show abc1`
    Then the command should match the full session ID
    And the session details should be displayed

  @planned
  Scenario: Ambiguous partial session ID
    Given sessions exist:
      | session_id     |
      | review-abc123  |
      | review-abc456  |
    When I run the command `fabbro show abc`
    Then an error should indicate multiple matching sessions
    And the matching sessions should be listed
