@pending
Feature: Container Setup
    We want to be able to trigger certain setup steps from the config file
    so we do not have to manually repeat those steps whenever the container
    has to be recreated.

    Background:
        Given docker is installed
        And   config file for project test-app
            """
            [container]
            image = "ubuntu"
            """

    Scenario: Setup via Script File
        Given the config file also contains
            """
            [container.setup]
            script = "setup.sh"
            """
        And the project contains a file setup.sh with content
            """
            #!/bin/sh

            echo 'I was here!' > /setup_witness
            """
        And setup.sh is executable
        When container-do is called with `cat /setup_witness`
        Then the command output is "I was here!"

    Scenario: Setup via Command List
        Given the config file also contains
            """
            [container.setup]
            commands = [
                "touch /setup_witness",
                "echo -n 'I was ' > /setup_witness",
                "echo 'here!' >> /setup_witness"
            ]
            """
        When container-do is called with `cat /setup_witness`
        Then the command output is "I was here!"

    @pending
    Scenario Outline: Setup as user
        Given the config file also contains
            """
            [container.setup]
            user = "<user>"
            commands = [ "whoami > /tmp/setup_witness" ]
            """
        When container-do is called with `cat /tmp/setup_witness`
        Then the command output is "<user>"

        Examples:
            | user |
            | root |
            | tbd  |
