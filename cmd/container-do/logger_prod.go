// +build !test

package main

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func makeLogger() (*zap.Logger, error) {
    prodConfig := zap.NewProductionConfig()
    prodConfig.Encoding = "console"
    prodConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    prodConfig.EncoderConfig.EncodeDuration= zapcore.StringDurationEncoder

    logger, err := prodConfig.Build()
    if err != nil {
        return nil, err
    }

    return logger, nil
}
