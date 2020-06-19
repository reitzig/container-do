package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
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

    if c.WorkDir != "" {
        args = append(args, "--workdir", c.WorkDir)
    }

    if c.Mounts == nil {
        hostWorkDir, err := os.Getwd()
        if err != nil {
            return err
        }
        hostWorkDir, err = filepath.EvalSymlinks(hostWorkDir)
        if err != nil {
            return err
        }

        containerWorkDir := c.WorkDir
        if c.WorkDir == "" {
            out, err := d.runDockerCommand("inspect", "--format={{.ContainerConfig.WorkingDir}}", c.Image)
            if err != nil {
                return err
            }
            if wd := strings.TrimSpace(string(out)); wd == "" {
                containerWorkDir = "/"
            } else {
                containerWorkDir = wd
            }
        }

        if containerWorkDir == "/" {
            // Forbidden by Docker!
            zap.L().Sugar().Warn("Can not set up default bind-mount to container root; " +
                "specify other working directory or declare valid mounts explicitly!")
        } else {
            bindMount := fmt.Sprintf("%s:%s", hostWorkDir, containerWorkDir)
            zap.L().Sugar().Debug("Using default bind-mount '%s'", bindMount)
            args = append(args, "--volume", bindMount)
        }
    } else if len(c.Mounts) == 0 {
        zap.L().Debug("Volume bind-mounts disabled by user.")
    } else {
        for _, bind := range c.Mounts {
            bind, err = expandHostPath(bind)
            if err != nil {
                return err
            }
            args = append(args, "--volume", bind)
        }
    }

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
    // TODO: run setup, if any; add -w if necessary
    return err
}

func (d DockerRunner) RestartContainer(c *container) error {
    _, err := d.runDockerCommand("start", c.Name)
    return err
}

func (d DockerRunner) setKeepAliveToken(c *container, value string) error {
    _, err := d.runDockerCommand("exec", c.Name, "sh", "-c", setKeepAliveTokenScript(value))
    return err
}

func (d DockerRunner) ExecutePredefined(c *container, thing thingToRun) error {
    _ = d.setKeepAliveToken(c, keepAliveIndefinitely)
    var err error = nil

    execCmd := []string{"exec"}
    if thing.User != "" {
        execCmd = append(execCmd, "--user", thing.User)
    }
    execCmd = append(execCmd, c.Name)

    if thing.ScriptFile != "" {
        _, err = d.runDockerCommand(append(execCmd, thing.ScriptFile)...)
    }
    for _, cmd := range thing.Commands {
        if err != nil {
            break
        }

        _, err = d.runDockerCommand(append(execCmd, "sh", "-c", cmd)...)
    }

    _ = d.setKeepAliveToken(c, nextContainerStopTime(*c))
    return err
}

func (d DockerRunner) ExecuteCommand(c *container, commandAndParameters []string) error {
    // Make sure container isn't killed while our command is running:
    _ = d.setKeepAliveToken(c, keepAliveIndefinitely)

    args := append([]string{"exec", "-i", "-w", c.WorkDir, c.Name}, commandAndParameters...)
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
