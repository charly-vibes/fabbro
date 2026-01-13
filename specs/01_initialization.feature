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
