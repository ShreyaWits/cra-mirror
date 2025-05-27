package logger

import "github.com/sirupsen/logrus"

type Logger interface {
	Info(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	Warn(args ...interface{})

	// Optional: You can expose WithFields to allow structured logging.
	WithFields(fields logrus.Fields) Logger
}
