Feature: Container Creation
    We want to create a suitable container wrapping our tool suite.

    Background:
        Given docker is installed
        And   config file for project test-app
            """
            [container]
            image = "ubuntu"
            """
        # TODO: find a small image with non-root USER, WORKDIR

    Scenario: Default Behaviour
        When container-do is called with `whoami`
        Then a container is started with name test-app-do
        And  the container is based on image ubuntu
        And  the container has working directory /
        And  the container has a volume mount for . at /
        And  the command exits with status 0
        And  its output is "root"

    Scenario: Set Container Name
        Given the config file also contains
            """
            name = "test-app-foo"
            """
        When container-do is called with `whoami`
        Then a container is started with name test-app-foo

    @pending
    Scenario: Set Working Directory
        Given the config file also contains
            """
            work_dir = "/foo"
            """
        When container-do is called with `whoami`
        Then a container is started with working directory /foo

    @pending
    Scenario: Set Up Volume Mounts
        Given temporary folders foo1, foo2
        And   the config file also contains
            """
            mounts = ["foo1:/foo", "foo2:/bar"]
            """
        When container-do is called with `whoami`
        Then the container has a volume mount for foo1 at /foo
        And  the container has a volume mount for foo2 at /bar

    Scenario: Set Environment Variables
        Given the config file also contains
            """
            [container.environment]
            FOO = "BAR"
            BAR = "FOO"
            """
        When container-do is called with `whoami`
        Then the container has an environment variable FOO with value "BAR"
        And  the container has an environment variable BAR with value "FOO"
