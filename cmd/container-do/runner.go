package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type runnerExec interface {
	DoesContainerExist(c container) (bool, error)
	IsContainerRunning(c container) (bool, error)
	CreateContainer(c container) error
	RestartContainer(c container) error
	ExecuteCommand(c container, commandAndParameters []string) error
}

func nextContainerKillTime(c container) string {
	return fmt.Sprintf("%d", time.Now().Add(c.KeepAlive()).Unix())
}

const keepAliveFile = "/keepalive"

func containerRunScript(c container) string {
	keepAliveString := fmt.Sprintf("%.0f", c.KeepAlive().Seconds())
	// NB: Can not hard-code the first stop-time because otherwise RestartContainer wouldn't work!
	return "date -d \"$(date '+%F %T') " + keepAliveString + " seconds\" +%s > " + keepAliveFile + "; " +
		"while [[ $(cat " + keepAliveFile + ") > $(date +%s) ]]; do sleep 1; done"
}

func setKeepAliveTokenScript(value string) string {
	return "echo '" + value + "' > " + keepAliveFile
}

type DockerRunner struct{}

func (d DockerRunner) DoesContainerExist(c container) (bool, error) {
	out, err := exec.Command("docker", "ps", "--all", "--format", "{{.Names}}").Output()
	if err != nil {
		return false, err
	}

	containers := strings.Split(string(out), "\n")
	return stringInSlice(c.Name, containers), nil
}

func (d DockerRunner) IsContainerRunning(c container) (bool, error) {
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.Running}}", c.Name).Output()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(out)) == "true", nil
}

func (d DockerRunner) CreateContainer(c container) error {
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

	args = append(args, c.Image, "bash", "-c", containerRunScript(c))

	_, err := exec.Command("docker", args...).Output()
	// TODO: avoid mishaps by storing container ID and checking for conflicts?
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

func (d DockerRunner) RestartContainer(c container) error {
	_, err := exec.Command("docker", "start", c.Name).Output()
	return err
}

func (d DockerRunner) setKeepAliveToken(c container, value string) error {
	out, err := exec.Command(
		"docker", "exec", c.Name, "bash", "-c", setKeepAliveTokenScript(value)).Output()

	if err != nil {
		// TODO proper logging
		_, _ = os.Stderr.WriteString(fmt.Sprintln(err, ":", string(out)))
	}

	return err
}

func (d DockerRunner) ExecuteCommand(c container, commandAndParameters []string) error {
	// Make sure container isn't killed while our command is running:
	_ = d.setKeepAliveToken(c, "running") // "running" > "159..."

	// TODO: add -w if necessary
	args := append([]string{"exec", "-it", c.Name}, commandAndParameters...)
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdErr := cmd.Run()

	// Make container stay alive for another keep-alive interval
	_ = d.setKeepAliveToken(c, nextContainerKillTime(c))

	return cmdErr
}

func makeRunner(r runner) runnerExec {
	switch r {
	case docker:
		// TODO: check if docker exists
		return DockerRunner{}
	case podman:
		panic("podman runner not yet implemented")
	default:
		panic("Invalid container runner: " + r)
	}
}
