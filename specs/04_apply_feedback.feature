Feature: Apply Feedback
  As a user (or an LLM agent),
  I want to extract annotations from a review session as structured data
  so that the feedback can be processed and acted upon.

  Background:
    Given I am in a directory that has been initialized with `fabbro init`

  # --- Basic Apply Command ---

  Scenario: Applying feedback outputs human-readable summary
    Given a session "review-123" exists with annotations
    When I run the command `fabbro apply review-123`
    Then the output should display a human-readable summary of annotations
    And each annotation should show its type, line range, and content

  Scenario: Applying feedback as JSON
    Given a session "review-123" exists with annotations
    When I run the command `fabbro apply review-123 --json`
    Then the output should be valid JSON
    And the JSON should contain a "sessionId" field
    And the JSON should contain an "annotations" array

  # --- JSON Output Structure ---

  Scenario: JSON contains all annotation fields
    Given a session exists with a comment annotation on lines 42-45
    When I run the command `fabbro apply <session-id> --json`
    Then the JSON output should include:
      """
      {
        "sessionId": "<session-id>",
        "sourceFile": "stdin",
        "createdAt": "<timestamp>",
        "annotations": [
          {
            "type": "comment",
            "startLine": 42,
            "endLine": 45,
            "text": "This section needs more examples"
          }
        ]
      }
      """

  Scenario: JSON includes all annotation types
    Given a session exists with multiple annotation types:
      | type     | startLine | endLine | text                          |
      | comment  | 10        | 15      | Good explanation              |
      | delete   | 20        | 30      | Too verbose                   |
      | expand   | 40        | 42      | Add more examples             |
      | question | 50        | 50      | Why this approach?            |
      | keep     | 60        | 70      |                               |
      | unclear  | 80        | 85      | The logic here is confusing   |
    When I run the command `fabbro apply <session-id> --json`
    Then the JSON should contain 6 annotations
    And each annotation should have type, startLine, endLine, and text fields

  # --- FEM Parsing ---

  Scenario: Parsing inline comment annotation
    Given a session file contains:
      """
      Line 41 content
      Line 42 content {>> This is a comment <<}
      Line 43 content
      """
    When I run the command `fabbro apply <session-id> --json`
    Then the JSON should contain a "comment" annotation on line 42

  Scenario: Parsing block delete annotation
    Given a session file contains:
      """
      Line 9 content
      {-- DELETE: Too much detail --}
      Line 10 content
      Line 11 content
      Line 12 content
      {--/--}
      Line 13 content
      """
    When I run the command `fabbro apply <session-id> --json`
    Then the JSON should contain a "delete" annotation on lines 10-12
    And the annotation text should be "Too much detail"

  Scenario: Parsing question annotation
    Given a session file contains:
      """
      Line 50 content {?? Why not use X instead? ??}
      """
    When I run the command `fabbro apply <session-id> --json`
    Then the JSON should contain a "question" annotation on line 50

  Scenario: Parsing expand annotation
    Given a session file contains:
      """
      Line 40 content {!! EXPAND: Add error handling !!}
      """
    When I run the command `fabbro apply <session-id> --json`
    Then the JSON should contain an "expand" annotation on line 40

  Scenario: Parsing keep annotation
    Given a session file contains:
      """
      Line 60 content {== KEEP: Excellent explanation ==}
      """
    When I run the command `fabbro apply <session-id> --json`
    Then the JSON should contain a "keep" annotation on line 60

  Scenario: Parsing unclear annotation
    Given a session file contains:
      """
      Line 80 content {~~ UNCLEAR: What does this mean? ~~}
      """
    When I run the command `fabbro apply <session-id> --json`
    Then the JSON should contain an "unclear" annotation on line 80

  # --- Line Number Handling ---

  Scenario: Annotations reference original line numbers
    Given the session was created from content with 100 lines
    And annotations were added via TUI
    When I run the command `fabbro apply <session-id> --json`
    Then line numbers in annotations should match the original content
    And line numbers should NOT include the frontmatter offset

  Scenario: Multi-line annotations span correct range
    Given a session exists with an annotation spanning lines 42-50
    When I run the command `fabbro apply <session-id> --json`
    Then the annotation should have startLine 42 and endLine 50

  # --- Error Handling ---

  Scenario: Applying non-existent session
    Given no session "nonexistent" exists
    When I run the command `fabbro apply nonexistent`
    Then an error message should indicate the session was not found
    And the command should exit with code 1

  Scenario: Applying session with malformed FEM
    Given a session file contains invalid FEM syntax
    When I run the command `fabbro apply <session-id> --json`
    Then an error message should indicate the parsing error
    And the error should include the line number of the syntax error
    And the command should exit with code 1

  # --- Content Hash Verification ---

  Scenario: Warning when source content has changed
    Given a session was created from file "document.md"
    And "document.md" has been modified since the session was created
    When I run the command `fabbro apply <session-id>`
    Then a warning should indicate the source file has changed
    And the annotations should still be output
    And the warning should suggest line numbers may have drifted

  # --- Output Formats ---

  Scenario: Compact JSON output for piping
    Given a session exists with annotations
    When I run the command `fabbro apply <session-id> --json --compact`
    Then the JSON should be output on a single line
    And the output should be suitable for piping to other tools

  Scenario: Pretty-printed JSON output
    Given a session exists with annotations
    When I run the command `fabbro apply <session-id> --json --pretty`
    Then the JSON should be formatted with indentation
    And the output should be human-readable
