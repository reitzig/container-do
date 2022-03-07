package main

import (
    "fmt"
    "os"
    "os/exec"
    "os/signal"
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
            zap.L().Sugar().Debugf("Docker command quit with status %d", exitCode)
        default:
            zap.L().Sugar().Debugf("Error running command: %s", err.Error())
        }
    }
    if len(out) > 0 {
        zap.L().Sugar().Debugf("Command output: '%s'", strings.TrimSpace(string(out)))
    }

    return out, err
}

func (d DockerRunner) runDockerCommandAttached(commandAndArguments ...string) error {
    cmd := exec.Command(d.RunnerExecutable(), commandAndArguments...)
    zap.L().Sugar().Debugf("Will run command: `%s`", strings.Join(cmd.Args, " "))

    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err := cmd.Start()
    if err != nil {
        return err
    }

    // Forward signals to `docker`:
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, os.Kill)
    go func() {
        for sig := range c {
            zap.L().Sugar().Debugf("Forwarding signal '%v' to `%s` (%d).",
                sig, d.RunnerExecutable(), cmd.Process.Pid)
            forwardErr := cmd.Process.Signal(sig)
            if forwardErr != nil {
                zap.L().Sugar().Warnf("Could not forward signal '%v' to `%s` (%d): %s",
                    sig, d.RunnerExecutable(), cmd.Process.Pid, forwardErr.Error())
            }
        }
    }()

    err = cmd.Wait()
    signal.Stop(c)

    return err
}

