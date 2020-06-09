package container_do

import "time"

type Setup struct {
    Privileged bool
    Script string
    Commands string
}

type runner string

const (
    docker runner = "docker"
    podman runner = "podman"
)

type Container struct {
    Runner runner
    Image  string
    Build string

    Name string
    WorkDir string
    Mounts []string

    Setup     Setup
    KeepAlive time.Duration
    KeepStopped bool
}
