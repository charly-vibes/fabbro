Feature: Web Notes Sidebar
  As a reviewer using the web interface,
  I want a sidebar listing all my annotations
  so that I can navigate, review, and manage them easily.

  # Implementation Status Legend:
  # @implemented - Working in current build
  # @planned     - Designed but not yet implemented

  # --- Sidebar Display ---

  @implemented
  Scenario: Notes panel appears in editor view
    Given I am in the editor view with a loaded document
    Then a notes panel should be visible on the right side
    And it should show a header with the annotation count

  @implemented
  Scenario: Empty state when no annotations exist
    Given I am in the editor view with no annotations
    Then the notes panel should display "No annotations yet. Select text to add a comment."

  @implemented
  Scenario: Note card displays annotation details
    Given I have a "comment" annotation on "some selected text" with text "Fix this"
    Then the notes panel should show a card with:
      | field   | value            |
      | badge   | Comment          |
      | snippet | some selected text |
      | text    | Fix this         |
      | line    | L1               |

  @implemented
  Scenario: Snippet preview is truncated at 60 characters
    Given I have an annotation on a selection longer than 60 characters
    Then the snippet preview should show the first 60 characters followed by "…"

  @implemented
  Scenario: Notes are sorted by position in document
    Given I have annotations at offsets 100, 10, and 50
    Then the notes panel should list them in order: offset 10, 50, 100

  @implemented
  Scenario: Counter updates when annotations change
    Given I have 3 annotations
    Then the notes header should display "Notes (3)"
    When I add another annotation
    Then the notes header should display "Notes (4)"

  # --- Type Badges ---

  @implemented
  Scenario: Comment annotation shows Comment badge
    Given I have a "comment" annotation
    Then its note card should show a "Comment" badge with blue styling

  @implemented
  Scenario: Suggest annotation shows Suggest badge
    Given I have a "suggest" annotation
    Then its note card should show a "Suggest" badge with green styling

  # --- Navigation: Note to Highlight ---

  @implemented
  Scenario: Clicking a note scrolls to its highlight
    Given I have an annotation on line 50
    When I click the note card in the sidebar
    Then the document should scroll to the highlighted text on line 50
    And the highlight should briefly flash with an active outline

  # --- Navigation: Highlight to Note ---

  @implemented
  Scenario: Clicking a highlight scrolls to its note
    Given I have an annotation with a highlighted region in the document
    When I click the highlighted text in the viewer
    Then the sidebar should scroll to the corresponding note card
    And the note card should briefly flash with an active outline

  # --- Delete ---

  @implemented
  Scenario: Deleting an annotation via the sidebar
    Given I have 2 annotations
    When I click the delete button on the first note card
    Then the annotation should be removed from the session
    And the notes panel should show "Notes (1)"
    And the highlight should be removed from the document

  @implemented
  Scenario: Delete button does not trigger note click
    Given I have an annotation in the sidebar
    When I click the delete button
    Then the document should not scroll to the highlight
    And the annotation should be deleted