// Run this before any command that needs `c.osFlavor` set.
func (d DockerRunner) DetermineOsFlavor(c *container) error {
    if c.osFlavor == "" {
        if c.RawOsFlavor == "" {
            out, err := d.runDockerCommand("run", "--rm", c.Image, "cat", "/etc/os-release")
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

// Run this before any command that needs `c.WorkDir` set.
func (d DockerRunner) DetermineWorkDir(c *container) error {
    if c.WorkDir == "" {
        out, err := d.runDockerCommand("inspect", "--format={{.Config.WorkingDir}}", c.Image)
        if err != nil {
            return err
        }
        if wd := strings.TrimSpace(string(out)); wd == "" {
            c.WorkDir = "/"
        } else {
            c.WorkDir = wd
        }
    }

    return nil
}

func (d DockerRunner) DoesImageExist(c *container) (bool, error) {
    out, err := d.runDockerCommand("inspect", "--type", "image", "--format", "{{.Id}}", c.Image)

    if err != nil {
        if strings.Contains(string(out), "No such image:") {
            return false, nil
        } else {
            return false, err
        }
    } else {
        return true, nil
    }
}

func (d DockerRunner) DoesContainerExist(c *container) (bool, error) {
    out, err := d.runDockerCommand("ps", "--all", "--format", "{{.Names}}")
    if err != nil {
        return false, err
    }

    containers := strings.Split(string(out), "\n")
    return stringInSlice(c.Name, containers), nil
}

func (d DockerRunner) IsContainerRunning(c *container) (bool, error) {
    out, err := d.runDockerCommand("inspect", "--format", "{{.State.Running}}", c.Name)
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
    args := []string{"create",
        "--name", c.Name,
        "--entrypoint", "sh", // override whatever shenanigans the image author came up with
    }

    if !c.KeepStopped {
        args = append(args, "--rm")
    }

    if c.WorkDir != "" {
        args = append(args, "--workdir", c.WorkDir)
    } else {
        err := d.DetermineWorkDir(c)
        if err != nil {
            return err
        }
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

        if c.WorkDir == "/" {
            // Forbidden by Docker!
            zap.L().Sugar().Warn("Can not set up default bind-mount to container root; " +
                "specify other working directory or declare valid mounts explicitly!")
        } else {
            bindMount := fmt.Sprintf("%s:%s", hostWorkDir, c.WorkDir)
            zap.L().Sugar().Debugf("Using default bind-mount '%s'", bindMount)
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

    for _, mapping := range c.Ports {
        args = append(args, "--publish", mapping)
    }

    for key, value := range c.Environment {
        if strings.HasPrefix(value, "$") {
            if hostValue, isSet := os.LookupEnv(strings.TrimPrefix(value, "$")); isSet {
                value = hostValue
            } else {
                zap.L().Sugar().Warnf("Environment variable '%s' not set; leaving %s empty.", value, key)
                value = ""
            }
        }

        args = append(args, "-e", key+"="+value)
    }

    // NB: '-c' is for 'sh' entrypoint, specified above
    args = append(args, c.Image, "-c", containerRunScript(*c))

    _, err = d.runDockerCommand(args...)
    // TODO: avoid mishaps by storing container ID and checking for conflicts?
    //       --> if we do that, default container name can be dropped (let runner do its thing)

    return err
}

func (d DockerRunner) RestartContainer(c *container) error {
    _, err := d.runDockerCommand("start", c.Name)
    return err
}

func (d DockerRunner) KillContainer(c *container) error {
    _, err := d.runDockerCommand("kill", c.Name)
    return err
}

func (d DockerRunner) setKeepAliveToken(c *container, value string) error {
    _, err := d.runDockerCommand("exec", c.Name, "sh", "-c", setKeepAliveTokenScript(value))
    return err
}

func (d DockerRunner) ExecutePredefined(c *container, thing thingToRun) error {
    if !thing.willRunSomething() {
        return nil
    }

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

        dockerCmd := append(execCmd, "sh", "-c", cmd)
        if thing.Attach {
            err = d.runDockerCommandAttached(dockerCmd...)
        } else {
            _, err = d.runDockerCommand(dockerCmd...)
        }
    }

    _ = d.setKeepAliveToken(c, nextContainerStopTime(*c))
    return err
}

func (d DockerRunner) CopyFilesTo(c *container, thing []thingToCopy) error {
    for _, filesAndTarget := range thing {
        zap.L().Sugar().Debugf("Will copy '%s' to '%s'", filesAndTarget.Files, filesAndTarget.Target)
        // `docker cp` doesn't expand globs, so we have to do it ourselves.
        files := []string{}
        for _, fileGlob := range filesAndTarget.Files {
            filesOfGlob, err := filepath.Glob(fileGlob)
            if err != nil {
                return err
            }
            files = append(files, filesOfGlob...)
        }
        zap.L().Sugar().Debugf("Files to copy: %s", files)

        if len(files) == 0 {
            continue
        }

        var target string
        if strings.HasPrefix(filesAndTarget.Target, "/") {
            // Absolute path given
            target = filesAndTarget.Target
        } else {
            // Relative path given; add working directory to get absolute path
            err := d.DetermineWorkDir(c)
            if err != nil {
                return err
            }
            target = fmt.Sprintf("%s/%s", c.WorkDir, filesAndTarget.Target)
        }

        if len(files) > 1 || strings.HasSuffix(target, "/") {
            // Copying files into a directory --> make sure it exists!
            zap.L().Sugar().Debugf("Creating target directory: %s", target)
            err := d.runDockerCommandAttached("exec", c.Name, "mkdir", "-p", target)
            if err != nil {
                return err
            }
        } else {
            // Copying a single file or directory --> make sure its parent exists!
            parentDir := filepath.Dir(target)
            zap.L().Sugar().Debugf("Creating target directory: %s", parentDir)
            err := d.runDockerCommandAttached("exec", c.Name, "mkdir", "-p", parentDir)
            if err != nil {
                return err
            }
        }

        for _, source := range files {
            err := d.runDockerCommandAttached("cp", source, fmt.Sprintf("%s:%s", c.Name, target))
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func (d DockerRunner) CopyFilesFrom(c *container, thing []thingToCopy) error {
    for _, filesAndTarget := range thing {
        err := d.DetermineWorkDir(c)
        if err != nil {
            return err
        }

        zap.L().Sugar().Debugf("Will copy '%s' to '%s'", filesAndTarget.Files, filesAndTarget.Target)
        // `docker cp` doesn't expand globs, so we have to do it ourselves.
        // We want the result to be equivalent to filepath.Glob above, so drop missing files as well.
        // The for loop in sh doesn't expand globs if the given values are quoted, so we have to escape first.
        escapedFileGlobs := ""
        for _, fileGlob := range filesAndTarget.Files {
            escapedFileGlobs += " " + strings.ReplaceAll(fileGlob, " ", "\\ ")
        }
        globExpandingLoop := fmt.Sprintf("for f in %s; do if [ -e \"$f\" ]; then echo $f; fi done", escapedFileGlobs)
        output, err := d.runDockerCommand("exec", "-w", c.WorkDir, c.Name, "sh", "-c", globExpandingLoop)
        if err != nil {
            return err
        }

        files := filter(
            strings.Split(strings.TrimSpace(string(output)), "\n"),
            func(s string) bool { return strings.TrimSpace(s) != "" },
        )
        zap.L().Sugar().Debugf("Files to copy: %s", files)

        if len(files) == 0 {
            continue
        } else if len(files) > 1 || strings.HasSuffix(filesAndTarget.Target, "/") {
            // Copying files into a directory --> make sure it exists!
            zap.L().Sugar().Debugf("Creating target directory: %s", filesAndTarget.Target)
            err := os.MkdirAll(filesAndTarget.Target, os.ModeDir|os.FileMode(0775))
            if err != nil {
                return err
            }
        } else {
            // Copying a single file or directory --> make sure its parent exists!
            parentDir := filepath.Dir(filesAndTarget.Target)
            zap.L().Sugar().Debugf("Creating target directory: %s", parentDir)
            err := os.MkdirAll(parentDir, os.ModeDir|os.FileMode(0775))
            if err != nil {
                return err
            }
        }

        for _, source := range files {
            if strings.HasPrefix(source, "/") {
                // Absolute path given
                source = fmt.Sprintf("%s:%s", c.Name, source)
            } else {
                // Relative path given; add working directory to get absolute path
                source = fmt.Sprintf("%s:%s/%s", c.Name, c.WorkDir, source)
            }
            err := d.runDockerCommandAttached("cp", source, filesAndTarget.Target)
            if err != nil {
                return err
            }
        }
    }

    return nil
}

func (d DockerRunner) ExecuteCommand(c *container, commandAndParameters []string) error {
    // Make sure container isn't killed while our command is running:
    _ = d.setKeepAliveToken(c, keepAliveIndefinitely)

    // NB: If c.WorkDir is empty, docker will accept -w without value and use the container working directory -- correct!
    args := append([]string{"exec", "-i", "-w", c.WorkDir, c.Name}, commandAndParameters...)
    cmdErr := d.runDockerCommandAttached(args...)

    // Make container stay alive for another keep-alive interval
    _ = d.setKeepAliveToken(c, nextContainerStopTime(*c))

    return cmdErr
}
