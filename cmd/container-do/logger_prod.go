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

    return prodConfig.Build()
}
