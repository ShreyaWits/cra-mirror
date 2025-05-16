package logger

import (
	"os"
	"reflect"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestInitLogger(t *testing.T) {
	// Define test parameters
	lokiUrl := "http://localhost:3100"

	// Initialize the logger
	InitLogger(lokiUrl)

	// Check if the logger is initialized correctly
	if Log == nil {
		t.Error("Log is nil")
	}

	// Check if the logger has the correct type
	if reflect.TypeOf(Log).String() != "*logrus.Logger" {
		t.Errorf("Expected logger type to be *logrus.Logger, but got %s", reflect.TypeOf(Log).String())
	}

	// Check if the logger has the correct level
	if Log.GetLevel() != logrus.InfoLevel {
		t.Errorf("Expected logger level to be %v, but got %v", logrus.InfoLevel, Log.GetLevel())
	}

	// Check if the logger has the correct formatter
	if reflect.TypeOf(Log.Formatter).String() != "*logrus.JSONFormatter" {
		t.Errorf("Expected logger formatter to be *logrus.JSONFormatter, but got %s", reflect.TypeOf(Log.Formatter).String())
	}

	// TODO: Find a way to check if the logger has the Loki hook
	// This is difficult to test without exposing the underlying hook
	// or using reflection to access the private fields.
}

func TestMain(m *testing.M) {
	// Set up logger for testing
	// You might need to start a Loki instance for testing

	exitCode := m.Run()

	// Clean up resources after testing
	os.Exit(exitCode)
}