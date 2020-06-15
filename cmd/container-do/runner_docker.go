package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type DockerRunner struct{}

// Run this before any command that needs `c.osFlavor` set.
func (d DockerRunner) DetermineOsFlavor(c *container) error {
	if c.osFlavor == "" {
		if c.RawOsFlavor == "" {
			out, err := exec.Command(d.RunnerExecutable(), "run", "--rm", c.Image, "cat", "/etc/os-release").Output()
			if err != nil {
				return err
			}
			// TODO: distro-less containers will probably fail here; handle this?
			// TODO: debug-log output

			flavor, err := extractOsFlavorFromReleaseFile(out)
			if err != nil {
				return err
			}

			c.osFlavor = flavor
		} else {
			if !stringInSlice(c.RawOsFlavor, OsFlavors) {
				// TODO: log warning
			}
			// Trust the user:
			c.osFlavor = c.RawOsFlavor
		}
	}

	return nil
}

func (d DockerRunner) DoesContainerExist(c *container) (bool, error) {
	out, err := exec.Command(d.RunnerExecutable(), "ps", "--all", "--format", "{{.Names}}").Output()
	if err != nil {
		return false, err
	}

	containers := strings.Split(string(out), "\n")
	return stringInSlice(c.Name, containers), nil
}

func (d DockerRunner) IsContainerRunning(c *container) (bool, error) {
	out, err := exec.Command(d.RunnerExecutable(), "inspect", "--format", "{{.State.Running}}", c.Name).Output()
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

	// TODO: debug-log comand
	_, err = exec.Command(d.RunnerExecutable(), args...).Output()
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
	out, err := exec.Command(d.RunnerExecutable(),
		"exec", c.Name, "sh", "-c", setKeepAliveTokenScript(value)).Output()

	if err != nil {
		// TODO proper logging
		_, _ = os.Stderr.WriteString(fmt.Sprintln(err, ":", string(out)))
	}

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
	cmdErr := cmd.Run()

	// Make container stay alive for another keep-alive interval
	_ = d.setKeepAliveToken(c, nextContainerStopTime(*c))

	return cmdErr
}
