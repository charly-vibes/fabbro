Feature: Project Initialization
  As a user, I want to initialize `fabbro` in my project directory
  so that the necessary file structure is created for managing review sessions.

  Scenario: Initializing a new project
    Given I am in a directory that has not been initialized
    When I run the command `fabbro init`
    Then a directory named ".fabbro" should be created in the current directory
    And the ".fabbro" directory should contain a subdirectory named "sessions"
    And the ".fabbro" directory should contain a subdirectory named "templates"
    And the ".fabbro" directory should contain a configuration file named "config.yaml"
    And the ".fabbro" directory should contain a ".gitignore" file
    And the ".fabbro/.gitignore" file should contain the line "sessions/"
