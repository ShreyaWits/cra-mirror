package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func InitLogger() {
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)

	// Environment based formatting
	env := strings.ToLower(os.Getenv("ENV"))
	if env == "production" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	// Set log level
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
}

// LogInfo logs an info-level message with optional fields
func LogInfo(message string, fields map[string]interface{}) {
	entry := log.WithFields(logrus.Fields(fields))
	entry.Info(message)
}

// LogError logs an error-level message with optional error + fields
func LogError(message string, err error, fields map[string]interface{}) {
	entry := log.WithFields(logrus.Fields(fields))
	if err != nil {
		entry = entry.WithError(err)
	}
	entry.Error(message)
}

// LogWarning logs a warning-level message with optional fields
func LogWarning(message string, fields map[string]interface{}) {
	entry := log.WithFields(logrus.Fields(fields))
	entry.Warn(message)
}

// LogDebug logs a debug-level message with optional fields
func LogDebug(message string, fields map[string]interface{}) {
	entry := log.WithFields(logrus.Fields(fields))
	entry.Debug(message)
}

func PadLeft(num int, width int) string {
	return fmt.Sprintf("%0*d", width, num)
}