package main

import (
    "fmt"
    "io/ioutil"
    . "path/filepath"
    "strings"
    "time"

    "github.com/pelletier/go-toml"
    "go.uber.org/zap"
)

type runner string

const (
    docker runner = "docker"
    podman runner = "podman"
)

// These are the known and explicitly supported container OS flavors
var OsFlavors = []string{
    "debian",
    "fedora",
    "alpine",
    "gnu/linux", // for users: assert GNU tools
    "busybox",   // for users: assert BusyBox tools
}

type container struct {
    Image string
    // Build  string // TODO: implement building image from Dockerfile?
    RawOsFlavor string `toml:"os_flavor" default:""`
    osFlavor    string // lazily populated by runnerExec.DetermineOsFlavor

    Name        string
    WorkDir     string `toml:"work_dir"`
    Environment map[string]string
    Mounts      []string
    Ports       []string

    RawKeepAlive string `toml:"keep_alive" default:"15m"`
    KeepStopped  bool   `toml:"keep_stopped" default:"false"`
}

func (c *container) KeepAlive() time.Duration {
    d, err := time.ParseDuration(c.RawKeepAlive)
    if err != nil {
        panic(err)
    }
    return d
}

type thingToRun struct {
    Attach     bool   `default:"false"`
    User       string `default:""`
    ScriptFile string `toml:"script_file" default:""`
    Commands   []string
}

func (t *thingToRun) willRunSomething() bool {
    return t.ScriptFile != "" || len(t.Commands) > 0
}

type thingsToRun struct {
    Setup  thingToRun
    Before thingToRun
    After  thingToRun
}

func (t *thingsToRun) asMap() map[string]thingToRun {
    return map[string]thingToRun{
        "at container creation": t.Setup,
        "before each command": t.Before,
        "after each command": t.After,
    }
}

type Config struct {
    Runner runner `default:"docker"`
    // TODO: allow setting executable explicitly?
    Container container

    ThingsToRun thingsToRun `toml:"run"`
}

const ConfigFileTemplate = `
[container]
# image = "insert name/URL here"
# os_flavor = ""

# name = "project-do"
# work_dir = "WORKDIR"

# mounts = [".:$work_dir"]
# ports = []
# keep_alive = "15m"
# keep_stopped = false

[container.environment]
# KEY = "value"

[run.setup]
# attach      = false
# user        = ""
# script_file = ""
# commands    = []

[run.before]
# attach      = false
# user        = ""
# script_file = ""
# commands    = []

[run.after]
# attach      = false
# user        = ""
# script_file = ""
# commands    = []
`

type ConfigError struct {
    Message string
}

func (e ConfigError) Error() string {
    return fmt.Sprintf("Bad config file: %s", e.Message)
}

func parseConfig(fileName string) (Config, error) {
    config := Config{}

    if !fileExists(doFile) {
        return config, UsageError{Message: fmt.Sprintf("Config file %s missing", doFile)}
    }

    bytes, err := ioutil.ReadFile(fileName)
    if err != nil {
        return config, ConfigError{Message: err.Error()}
    }

    err = toml.Unmarshal(bytes, &config)
    if err != nil {
        return config, ConfigError{Message: err.Error()}
    }

    // Validation & Defaults
    if config.Container.Image == "" {
        return config, ConfigError{Message: "No image given"}
    }

    // NB: Go's fake enums don't protect against wrong values!
    switch r := config.Runner; r {
    case docker, podman:
        break
    default:
        return config, ConfigError{Message: "Invalid container runner: " + string(r)}
    }

    if config.Container.Name == "" {
        absolutePath, err := Abs(fileName)
        if err != nil {
            return config, ConfigError{Message: err.Error()}
        }

        config.Container.Name = strings.ToLower(Base(Dir(absolutePath))) + "-do"
    }

    for label, thingToRun := range config.ThingsToRun.asMap() {
        if thingToRun.ScriptFile != "" {
            zap.L().Sugar().Infof("Will run %s: %s", label, thingToRun.ScriptFile)
        }
        for _, command := range thingToRun.Commands {
            zap.L().Sugar().Infof("Will run %s: %s", label, command)
        }
    }

    zap.L().Sugar().Debugf("Parsed config: %+v", config)
    return config, err
}
