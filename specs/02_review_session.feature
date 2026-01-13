Feature: Review Session Creation
  As a user, I want to create a review session for LLM-generated content
  so that I can annotate and provide feedback on long documents.

  # Implementation Status Legend:
  # @implemented - Working in current build
  # @planned     - Designed but not yet implemented

  Background:
    Given I am in a directory that has been initialized with `fabbro init`

  # --- Creating sessions from different sources ---

  @implemented
  Scenario: Creating a review session from stdin
    Given I have content piped to stdin
    When I run the command `fabbro review --stdin`
    Then a new session file should be created in ".fabbro/sessions/"
    And the session file should have a ".fem" extension
    And the session file should contain the piped content
    And the TUI should open with the content displayed

  @planned
  Scenario: Creating a review session from a file
    Given a file named "document.md" exists with content
    When I run the command `fabbro review document.md`
    Then a new session file should be created in ".fabbro/sessions/"
    And the session file should contain the content from "document.md"
    And the TUI should open with the content displayed

  @planned
  Scenario: Creating a review session with a custom session ID
    Given I have content piped to stdin
    When I run the command `fabbro review --stdin --id my-review`
    Then a session file named "my-review.fem" should be created
    And the TUI should open with the content displayed

  # --- Session file format ---

  @implemented
  Scenario: Session file contains metadata header
    Given I have content piped to stdin
    When I run the command `fabbro review --stdin`
    Then the session file should contain a YAML frontmatter header
    And the frontmatter should include "session_id"
    And the frontmatter should include "created_at" timestamp
    And the frontmatter should include "source" as "stdin"

  @implemented
  Scenario: Session file preserves original content
    Given I have the following content piped to stdin:
      """
      # My Document
      
      This is paragraph one.
      
      This is paragraph two.
      """
    When I run the command `fabbro review --stdin`
    Then the session file should contain the original content unchanged
    And line numbers in the TUI should match the original content

  # --- Error handling ---

  @implemented
  Scenario: Attempting to review without initialization
    Given I am in a directory that has NOT been initialized
    When I run the command `fabbro review --stdin`
    Then an error message should indicate the project is not initialized
    And the command should exit with code 1

  @planned
  Scenario: Attempting to review a non-existent file
    # Depends on file input support
    Given no file named "missing.md" exists
    When I run the command `fabbro review missing.md`
    Then an error message should indicate the file was not found
    And the command should exit with code 1

  @implemented
  Scenario: Attempting to review with no input
    Given no content is piped to stdin
    And no file argument is provided
    When I run the command `fabbro review`
    Then an error message should indicate no input was provided
    And the command should suggest using --stdin or providing a file path

  # --- Editor fallback mode ---

  @planned
  Scenario: Opening session in external editor instead of TUI
    Given I have content piped to stdin
    When I run the command `fabbro review --stdin --editor`
    Then the session file should be created
    And the $EDITOR should be opened with the session file
    And the TUI should NOT be launched

  @planned
  Scenario: Non-interactive mode creates session without opening anything
    Given I have content piped to stdin
    When I run the command `fabbro review --stdin --no-interactive`
    Then the session file should be created
    And the session ID should be printed to stdout
    And neither TUI nor editor should be opened
