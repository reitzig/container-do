package main

import (
    "fmt"
    "os"
    "os/exec"
    "go.uber.org/zap"
)

// TODO: better name?
// TODO: make configurable, maybe via ENV?
const doFile = "ContainerDo.toml"

func handle(err error) {
    if err != nil {
        switch err.(type) {
        case UsageError, ConfigError:
            _, _ = os.Stderr.WriteString(fmt.Sprintln(err.Error()))
            os.Exit(1)
        case *exec.ExitError:
            _, _ = os.Stderr.WriteString(fmt.Sprintln(err.Error()))
            os.Exit(err.(*exec.ExitError).ExitCode())
        default:
            panic(err)
        }
    }
}

// TODO: replace with CLI parser and its error
type UsageError struct {
    Message string
}

func (e UsageError) Error() string {
    return fmt.Sprintf("Wrong usage: %s", e.Message)
}

func main() {
    // TODO: capture --help, --init
    // TODO: add command to stop/kill/"purge", i.e. stop/kill/remote container?

    // TODO: Add CLI or ENV flag to turn on debug logging?
    logger, err := makeLogger()
    if err != nil {
        handle(err)
    }
    defer logger.Sync()
    undo := zap.ReplaceGlobals(logger)
    defer undo()

    if len(os.Args[1:]) < 1 {
        handle(UsageError{Message: "No command given"})
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
