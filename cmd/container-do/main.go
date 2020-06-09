package main

import (
	"os"
	"os/exec"
)

const doFile = "ContainerDo.toml" // TODO: better name

func main() {
	// TODO: capture --help, --init

	if len(os.Args[1:]) < 1 {
		_, _ = os.Stderr.WriteString("Error: No command given\n")
		os.Exit(1)
	}

	config, err := parseConfig(doFile)
	if err != nil {
		panic(err)
	}

	runner := makeRunner(config.Container.Runner)
	containerExists, err := runner.DoesContainerExist(config.Container)
	if err != nil {
		panic(err)
	}

	if !containerExists {
		err = runner.CreateContainer(config.Container)
		if err != nil {
			panic(err)
		}
	}

	containerRunning, err := runner.IsContainerRunning(config.Container)
	if err != nil {
		panic(err)
	}

	if !containerRunning {
		err = runner.RestartContainer(config.Container)
		if err != nil {
			panic(err)
		}
	}

	err = runner.ExecuteCommand(config.Container, os.Args[1:])
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		} else {
			panic(err)
		}
	}
}
