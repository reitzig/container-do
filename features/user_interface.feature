Feature: Limited User Interface
    Due to its nature, container-do has a very limited UI:
    it creates a template config file, and
    informs the user on how to use it.

    Background:
        Given docker is installed

    Scenario: No Config File
        Given an empty project test-app
        When  container-do is called with `sh -c 'exit 77'`
        Then  the command exits with status 1
        And   no container was started

    Scenario: Initialize Config File
        Given an empty project test-app
        When  container-do is called with `--init`
        Then  the command exits with status 0
        Then  file ContainerDo.toml is a valid config file
        And  no container was started

    Scenario: Do not overwrite Config File
        Given config file for project test-app
            """
            [container]
            image = "ubuntu"
            """
        When  container-do is called with `--init`
        Then  the command exits with status 1
        And   file ContainerDo.toml still contains
            """
            [container]
            image = "ubuntu"
            """
        And  no container was started

    Scenario: Print Help
        When  container-do is called with `--help`
        Then  the command exits with status 0
        And   the command output contains "Usage:"
        And   no container was started
