// +build test

package main

import "go.uber.org/zap"

func makeLogger() (*zap.Logger, error) {
    logger, err := zap.NewDevelopment()
    if err != nil {
        return nil, err
    }

    return logger, nil
}
