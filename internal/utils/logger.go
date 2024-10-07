package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

func NewLogger() *logrus.Logger {
	 logger := logrus.New()
	 logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	 })
	 logger.SetOutput(os.Stdout)
	 logger.SetLevel(logrus.InfoLevel)
	 return logger
}
