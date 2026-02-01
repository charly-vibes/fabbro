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

  @implemented
  Scenario: Center cursor in viewport
    Given the cursor is on line 50
    When I press "zz"
    Then the viewport should scroll to center line 50 vertically
    When I press "zt"
    Then the viewport should scroll to place line 50 at the top
    When I press "zb"
    Then the viewport should scroll to place line 50 at the bottom

  @implemented
  Scenario: Search within document
    When I press "/"
    Then a search prompt should appear
    When I type "function" and press Enter
    Then the cursor should jump to the first match
    And the match should be highlighted
    When I press "n"
    Then the cursor should jump to the next match
    When I press "N"
    Then the cursor should jump to the previous match

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
      | r   | replace  | suggest replacement   |

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

  @implemented
  Scenario: Expand selection to paragraph
    Given the cursor is on line 42 within a paragraph
    When I press "v" to start selection
    And I press "ap" (a paragraph)
    Then the selection should expand to include the entire paragraph
    And paragraph boundaries are determined by blank lines

  @implemented
  Scenario: Expand selection to code block
    Given the cursor is inside a fenced code block
    When I press "v" to start selection
    And I press "ab" (a block)
    Then the selection should expand to include the entire code block
    And the ``` delimiters should be included

  @implemented
  Scenario: Expand selection to section
    Given the cursor is under a markdown heading
    When I press "v" to start selection
    And I press "as" (a section)
    Then the selection should expand to include content until the next heading

  @implemented
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
  Scenario: Adding a change annotation (suggest replacement)
    Given I have lines 42-45 selected
    When I press "r" (or SPC then r)
    Then a text input prompt should appear with label "Replacement text:"
    When I type "const result = calculate()" and press Enter
    Then an annotation of type "change" should be added to lines 42-45
    And the annotation text should contain "[lines 42-45] -> const result = calculate()"
    And the selected lines should be visually marked as having a change suggestion

  @implemented
  Scenario: Canceling annotation input
    Given I have lines selected
    And an annotation prompt is open
    When I press Escape
    Then the prompt should close
    And no annotation should be added
    And the selection should remain

  @implemented
  Scenario: Text input wraps when content is long
    Given an annotation prompt is open
    When I type text that exceeds the input width
    Then the text should wrap to the next line
    And all content should remain visible

  @implemented
  Scenario: Adding newlines in annotation input
    Given an annotation prompt is open
    When I press Shift+Enter
    Then a newline should be inserted in the input
    And I should be able to continue typing on the new line
    When I press Enter (without Shift)
    Then the annotation should be submitted with the multiline content

  # --- Editing Existing Annotations ---

  @implemented
  Scenario: Editing annotation text on current line
    Given I have added a comment annotation on line 42
    And the cursor is on line 42
    When I press "e" (with no selection active)
    Then an editor should open with the annotation text pre-filled
    When I modify the text and press Ctrl+S
    Then the annotation text should be updated
    And the editor should close

  @implemented
  Scenario: Picking annotation when multiple exist on same line
    Given I have added a comment annotation on line 42
    And I have added a question annotation on line 42
    And the cursor is on line 42
    When I press "e" (with no selection active)
    Then an annotation picker should appear
    And it should list both annotations with their types and preview text
    When I select one and press Enter
    Then the editor should open with that annotation's text

  @planned
  Scenario: Editing annotation range
    Given I have added a comment annotation on lines 42-45
    And the cursor is on line 42
    When I press "R" (with no selection active)
    Then the selection should activate on lines 42-45 (the annotation's range)
    When I extend the selection to line 50 using j/k
    And I press Enter
    Then the annotation range should be updated to lines 42-50
    And the selection should be cleared

  @implemented
  Scenario: Canceling annotation edit
    Given I have added a comment annotation on line 42
    And the cursor is on line 42
    When I press "e" to edit the annotation
    And the editor opens with the annotation text
    When I press Escape twice to cancel
    Then the editor should close
    And the annotation should remain unchanged

  @implemented
  Scenario: No annotation on current line
    Given the cursor is on line 42
    And there are no annotations on line 42
    When I press "e" (with no selection active)
    Then an error message should display "No annotation on this line"

  # --- Viewing Annotations ---

  @implemented
  Scenario: Visual indicator for annotated lines
    Given I have added annotations to lines 5, 10, and 15
    When I view the document in the TUI
    Then lines 5, 10, and 15 should show a "●" indicator after the line number
    And unannotated lines should show a space in the indicator column

  @implemented
  Scenario: Show annotation preview when cursor on annotated line
    Given I have a session with content "line1\nline2\nline3"
    And line 2 has a comment annotation "This needs review"
    When I move cursor to line 2
    Then I should see the annotation preview panel
    And the preview shows "comment [2-2]"
    And the preview shows "This needs review"

  @implemented
  Scenario: Multiple annotations show count
    Given I have a session with content "line1\nline2"
    And line 2 has a comment annotation "First note"
    And line 2 has a question annotation "Is this correct?"
    When I move cursor to line 2
    Then I should see "(1 of 2 annotations)"
    And I should see the first annotation content

  @implemented
  Scenario: Annotation preview disappears when cursor leaves annotated line
    Given I have a session with content "line1\nline2\nline3"
    And line 2 has a comment annotation "Some note"
    When I move cursor to line 2
    Then I should see the annotation preview panel
    When I move cursor to line 1
    Then I should see the normal help text
    And the preview panel should not be visible

  @implemented
  Scenario: Annotation range highlighting in preview
    Given I have a session with content "line1\nline2\nline3\nline4\nline5"
    And lines 2-4 have a comment annotation "Multi-line note"
    When I move cursor to line 2
    Then I should see the annotation preview panel
    And lines 2, 3, and 4 should show a "▐" range highlight indicator
    When I move cursor to line 5
    Then the range highlight indicators should disappear

  @implemented
  Scenario: Tab cycling updates annotation range highlighting
    Given I have a session with content "line1\nline2\nline3\nline4"
    And line 2 has a comment annotation "First" spanning lines 2-2
    And line 2 has a question annotation "Second" spanning lines 2-4
    When I move cursor to line 2
    Then only line 2 should show the range highlight indicator
    When I press Tab
    Then lines 2, 3, and 4 should show the range highlight indicator

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

  # --- Direct Content Editing ---

  @implemented
  Scenario: Opening inline editor for direct content changes
    Given I have lines 42-45 selected
    When I press "i" (or SPC then i)
    Then an inline text editor should open
    And it should show the selected content
    And I should be able to edit the text directly

  @implemented
  Scenario: Saving inline edit
    Given the inline editor is open with modified content
    When I press Ctrl+S (or Ctrl+Enter)
    Then a "change" annotation should be created
    And the annotation should contain my edited version as the replacement
    And the annotation replacement contains \\n for new lines
    And the inline editor should close

  @implemented
  Scenario: Canceling inline edit
    Given the inline editor is open
    When I press Escape twice (or Ctrl+C)
    Then the inline editor should close
    And no annotation should be created

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

  @implemented
  Scenario: Viewing help
    When I press "?"
    Then a help panel should appear
    And it should display all available keybindings
    When I press any key
    Then the help panel should close
