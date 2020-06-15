package main

import (
    "fmt"
    "os"
    "os/exec"
)

// TODO: better name?
// TODO: make configurable, maybe via ENV?
const doFile = "ContainerDo.toml"

func handle(err error) {
    if err != nil {
        switch err.(type) {
        case ConfigError:
            _, _ = os.Stderr.WriteString(fmt.Sprintln(err.Error()))
            os.Exit(2)
        case *exec.ExitError:
            os.Exit(err.(*exec.ExitError).ExitCode())
        default:
            panic(err)
        }
    }
}

func main() {
    // TODO: capture --help, --init
    // TODO: add command to stop/kill/"purge", i.e. stop/kill/remote container?

    if len(os.Args[1:]) < 1 {
        _, _ = os.Stderr.WriteString(fmt.Sprintln("Error: No command given"))
        os.Exit(1)
    }

    config, err := parseConfig(doFile)
    handle(err)
    runner := makeRunner(config.Runner)

    containerExists, err := runner.DoesContainerExist(&config.Container)
    handle(err)

    if !containerExists {
        err = runner.CreateContainer(&config.Container)
        handle(err)
    }

    containerRunning, err := runner.IsContainerRunning(&config.Container)
    handle(err)

    if !containerRunning {
        err = runner.RestartContainer(&config.Container)
        handle(err)
    }

    err = runner.ExecuteCommand(&config.Container, os.Args[1:])
    handle(err)
}
