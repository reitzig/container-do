@slow @timing
Feature: Keep Container Alive
    We want to keep the container alive to avoid unnecessary restarts,
    but also for it to vanish if we haven't needed it in a while.

    Background:
        Given docker is installed
        And   image ubuntu exists
        And   config file for project test-app
            """
            [container]
            image = "ubuntu"
            keep_alive = "3s"

            # Override defaults that would impact timing:
            mounts = []
            """

    Scenario: Keep Container Alive for Configured Interval
        When container-do is called with `whoami`
        Then a container is started with name test-app-do
        When we wait for 2s
        Then the container is still running
        When we wait for another 2s
        Then the container is not running anymore
        And  the container is gone

    Scenario: Remove Container After Exit
        Given the config file also contains
            """
            keep_stopped = false
            """
        When container-do is called with `whoami`
        Then a container is started with name test-app-do
        When we wait for 4s
        Then the container is not running anymore
        And  the container is gone

    Scenario: Keep Container After Exit
        Given the config file also contains
            """
            keep_stopped = true
            """
        When container-do is called with `whoami`
        Then a container is started with name test-app-do
        When we wait for 4s
        Then the container is not running anymore
        And  the container is still there

    Scenario: Reset interval when commands are run
        When container-do is called with `whoami`
        Then a container is started with name test-app-do
        When we wait for 2s
        And  container-do is called with `whoami`
        And  we wait for another 1.5s
        Then the container is still running
        When we wait for another 3s
        Then the container is not running anymore

    Scenario: Reuse kept container
        Given the config file also contains
            """
            keep_stopped = true
            """
        When container-do is called with `touch witness`
        Then a container is started with name test-app-do
        When we wait for 4s
        Then the container is not running anymore
        When container-do is called with `cat witness`
        Then the command exits with status 0
