package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "strings"

    "go.uber.org/zap"
)

// TODO: better name?
// TODO: make configurable, maybe via ENV?
const doFile = "ContainerDo.toml"

var (
    Version string
    OsArch  string
    Build   string
)

const usageMessage = `container-do %s %s %s

Usage:  container-do --help
            Print this message

        container-do --init
            Create a template configuration file %s,
            unless that file already exists.

        container-do COMMAND [ARGUMENT...]
            Run the given command in a container as
            specified in %s.
`

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
    logger, err := makeLogger()
    if err != nil {
        handle(err)
    }
    //noinspection GoUnhandledErrorResult
    defer logger.Sync()
    undo := zap.ReplaceGlobals(logger)
    defer undo()

    if len(os.Args[1:]) < 1 {
        handle(UsageError{Message: "No command given"})
    }

    requestedContainerKill := false
    switch os.Args[1] {
        case "--help":
            fmt.Printf(usageMessage, Version, OsArch, Build, doFile, doFile)
            os.Exit(0)
        case "--init":
            if fileExists(doFile) {
                handle(UsageError{Message: fmt.Sprintf("Config file '%s' already exists.", doFile)})
            } else {
                handle(ioutil.WriteFile(doFile, []byte(strings.TrimSpace(ConfigFileTemplate)), 0o644))
                zap.L().Sugar().Infof("Created new %s from template.", doFile)
            }
            os.Exit(0)
        case "--kill":
            // Need to parse config before we can act on this!
            requestedContainerKill = true
    }

    config, err := parseConfig(doFile)
    handle(err)
    runner := makeRunner(config.Runner)

    containerExists, err := runner.DoesContainerExist(&config.Container)
    handle(err)

    if requestedContainerKill {
        if containerExists {
            err = runner.KillContainer(&config.Container)
            handle(err)
        }
        os.Exit(0)
    } else if !containerExists {
        // TODO: If we don't already have the image, we might block here for a while without input
        //       Pull explicitly and attach?
        err = runner.CreateContainer(&config.Container)
        handle(err)
        err = runner.RestartContainer(&config.Container)
        handle(err)
        err = runner.ExecutePredefined(&config.Container, config.ThingsToRun.Setup)
        handle(err)
    } else {
        containerRunning, err := runner.IsContainerRunning(&config.Container)
        handle(err)

        if !containerRunning {
            err = runner.RestartContainer(&config.Container)
            handle(err)
        }
    }

    err = runner.ExecutePredefined(&config.Container, config.ThingsToRun.Before)
    handle(err)
    err = runner.ExecuteCommand(&config.Container, os.Args[1:])
    handle(err) // TODO: Is aborting here what we want, usually?
    err = runner.ExecutePredefined(&config.Container, config.ThingsToRun.After)
    handle(err)
}
