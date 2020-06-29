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
        Then  file ContainerDo.toml is a commented valid config file
        And   no container was started

    Scenario: Handle Config File Template
        Given an empty project test-app
        When  container-do is called with `--init`
        And   container-do is called with `whoami`
        Then  the command exits with status 1

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

    Scenario: Regular Logging
        Given   config file for project test-app
            """
            [container]
            image = "ubuntu"
            mounts = []
            """
        And  environment variable CONTAINER_DO_LOGGING is set to "anything-but-debug"
        When container-do is called with `whoami`
        Then its error output is ""

    Scenario: Debug Logging
        Given config file for project test-app
            """
            [container]
            image = "ubuntu"
            mounts = []
            """
        And  environment variable CONTAINER_DO_LOGGING is set to "debug"
        When container-do is called with `whoami`
        Then its error output contains "DEBUG"

    @slow @timing
    Scenario: Cancel command
        Given config file for project test-app
            """
            [container]
            image = "ubuntu"
            mounts = []

            keep_alive = "3s"
            """
        When container-do is called with long-running `tail -f /dev/null`
        Then a container is started with name test-app-do
        When we wait for 1s
        Then a command matching /tail/ is running in the container
        When we send SIGINT to container-do
        And  wait for 2s
        Then the container is still running
        And  no command matching /tail/ is running in the container
        When we wait for another 2s
        Then the container is not running anymore
