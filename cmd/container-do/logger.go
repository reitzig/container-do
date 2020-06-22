package main

import (
    "os"

    "go.uber.org/zap"
)

func makeLogger() (*zap.Logger, error) {
    if value, isSet := os.LookupEnv("CONTAINER_DO_LOGGING"); isSet && value == "debug"  {
        return zap.NewDevelopment()
    } else {
        prodConfig := zap.NewProductionConfig()
        prodConfig.Encoding = "console"
        prodConfig.EncoderConfig.TimeKey = ""
        prodConfig.EncoderConfig.NameKey = ""
        prodConfig.EncoderConfig.CallerKey = ""

        return prodConfig.Build()
    }
}
