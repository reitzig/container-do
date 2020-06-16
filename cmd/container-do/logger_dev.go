// +build test

package main

import "go.uber.org/zap"

func makeLogger() (*zap.Logger, error) {
    return zap.NewDevelopment()
}
