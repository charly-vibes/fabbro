Feature: Web HTML-to-Text Fallback
  As a user, I want HTML responses from URL fetching to be automatically
  converted to readable text so that I can review content from any
  CORS-accessible website without seeing raw HTML tags.

  # Implementation Status Legend:
  # @implemented - Working in current build
  # @planned     - Designed but not yet implemented

  Background:
    Given I am on the fabbro web app
    And I enter a non-GitHub URL in the URL input

  # --- Content-Type detection ---

  @planned
  Scenario: Markdown response is used as-is
    Given the URL returns Content-Type "text/markdown"
    When the content is fetched
    Then the raw response text should be used without conversion
    And the content type should be reported as "text/markdown"

  @planned
  Scenario: HTML response is converted to plain text
    Given the URL returns Content-Type "text/html"
    And the response body contains "<h1>Title</h1><p>Hello world</p>"
    When the content is fetched
    Then the content should contain "Title" and "Hello world"
    And no HTML tags should appear in the content
    And the content type should be reported as "text/html"

  @planned
  Scenario: Plain text response is used as-is
    Given the URL returns Content-Type "text/plain"
    When the content is fetched
    Then the raw response text should be used without conversion

  # --- Non-content element stripping ---

  @planned
  Scenario: Script and style elements are removed
    Given the URL returns HTML containing script and style tags
    When the content is fetched
    Then no JavaScript code should appear in the content
    And no CSS rules should appear in the content

  @planned
  Scenario: Navigation and footer elements are removed
    Given the URL returns HTML with nav, header, and footer elements
    When the content is fetched
    Then navigation text should not appear in the content
    And footer text should not appear in the content

  # --- Content extraction priority ---

  @planned
  Scenario: Article element content is preferred
    Given the URL returns HTML with a nav and an article element
    When the content is fetched
    Then the content should come from the article element
    And navigation text should not appear

  @planned
  Scenario: Main element content is preferred when no article exists
    Given the URL returns HTML with a main element but no article
    When the content is fetched
    Then the content should come from the main element

  @planned
  Scenario: Body content is used as fallback
    Given the URL returns HTML with no article or main element
    When the content is fetched
    Then the content should come from the body element
    With non-content elements stripped

  # --- Response metadata ---

  @planned
  Scenario: Markdown token count is surfaced when available
    Given the URL returns a response with header "x-markdown-tokens: 3150"
    When the content is fetched
    Then the result should include a markdown token count of 3150

  @planned
  Scenario: Missing token count header returns null
    Given the URL returns a response without the x-markdown-tokens header
    When the content is fetched
    Then the result should include a null markdown token count

  # --- No regression ---

  @planned
  Scenario: GitHub URLs still use the GitHub API
    Given I enter a GitHub file URL
    When the content is fetched
    Then the GitHub API should be used to fetch raw content
    And no HTML-to-text conversion should occur
