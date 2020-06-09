package main

import (
    "fmt"
    "github.com/pelletier/go-toml"
    "io/ioutil"
)

func main() {
    bytes, _ := ioutil.ReadFile("Dofile")
    containerConfig := Container{}
    _ = toml.Unmarshal(bytes, &containerConfig)
    fmt.Println(containerConfig.Image)
}
