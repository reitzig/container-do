Feature: Copy file to and from container
    We want to be able to selectively inject and extract files during setup,
    before each command, and after each command.

    Background:
        Given docker is installed
        And Docker image container-do-test-copy-workdir exists based on
            """
            FROM ubuntu
            WORKDIR /work
            """
        And   config file for project test-app
            """
            [container]
            image = "container-do-test-copy-workdir"
            work_dir = "/work"
            mounts = []
            """
        And the project contains a file spam_a with content
            """
            A
            """
        And the project contains a file spam_b with content
            """
            B
            """
        And the project contains a file ham with content
            """
            C
            """

    Scenario: Copy files during setup
        Given the config file also contains
            """
            [[copy.setup]]
            files = ["spam_*"]
            to = "spam"

            [[copy.setup]]
            files = ["ham"]

            [run.setup]
            commands = [
                "cd spam; for f in *; do mv ${f} ${f#spam_}; done"
            ]
            """
        When container-do is called with `ls -1`
        Then the command output is
            """
            ham
            spam
            """
        When container-do is called with `ls -1 spam`
        Then the command output is
            """
            a
            b
            """
        When container-do is called with `cat ham`
        Then the command output is "C"

    Scenario: Copy files during setup fails
        Given the config file also contains
            """
            [[copy.setup]]
            files = ["spam_a"]
            to = "spam"

            [[copy.setup]]
            files = ["spam_a"]
            to = "spam/dontyoudare"
            """
        When container-do is called with `cat spam`
        Then its output is ""
        And the command exits with status 1
        And no container was started

    Scenario: Copy files before command
        Given the config file also contains
            """
            [run.setup]
            attach = true
            commands = [
                "ls"
            ]

            [[copy.before]]
            files = ["spam_*"]
            to = "spam"

            [[copy.before]]
            files = ["ham"]

            [run.before]
            commands = [
                "cd spam; for f in *; do mv ${f} ${f#spam_}; done"
            ]
            """
        When container-do is called with `ls -1`
        Then the command output is
            """
            ham
            spam
            """
        When container-do is called with `ls -1 spam`
        Then the command output is
            """
            a
            b
            """
        When container-do is called with `cat ham`
        Then the command output is "C"

    Scenario: Copy files after command
        Given the config file also contains
            """
            [run.setup]
            commands = [
                "mkdir -p outputs",
                "echo 'Hidden but there.' > outputs/more"
            ]

            [run.before]
            commands = [
                "echo 'We see this?' > echoed"
            ]

            [run.after]
            commands = [
                "echo 'And this, too!' >> echoed"
            ]

            [[copy.after]]
            files = ["echoed"]
            to = "main_output"

            [[copy.after]]
            files = ["outputs/*"]
            to = "more_outputs/"
            """
        When container-do is called with `cat echoed`
        Then the command output is "We see this?"
        And file main_output now contains
            """
            We see this?
            And this, too!
            """
        And file more_outputs/more now contains
            """
            Hidden but there.
            """

    Scenario: Copy missing file into container
        Given the config file also contains
            """
            [[copy.setup]]
            files = ["does_not_exist"]

            [[copy.setup]]
            files = ["spam_a"]
            """
        When container-do is called with `cat spam_a`
        Then the command exits with status 0
            # TODO: This has potential for silent errors, but might also be useful ("copy if exists") -- trade off!
        And its output is "A"

    Scenario: Copy missing file from container
        Given the config file also contains
            """
            [run.after]
            commands = ["echo 'Was here!' > echoed"]

            [[copy.after]]
            files = ["does_not_exist"]

            [[copy.after]]
            files = ["echoed"]
            """
        When container-do is called with `whoami`
        Then the command exits with status 0
            # TODO: This has potential for silent errors, but might also be useful ("copy if exists") -- trade off!
        And file echoed now contains
            """
            Was here!
            """

    Scenario: Copy file to nested new location in container
        Given the config file also contains
            """
            [[copy.setup]]
            files = ["spam_a"]
            to = "quite/a/long/path/spam"
            """
        When container-do is called with `cat quite/a/long/path/spam`
        Then the command exits with status 0
        And  its output is "A"

    Scenario: Copy file to nested new location on host
        Given the config file also contains
            """
            [run.before]
            commands = ["echo 'Was there!' > echoed"]

            [[copy.after]]
            files = ["echoed"]
            to = "quite/a/long/path/echoed"
            """
        When container-do is called with `ls -1`
        Then the command exits with status 0
        And  its output is "echoed"
        And file quite/a/long/path/echoed now contains
            """
            Was there!
            """

    Scenario: Copy files to relative path in container but non-root work_dir is not configured
        Given   config file for project test-app
            """
            [container]
            image = "container-do-test-copy-workdir"
            mounts = []

            [[copy.setup]]
            files = ["spam_a"]
            to = "some/dir/"
            """
        When container-do is called with `cat some/dir/spam_a`
        Then the command exits with status 0
        And  its output is "A"
