package main

import (
    "os"
    "strings"
)

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func firstMatchInString(value string, candidates []string) (string, bool) {
    for _, candidate := range candidates {
        if strings.Contains(value, candidate) {
            return candidate, true
        }
    }

    return "", false
}

func fileExists(filename string) bool {
    _, err := os.Stat(filename)
    return ! os.IsNotExist(err)
}
