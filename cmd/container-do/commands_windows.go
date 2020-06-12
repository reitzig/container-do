// +build windows

package main

func (d DockerRunner) RunnerExecutable() string {
	return "docker.exe"
}
