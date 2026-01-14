Feature: FEM Markup Language
  As a power user or developer,
  I want to understand the FEM (Fabbro Editor Markup) syntax
  so that I can manually edit session files if needed.

  # Implementation Status Legend:
  # @implemented - Working in current build
  # @planned     - Designed but not yet implemented

  # FEM is designed to be inline, non-destructive markup that can be
  # added to any text content without breaking the original format.

  # --- Comment Annotation ---

  @implemented
  Scenario: Inline comment syntax
    Given content with the following FEM markup:
      """
      This is a paragraph. {>> This needs clarification <<}
      """
    When the FEM is parsed
    Then a "comment" annotation should be extracted
    And the annotation text should be "This needs clarification"
    And the annotation should be on line 1

  @planned
  Scenario: Comment with line reference (sidecar style)
    Given content with the following FEM markup:
      """
      Original content here.
      
      {>> [line 1] This needs clarification <<}
      """
    When the FEM is parsed
    Then a "comment" annotation should be extracted for line 1

  # --- Delete Annotation ---

  @planned
  Scenario: Block delete with reason
    # Block markers {--/--} not implemented
    Given content with the following FEM markup:
      """
      Keep this line.
      {-- DELETE: Too verbose --}
      Delete this line.
      And this line too.
      {--/--}
      Keep this line as well.
      """
    When the FEM is parsed
    Then a "delete" annotation should be extracted
    And the annotation should span lines 3-4
    And the annotation text should be "Too verbose"

  @implemented
  Scenario: Inline delete (single line)
    Given content with the following FEM markup:
      """
      This line is fine. {-- DELETE: Remove this redundant sentence --}
      """
    When the FEM is parsed
    Then a "delete" annotation should be extracted for line 1

  # --- Question Annotation ---

  @implemented
  Scenario: Question syntax
    Given content with the following FEM markup:
      """
      The algorithm uses a greedy approach. {?? Why not dynamic programming? ??}
      """
    When the FEM is parsed
    Then a "question" annotation should be extracted
    And the annotation text should be "Why not dynamic programming?"

  # --- Expand Annotation ---

  @implemented
  Scenario: Expand syntax
    Given content with the following FEM markup:
      """
      Error handling is important. {!! EXPAND: Add examples of error cases !!}
      """
    When the FEM is parsed
    Then an "expand" annotation should be extracted
    And the annotation text should be "Add examples of error cases"

  # --- Keep Annotation ---

  @implemented
  Scenario: Keep syntax (mark as good)
    Given content with the following FEM markup:
      """
      {== KEEP: Excellent explanation of the concept ==}
      This paragraph explains the core concept clearly...
      """
    When the FEM is parsed
    Then a "keep" annotation should be extracted
    And the annotation text should be "Excellent explanation of the concept"

  # --- Unclear Annotation ---

  @implemented
  Scenario: Unclear syntax
    Given content with the following FEM markup:
      """
      The flux capacitor inverts the polarity. {~~ UNCLEAR: What does this mean? ~~}
      """
    When the FEM is parsed
    Then an "unclear" annotation should be extracted
    And the annotation text should be "What does this mean?"

  # --- Change Annotation ---

  @implemented
  Scenario: Change annotation syntax (inline replacement suggestion)
    Given content with the following FEM markup:
      """
      const x = foo(); {++ [line 1] -> const x = bar() ++}
      """
    When the FEM is parsed
    Then a "change" annotation should be extracted
    And the annotation text should be "[line 1] -> const x = bar()"

  @implemented
  Scenario: Multi-line change annotation
    Given content with the following FEM markup:
      """
      old code here {++ [lines 1-3] -> new code here ++}
      more old code
      end of old code
      """
    When the FEM is parsed
    Then a "change" annotation should be extracted
    And the annotation text should contain "[lines 1-3] ->"

  # --- Emphasize Annotation ---

  @planned
  Scenario: Emphasize syntax
    # {** ... **} not implemented
    Given content with the following FEM markup:
      """
      {** EMPHASIZE: This is the key takeaway **}
      Users should always validate input before processing.
      """
    When the FEM is parsed
    Then an "emphasize" annotation should be extracted

  # --- Section-level Annotation ---

  @planned
  Scenario: Section annotation
    # {## ... ##} not implemented
    Given content with the following FEM markup:
      """
      {## SECTION: This entire section needs rewriting ##}
      ## Introduction
      The introduction text...
      """
    When the FEM is parsed
    Then a "section" annotation should be extracted

  # --- Multiple Annotations on Same Line ---

  @implemented
  Scenario: Multiple annotations on single line
    Given content with the following FEM markup:
      """
      Important concept here. {>> Good <<} {!! EXPAND: Add more detail !!}
      """
    When the FEM is parsed
    Then 2 annotations should be extracted
    And one should be type "comment"
    And one should be type "expand"

  # --- Escaping FEM Syntax ---

  @planned
  Scenario: Escaped markup is not parsed
    # Escaping not implemented
    Given content with the following FEM markup:
      """
      To add a comment, use the syntax \{>> comment <<\}
      """
    When the FEM is parsed
    Then no annotations should be extracted
    And the content should preserve the literal markup characters

  # --- Session File Format ---

  @implemented
  Scenario: Session file with YAML frontmatter
    Given a session file with content:
      """
      ---
      session_id: abc123
      created_at: 2026-01-11T10:00:00Z
      source: stdin
      content_hash: sha256:abcdef123456
      ---
      
      # Document Title
      
      Content starts here. {>> First annotation <<}
      """
    When the FEM is parsed
    Then the session_id should be "abc123"
    And the annotation should reference line 3 of the content (not including frontmatter)

  # --- Whitespace Handling ---

  @implemented
  Scenario: Annotations preserve surrounding whitespace
    Given content with the following FEM markup:
      """
      Text before {>> annotation <<} text after.
      """
    When the FEM is parsed and rendered without annotations
    Then the output should be "Text before  text after."

  @planned
  Scenario: Newlines in annotation text
    # Multi-line annotation text not supported
    Given content with the following FEM markup:
      """
      {>> This is a multi-line
      annotation that spans
      multiple lines <<}
      """
    When the FEM is parsed
    Then the annotation text should preserve the newlines

  # --- Edge Cases ---

  @implemented
  Scenario: Empty annotation text
    Given content with the following FEM markup:
      """
      Content here. {>> <<}
      """
    When the FEM is parsed
    Then the annotation should have empty text
    And parsing should not fail

  @planned
  Scenario: Nested braces in annotation text
    # Nested braces not handled correctly
    Given content with the following FEM markup:
      """
      Code example. {>> Use {curly braces} in the output <<}
      """
    When the FEM is parsed
    Then the annotation text should be "Use {curly braces} in the output"

  @planned
  Scenario: Unclosed annotation marker
    # No syntax error reporting
    Given content with the following FEM markup:
      """
      Content here. {>> This annotation is not closed
      """
    When the FEM is parsed
    Then a parsing error should be reported
    And the error should indicate the unclosed marker on line 1

  # --- FEM Syntax Summary ---

  # Reference table for FEM markup:
  # | Marker      | Type     | Description                    |
  # |-------------|----------|--------------------------------|
  # | {>> ... <<} | comment  | General feedback               |
  # | {-- ... --} | delete   | Mark for deletion (inline)     |
  # | {--/--}     | delete   | End of block deletion          |
  # | {?? ... ??} | question | Ask a question                 |
  # | {!! ... !!} | expand   | Request more detail            |
  # | {== ... ==} | keep     | Mark as good/preserve          |
  # | {~~ ... ~~} | unclear  | Flag as confusing              |
  # | {++ ... ++} | change   | Suggest replacement text       |
  # | {** ... **} | emphasize| Request emphasis               |
  # | {## ... ##} | section  | Section-level feedback         |
