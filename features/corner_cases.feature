Feature: Corner Cases
    Reproducing the odd issue

    Background:
        Given docker is installed

    # Control for issue #19
    Scenario: Custom work dir in Docker-built image but not in config
        Given config file for project test-app
            """
            [container]
            image = "container-do-test-workdir"
            """
        And Docker image container-do-test-workdir exists based on
            """
            FROM ubuntu
            WORKDIR /work
            """
        When container-do is called with `pwd`
        Then a container is started with name test-app-do
        And  the command exits with status 0
        And  its output is "/work"

    # Reproduce issue #19
    Scenario: Custom work dir in Kaniko-built image but not in config
        Given config file for project test-app
            """
            [container]
            image = "container-do-test-workdir-kaniko"
            """
        And Kaniko image container-do-test-workdir-kaniko exists based on
            """
            FROM ubuntu
            WORKDIR /work
            """
        When container-do is called with `pwd`
        Then a container is started with name test-app-do
        And  the command exits with status 0
        And  its output is "/work"

    # Reproduce issue #21
    Scenario: Image missing
        Given config file for project test-app
            """
            [container]
            image = "funky-foo"
            """
        When container-do is called with `pwd`
        Then no container was started
        And the command exits with status 77
        And its error output contains "Unable to find image"
        And its error output contains "funky-foo"
