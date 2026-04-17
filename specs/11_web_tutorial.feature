Feature: Web Interactive Tutorial
  As a first-time reviewer using the web interface,
  I want an interactive tutorial that guides me through selecting, annotating, and exporting
  so that I can learn the review workflow quickly on desktop or touch devices.

  # Implementation Status Legend:
  # @implemented - Working in current build
  # @planned     - Designed but not yet implemented

  # --- Tutorial Entry ---

  @implemented
  Scenario: Start the tutorial from the landing page
    Given I am on the web landing page
    When I click "New here? Try the interactive tutorial"
    Then a tutorial panel should appear
    And a sample review document should load in the editor

  # --- Selection Guidance ---

  @implemented
  Scenario: Tutorial explains desktop and touch selection
    Given the tutorial is showing the "Select text" step
    Then it should explain click-and-drag selection for desktop users
    And it should explain long-press selection with native handles for touch users

  @implemented
  Scenario: Tutorial advances after touch selection in the viewer
    Given the tutorial is waiting for a text selection step
    When I select non-empty text inside the viewer on a touch device
    Then the tutorial should advance to the annotation type step

  # --- Annotation Guidance ---

  @implemented
  Scenario: Tutorial explains mobile-friendly save action
    Given the tutorial is showing the "Write your note" step
    Then it should explain pressing Enter to save on desktop
    And it should explain using the on-screen "Save" button on touch devices

  # --- Completion ---

  @implemented
  Scenario: Exit the tutorial
    Given the tutorial panel is visible
    When I click the close button
    Then the tutorial panel should disappear
