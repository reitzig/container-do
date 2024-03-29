Feature: Container Creation
    We want to create a suitable container wrapping our tool suite.

    Background:
        Given docker is installed
        And   image ubuntu exists
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
        # this is not possible; see above TODO
        # And  the container has a volume mount for . at /
        And  the container has no volume mounts
        And  the container publishes no ports
        And  the command exits with status 0
        And  its output is "root"

    Scenario: Set Container Name
        Given the config file also contains
            """
            name = "test-app-foo"
            """
        When container-do is called with `whoami`
        Then a container is started with name test-app-foo

    Scenario: Set Working Directory
        Given the config file also contains
            """
            work_dir = "/foo"
            """
        When container-do is called with `whoami`
        Then a container is started with name test-app-do
        And  the container has working directory /foo

    Scenario: Disable Volume Mounts
        Given the config file also contains
            """
            mounts = []
            """
        When container-do is called with `whoami`
        Then a container is started with name test-app-do
        And  the container has no volume mounts

    Scenario: Set Up Volume Mounts
        Given temporary folders foo1, foo2
        And   the config file also contains
            """
            mounts = ["foo1:/foo", "foo2:/bar"]
            """
        When container-do is called with `whoami`
        Then a container is started with name test-app-do
        And  the container has a volume mount for foo1 at /foo
        And  the container has a volume mount for foo2 at /bar

    Scenario: Set Environment Variables
        Given environment variable SOME_VAR is set to "some value"
        And the config file also contains
            """
            [container.environment]
            FOO = "BAR"
            BAR = "FOO"
            VAR = "$SOME_VAR"
            NOP = "$OTHER_VAR"
            """
        When container-do is called with `whoami`
        Then a container is started with name test-app-do
        And  the container has an environment variable FOO with value "BAR"
        And  the container has an environment variable BAR with value "FOO"
        And  the container has an environment variable VAR with value "some value"
        And  the container has an environment variable NOP with value ""

    Scenario: Publish Ports
        Given the config file also contains
            """
            ports = ["8080:80", "4444"]
            """
        When container-do is called with `whoami`
        Then a container is started with name test-app-do
        Then the container publishes port 80 as 8080
        And  the container publishes port 4444
