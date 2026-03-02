Feature: Web DOCX File Upload
  As a user, I want to drop or select a .docx file in the web app
  so that I can review Word documents without converting them manually.

  # Implementation Status Legend:
  # @implemented - Working in current build
  # @planned     - Designed but not yet implemented

  Background:
    Given I am on the fabbro web app landing page

  # --- Drag and drop ---

  @planned
  Scenario: Dropping a .docx file starts a review session
    When I drop a file named "report.docx" onto the drop zone
    Then the text content should be extracted from the .docx file
    And a new review session should start with the extracted text
    And the filename should show "report.docx"

  @planned
  Scenario: Extracted text preserves paragraph structure
    Given I have a .docx file with multiple paragraphs and headings
    When I drop the file onto the drop zone
    Then paragraphs should be separated by blank lines
    And heading text should be preserved as plain text
    And no HTML tags should appear in the content

  # --- Drop zone label ---

  @planned
  Scenario: Drop zone label includes .docx
    Then the drop zone should mention ".docx" as a supported format

  # --- Error handling ---

  @planned
  Scenario: Handling a corrupt .docx file
    When I drop a corrupt or invalid .docx file onto the drop zone
    Then an error message should say "Could not read .docx file. The file may be corrupt."
    And I should remain on the landing page

  @planned
  Scenario: Handling an empty .docx file
    When I drop a .docx file that contains no text
    Then an error message should say "This document appears to be empty."
    And I should remain on the landing page

  @planned
  Scenario: Rejecting .doc (legacy Word) files
    When I drop a file named "legacy.doc" onto the drop zone
    Then an error message should say "Legacy .doc files are not supported. Please save as .docx."
