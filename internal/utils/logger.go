package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// NewLogger initializes and returns a new logger instance.
func NewLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	return logger
}
