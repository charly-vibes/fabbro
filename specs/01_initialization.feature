Feature: Project Initialization
  As a user, I want to initialize `fabbro` in my project directory
  so that the necessary file structure is created for managing review sessions.

  # Implementation Status Legend:
  # @implemented - Working in current build
  # @partial     - Core works, some aspects missing
  # @planned     - Designed but not yet implemented

  @partial
  Scenario: Initializing a new project
    # Note: Creates .fabbro/sessions/ but not templates/, config.yaml, or .gitignore
    Given I am in a directory that has not been initialized
    When I run the command `fabbro init`
    Then a directory named ".fabbro" should be created in the current directory
    And the ".fabbro" directory should contain a subdirectory named "sessions"
    And the ".fabbro" directory should contain a subdirectory named "templates"
    And the ".fabbro" directory should contain a configuration file named "config.yaml"
    And the ".fabbro" directory should contain a ".gitignore" file
    And the ".fabbro/.gitignore" file should contain the line "sessions/"
    And the command should exit with code 0

  @implemented
  Scenario: Initializing an already initialized project
    Given I am in a directory that has already been initialized
    When I run the command `fabbro init`
    Then the existing ".fabbro" directory should not be modified
    And a message should indicate the project is already initialized
    And the command should exit with code 0

  @planned
  Scenario: Quiet initialization
    Given I am in a directory that has not been initialized
    When I run the command `fabbro init --quiet`
    Then a directory named ".fabbro" should be created in the current directory
    And no output should be printed to stdout
    And the command should exit with code 0

  @planned
  Scenario: Initializing in a subdirectory of an initialized project
    Given I am in a subdirectory of a project that has been initialized
    When I run the command `fabbro init`
    Then a new ".fabbro" directory should be created in the current subdirectory
    And a warning should indicate a parent directory is already initialized

  @planned
  Scenario: Initializing with agent integration scaffolding
    # Creates custom commands for various AI coding agents (Amp, Claude Code, Cursor)
    Given I am in a directory that has not been initialized
    When I run the command `fabbro init --agents`
    Then a directory named ".fabbro" should be created in the current directory
    And the ".fabbro" directory should contain a subdirectory named "sessions"
    And a directory named ".agents/commands" should be created
    And the ".agents/commands" directory should contain "fabbro-review.md"
    And a directory named ".claude/commands" should be created
    And the ".claude/commands" directory should contain "fabbro-review.md"
    And the command files should contain fabbro workflow instructions
    And the command should exit with code 0

  @planned
  Scenario: Initializing with agents updates AGENTS.md
    Given I am in a directory with an existing AGENTS.md file
    When I run the command `fabbro init --agents`
    Then the AGENTS.md file should be updated with a fabbro workflow section
    And the fabbro workflow section should document the review process
    And existing AGENTS.md content should be preserved

  @planned
  Scenario: Agent scaffolding detects available agents
    # Only creates command files for agents that appear to be in use
    Given I am in a directory with a ".claude" directory
    And no ".cursor" directory exists
    When I run the command `fabbro init --agents`
    Then the ".claude/commands" directory should contain "fabbro-review.md"
    And no ".cursor/commands" directory should be created
    And the ".agents/commands" directory should always be created
