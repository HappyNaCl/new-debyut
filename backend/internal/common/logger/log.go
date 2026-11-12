package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func Init() error {
	env := os.Getenv("ENVIRONMENT")
	var logger *zap.Logger
	var err error

	if strings.EqualFold(env, "production") {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		return err
	}

	Logger = logger
	return nil
}

func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}
