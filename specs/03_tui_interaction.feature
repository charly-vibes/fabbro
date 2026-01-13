Feature: TUI Interaction
  As a user reviewing content in the TUI,
  I want to navigate, select text, and add annotations
  so that I can provide feedback without memorizing markup syntax.

  # Implementation Status Legend:
  # @implemented - Working in current build
  # @planned     - Designed but not yet implemented
  
  # Authoritative keybinding reference: docs/keybindings.md

  Background:
    Given I am in a review session with the TUI open
    And the document has 100 lines of content

  # --- Navigation ---

  @implemented
  Scenario: Navigating with keyboard
    When I press "j" or the down arrow
    Then the cursor should move down one line
    When I press "k" or the up arrow
    Then the cursor should move up one line

  @implemented
  Scenario: Page navigation
    When I press "Ctrl+d" or "Page Down"
    Then the viewport should scroll down half a page
    When I press "Ctrl+u" or "Page Up"
    Then the viewport should scroll up half a page

  @implemented
  Scenario: Jump to beginning and end
    When I press "g" twice (gg)
    Then the cursor should jump to the first line
    When I press "G"
    Then the cursor should jump to the last line

  @planned
  Scenario: Search within document
    When I press "/"
    Then a search prompt should appear
    When I type "function" and press Enter
    Then the cursor should jump to the first match
    And the match should be highlighted

  # --- Helix-style SPC Menu (Discoverability) ---

  @implemented
  Scenario: Opening the command palette
    When I press Space (SPC)
    Then a command palette menu should appear
    And the menu should display available annotation types:
      | key | label    | description           |
      | c   | comment  | add comment           |
      | d   | delete   | mark for deletion     |
      | e   | expand   | request more detail   |
      | q   | question | ask a question        |
      | k   | keep     | mark as good          |
      | u   | unclear  | flag confusion        |

  @implemented
  Scenario: Selecting action from command palette
    Given the command palette is open
    When I press "c"
    Then the command palette should close
    And an annotation prompt for "comment" should appear

  @implemented
  Scenario: Dismissing the command palette
    Given the command palette is open
    When I press Escape
    Then the command palette should close
    And no action should be taken

  # --- Text Selection ---

  @implemented
  Scenario: Selecting a single line
    Given the cursor is on line 42
    When I press "v" to start selection
    Then line 42 should be marked as selected
    And the status bar should show "1 line selected"

  @implemented
  Scenario: Selecting a range of lines
    Given the cursor is on line 42
    When I press "v" to start selection
    And I move the cursor to line 50
    Then lines 42-50 should be marked as selected
    And the status bar should show "9 lines selected"

  @implemented
  Scenario: Canceling selection
    Given I have lines 42-50 selected
    When I press Escape
    Then the selection should be cleared
    And the cursor should remain on line 50

  # --- Vim-surround Style Context Expansion ---

  @planned
  Scenario: Expand selection to paragraph
    Given the cursor is on line 42 within a paragraph
    When I press "v" to start selection
    And I press "ap" (a paragraph)
    Then the selection should expand to include the entire paragraph
    And paragraph boundaries are determined by blank lines

  @planned
  Scenario: Expand selection to code block
    Given the cursor is inside a fenced code block
    When I press "v" to start selection
    And I press "ab" (a block)
    Then the selection should expand to include the entire code block
    And the ``` delimiters should be included

  @planned
  Scenario: Expand selection to section
    Given the cursor is under a markdown heading
    When I press "v" to start selection
    And I press "as" (a section)
    Then the selection should expand to include content until the next heading

  @planned
  Scenario: Shrink and grow selection by line
    Given I have lines 40-50 selected
    When I press "{"
    Then the selection should shrink by one line from the end
    When I press "}"
    Then the selection should grow by one line

  # --- Adding Annotations ---

  @implemented
  Scenario: Adding a comment annotation
    Given I have lines 42-45 selected
    When I press "c" (or SPC then c)
    Then a text input prompt should appear with label "Comment:"
    When I type "This section needs more examples" and press Enter
    Then an annotation of type "comment" should be added to lines 42-45
    And the selected lines should be visually marked as annotated
    And the selection should be cleared

  @implemented
  Scenario: Adding a delete annotation
    Given I have lines 10-20 selected
    When I press "d" (or SPC then d)
    Then a text input prompt should appear with label "Reason for deletion:"
    When I type "Too verbose" and press Enter
    Then an annotation of type "delete" should be added to lines 10-20
    And the selected lines should be visually marked for deletion

  @implemented
  Scenario: Adding a question annotation
    # Note: "q" for question requires selection; "Q" quits (case-sensitive)
    Given the cursor is on line 72
    When I press "q" (or SPC then q)
    Then a text input prompt should appear with label "Question:"
    When I type "Why not use dependency injection here?" and press Enter
    Then an annotation of type "question" should be added to line 72

  @implemented
  Scenario: Adding an expand annotation
    Given I have lines 50-55 selected
    When I press "e" (or SPC then e)
    Then a text input prompt should appear with label "What to expand:"
    When I type "Add error handling examples" and press Enter
    Then an annotation of type "expand" should be added to lines 50-55

  @implemented
  Scenario: Adding a keep annotation (mark as good)
    Given I have lines 80-90 selected
    When I press "k" (or SPC then k)
    Then an annotation of type "keep" should be added without prompting for text
    And the selected lines should be visually marked as "keep"

  @implemented
  Scenario: Canceling annotation input
    Given I have lines selected
    And an annotation prompt is open
    When I press Escape
    Then the prompt should close
    And no annotation should be added
    And the selection should remain

  # --- Viewing Annotations ---

  @planned
  Scenario: Viewing all annotations in session
    Given I have added 5 annotations to the document
    When I press "a" (or SPC then a)
    Then an annotations panel should appear
    And it should list all 5 annotations with their line numbers and types

  @planned
  Scenario: Jumping to annotation from list
    Given the annotations panel is open
    When I select an annotation and press Enter
    Then the viewport should jump to that annotation's location
    And the annotations panel should close

  # --- Mouse Interaction (Phase 1.5) ---

  @planned
  Scenario: Clicking to position cursor
    When I click on line 42
    Then the cursor should move to line 42

  @planned
  Scenario: Click-drag to select range
    When I click on line 42 and drag to line 50
    Then lines 42-50 should be selected

  @planned
  Scenario: Right-click context menu
    Given the cursor is on line 42
    When I right-click
    Then a context menu should appear with annotation options
    And the menu should match the SPC command palette options

  # --- Saving and Exiting ---

  @implemented
  Scenario: Saving and exiting the review
    Given I have added annotations to the document
    When I press "w" (or Ctrl+S)
    Then all annotations should be saved to the session file
    And the TUI should exit
    And the session ID should be printed to stdout

  @implemented
  Scenario: Quitting without saving
    # Note: "Q" (uppercase) or Ctrl+C quits immediately without saving
    # This differs from the original design which prompted for confirmation
    Given I have added annotations that are not saved
    When I press "Q" or "Ctrl+C"
    Then the TUI should exit immediately without saving

  @planned
  Scenario: Exiting with confirmation prompt
    # Future: Add confirmation before discarding unsaved changes
    Given I have added annotations that are not saved
    When I press "Ctrl+C"
    Then a confirmation prompt should appear
    And it should ask "Save changes before exiting? (y/n/cancel)"

  @planned
  Scenario: Viewing help
    When I press "?"
    Then a help panel should appear
    And it should display all available keybindings
    When I press any key
    Then the help panel should close
