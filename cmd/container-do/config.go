package main

import (
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"time"
)

type setup struct {
	Privileged bool   `default:"false"`
	Script     string `default:""`
	Commands   string `default:""`
}

type runner string

const (
	docker runner = "docker"
	podman runner = "podman"
)

// TODO: Doesn't work?
//type myDuration struct {
//    Value duration.Duration
//}
//
//func (m *myDuration) UnmarshalTOML(p interface{}) error {
//    d, err := duration.ParseISO8601(p.(string))
//    m.Value = d
//    return err
//}

// TODO: add environment variables
type container struct {
	Runner runner `default:"docker"`
	Image  string
	Build  string

	Name        string
	WorkDir     string
	Environment map[string]string
	Mounts      []string

	Setup setup
	//KeepAlive   myDuration //`toml:"keep_alive"`
	RawKeepAlive string `toml:"keep_alive"`
	KeepStopped  bool   `default:"false"`
}

func (c *container) KeepAlive() time.Duration {
	d, err := time.ParseDuration(c.RawKeepAlive)
	if err != nil {
		panic(err)
	}
	return d
}

type Config struct {
	Container container
}

func parseConfig(fileName string) (Config, error) {
	config := Config{}
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return config, err
	}

	err = toml.Unmarshal(bytes, &config)
	return config, err
}
