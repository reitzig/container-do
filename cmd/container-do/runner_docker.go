package main

import (
    "os"
    "os/exec"
    "strings"

    "go.uber.org/zap"
)

type DockerRunner struct{}

func (d DockerRunner) runDockerCommand(commandAndArguments ...string) ([]byte, error) {
    zap.L().Sugar().Debugf(
        "Will run command: `%s %s`",
        d.RunnerExecutable(),
        strings.Join(commandAndArguments, " "))
    out, err := exec.Command(d.RunnerExecutable(), commandAndArguments...).CombinedOutput()

    if err != nil {
        switch err.(type) {
            case *exec.ExitError:
                exitCode := err.(*exec.ExitError).ExitCode()
                exitMsg := err.(*exec.ExitError).Error()
                zap.L().Sugar().Errorf("Command exit status %d (%s)", exitCode, exitMsg)
            default:
                zap.L().Sugar().Errorf("Error running command: %s", err.Error())
        }
    }
    if len(out) > 0 {
        zap.L().Sugar().Debugf("Command output: '%s'", string(out))
    }

    return out, err
}

// Run this before any command that needs `c.osFlavor` set.
func (d DockerRunner) DetermineOsFlavor(c *container) error {
    if c.osFlavor == "" {
        if c.RawOsFlavor == "" {
            out, err := d.runDockerCommand( "run", "--rm", c.Image, "cat", "/etc/os-release")
            if err != nil {
                return err
            }
            // TODO: distro-less containers will probably fail here; handle this?

            flavor, err := extractOsFlavorFromReleaseFile(out)
            if err != nil {
                return err
            }

            c.osFlavor = flavor
        } else {
            if !stringInSlice(c.RawOsFlavor, OsFlavors) {
                zap.L().Sugar().Warnf("Unsupported OS flavor '%s' in config -- fingers crossed!", c.RawOsFlavor)
            }
            // Trust the user:
            c.osFlavor = c.RawOsFlavor
        }
    }

    return nil
}

func (d DockerRunner) DoesContainerExist(c *container) (bool, error) {
    out, err := d.runDockerCommand( "ps", "--all", "--format", "{{.Names}}")
    if err != nil {
        return false, err
    }

    containers := strings.Split(string(out), "\n")
    return stringInSlice(c.Name, containers), nil
}

func (d DockerRunner) IsContainerRunning(c *container) (bool, error) {
    out, err := d.runDockerCommand( "inspect", "--format", "{{.State.Running}}", c.Name)
    if err != nil {
        return false, err
    }

    return strings.TrimSpace(string(out)) == "true", nil
}

func (d DockerRunner) CreateContainer(c *container) error {
    err := d.DetermineOsFlavor(c)
    if err != nil {
        return err
    }

    // Create container according to config
    args := []string{"create", "--name", c.Name}

    if !c.KeepStopped {
        args = append(args, "--rm")
    }

    // TODO: add mounts
    // TODO: add -w if necessary

    for key, value := range c.Environment {
        args = append(args, "-e", key+"="+value)
    }

    args = append(args, c.Image, "sh", "-c", containerRunScript(*c))

    _, err = d.runDockerCommand(args...)
    // TODO: avoid mishaps by storing container ID and checking for conflicts?
    //       --> if we do that, default container name can be dropped (let runner do its thing)
    if err != nil {
        return err
    }

    err = d.RestartContainer(c)
    if err != nil {
        return err
    }

    // Run requested setup
    // TODO: run setup, if any
    // TODO: add -w if necessary
    return err
}

func (d DockerRunner) RestartContainer(c *container) error {
    _, err := exec.Command(d.RunnerExecutable(), "start", c.Name).Output()
    return err
}

func (d DockerRunner) setKeepAliveToken(c *container, value string) error {
    _, err := d.runDockerCommand("exec", c.Name, "sh", "-c", setKeepAliveTokenScript(value))
    return err
}

func (d DockerRunner) ExecuteCommand(c *container, commandAndParameters []string) error {
    // Make sure container isn't killed while our command is running:
    _ = d.setKeepAliveToken(c, keepAliveIndefinitely)

    // TODO: add -w if necessary
    args := append([]string{"exec", "-i", c.Name}, commandAndParameters...)
    cmd := exec.Command(d.RunnerExecutable(), args...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    zap.L().Sugar().Debug("Will run command: `%s %s`", cmd.Path, strings.Join(cmd.Args, " "))
    cmdErr := cmd.Run()

    // Make container stay alive for another keep-alive interval
    _ = d.setKeepAliveToken(c, nextContainerStopTime(*c))

    return cmdErr
}
