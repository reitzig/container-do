@slow
Feature: Support Many Linux Flavors
    The tool should have as few dependencies on the container image as possible and
    thus work with a range of base distributions.

    Scenario Outline:
        Given docker is installed
        And   image <image> exists
        And   config file for project test-app
            """
            [container]
            image = "<image>"
            """
        When container-do is called with `cat /etc/os-release`
        Then a container is started with name test-app-do
        And  the container is based on image <image>
        And  the command exits with status 0
        And  its output contains "NAME=\"<distribution>\""

        Examples:
            | image                                       | distribution             |
            | centos                                      | CentOS Linux             |
            | debian:buster-slim                          | Debian GNU/Linux         |
            | registry.access.redhat.com/ubi8/ubi-minimal | Red Hat Enterprise Linux |
            | ubuntu                                      | Ubuntu                   |
            | alpine                                      | Alpine Linux             |
            | oraclelinux:7-slim                          | Oracle Linux Server      |
