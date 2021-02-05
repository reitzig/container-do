@slow
Feature: Container Setup & Command pre-/post-processing
    We want to be able to trigger commands from the config file so we do not
    have to manually repeat those steps whenever the container has to be recreated,
    or before/after we run some command.

    Background:
        Given docker is installed
        And   config file for project test-app
            """
            [container]
            image = "ubuntu"
            work_dir = "/scripts"
            """
            # Leaving bind-mount implicit -- the default should work!

    Scenario: Setup via Script File
        Given the config file also contains
            """
            [run.setup]
            script_file = "./setup.sh"
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
            [run.setup]
            commands = [
                "touch /setup_witness",
                "echo -n 'I was ' > /setup_witness",
                "echo 'here!' >> /setup_witness"
            ]
            """
        When container-do is called with `cat /setup_witness`
        Then the command output is "I was here!"

    Scenario: Setup without attach
        Given the config file also contains
            """
            [run.setup]
            attach = false
            commands = [ "echo 'Do you see me?'" ]
            """
        When container-do is called with `whoami`
        Then its output is "root"

    Scenario: Setup with attach
        Given the config file also contains
            """
            [run.setup]
            attach = true
            commands = [ "echo 'Do you see me?'" ]
            """
        When container-do is called with `whoami`
        Then its output contains "Do you see me?"

    Scenario: Setup fails with attach
        Given the config file also contains
            """
            [run.setup]
            attach = true
            commands = [ "echo 'Do you see me?'; exit 77" ]

            [run.before]
            attach = true
            commands = [ "echo 'Or you don't!'" ]
            """
        When container-do is called with `whoami`
        Then its output is "Do you see me?"
        And the command exits with status 77
        And no container was started

    Scenario: Setup fails without attach
        Given the config file also contains
            """
            [run.setup]
            attach = false
            commands = [ "echo 'Do you see me?'; exit 77" ]

            [run.before]
            attach = true
            commands = [ "echo 'Or you don't!'" ]
            """
        When container-do is called with `whoami`
        Then its output is ""
        And the command exits with status 77
        And no container was started

    # TODO: Test container doesn't have a non-root user yet.
    @pending
    Scenario Outline: Setup as user
        Given the config file also contains
            """
            [run.setup]
            user = "<user>"
            commands = [ "whoami > /tmp/setup_witness" ]
            """
        When container-do is called with `cat /tmp/setup_witness`
        Then the command output is "<user>"

        Examples:
            | user |
            | root |
            | tbd  |

    Scenario: Before via Script File
        Given the config file also contains
            """
            [run.before]
            script_file = "./before.sh"
            """
        And the project contains a file before.sh with content
            """
            #!/bin/sh

            echo 'I was before!' > /before_witness
            """
        And before.sh is executable
        When container-do is called with `cat /before_witness`
        Then the command output is "I was before!"

    Scenario: Before via Command List
        Given the config file also contains
            """
            [run.before]
            commands = [
                "touch /before_witness",
                "echo -n 'I was ' > /before_witness",
                "echo 'before!' >> /before_witness"
            ]
            """
        When container-do is called with `cat /before_witness`
        Then the command output is "I was before!"

    Scenario: After via Script File
        Given the config file also contains
            """
            [run.setup]
            commands = [ "touch /after_witness" ]

            [run.after]
            script_file = "./after.sh"
            """
        And the project contains a file after.sh with content
            """
            #!/bin/sh

            echo 'I was after!' > /after_witness
            """
        And after.sh is executable
        When container-do is called with `cat /after_witness`
        Then the command output is ""
        When container-do is called with `cat /after_witness`
        Then the command output is "I was after!"

    Scenario: After via Command List
        Given the config file also contains
            """
            [run.setup]
            commands = [ "touch /after_witness" ]

            [run.after]
            commands = [
                "echo -n 'I was ' > /after_witness",
                "echo 'after!' >> /after_witness"
            ]
            """
        When container-do is called with `cat /after_witness`
        Then the command output is ""
        When container-do is called with `cat /after_witness`
        Then the command output is "I was after!"

    Scenario: Run all the Things
        Given the config file also contains
            """
            [run.setup]
            commands = [
                "echo '0,0,0' > /witnesses"
            ]

            [run.before]
            commands = [
                "new=$(awk -F, '{$2=$2+1}1' OFS=, /witnesses); echo $new > /witnesses"
            ]

            [run.after]
            commands = [
                "new=$(awk -F, '{$3=$3+1}1' OFS=, /witnesses); echo $new > /witnesses"
            ]
            """
        When container-do is called with `cat /witnesses`
        Then the command output is "0,1,0"
        When container-do is called with `cat /witnesses`
        Then the command output is "0,2,1"
        When container-do is called with `cat /witnesses`
        Then the command output is "0,3,2"

    Scenario: Copy files during setup
        Given the config file also contains
            """
            mounts = []

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

    Scenario: Copy files before command
        Given the config file also contains
            """
            mounts = []

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
            mounts = []

            [run.setup]
            commands = [
                "mkdir -p outputs",
                "echo 'Hidden but there.' > outputs/more"
            ]

            [run.after]
            commands = [
                "echo 'And this, too!' >> output"
            ]

            [[copy.after]]
            files = ["output"]
            to = "main_output"

            [[copy.after]]
            files = ["outputs/*"]
            to = "more_outputs"
            """
        When container-do is called with `echo 'We see this?' > output`
        Then file main_output now contains
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
            mounts = []

            [[copy.setup]]
            files = ["does_not_exist"]
            """
        When container-do is called with `whoami`
        Then the command exits with status 0
        # TODO: This has potential for silent errors, but might also be useful ("copy if exists") -- trade off!

    Scenario: Copy missing file from container
        Given the config file also contains
            """
            mounts = []

            [[copy.after]]
            files = ["does_not_exist"]
            """
        When container-do is called with `whoami`
        Then the command exits with status 0
