package logger

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger
var srcIP string

type myFormatter struct{}

func (f *myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Format timestamp with milliseconds, replacing the dot with a comma.
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	// Replace dot with comma in milliseconds part.
	if len(timestamp) > 19 {
		timestamp = timestamp[:19] + "," + timestamp[20:]
	}

	// Get caller info (file, line and function)
	// var fileName string
	var line int
	// var funcName string
	if entry.HasCaller() {
		// You can further process the file name if desired.
		// _, file := path.Split(entry.Caller.File)
		// fileName = file
		line = entry.Caller.Line

		// Function name is usually the full path, so you might want to trim it.
		// funcParts := strings.Split(entry.Caller.Function, ".")
		// funcName = funcParts[len(funcParts)-1]
	}

	// Build a pseudo "package.class.method" string.
	// In Go, we don't have classes, but you can use the caller's function name.
	// For more detailed info, you could process entry.Caller.Function.
	callerInfo := entry.Caller.Function
	if callerInfo == "" {
		callerInfo = "unknown"
	}

	// Assemble the final log message.
	// Format:
	// <timestamp callerInfo line logLevel : srcIP message>
	logLine := fmt.Sprintf("%s %s %d %s : %s %s\n",
		timestamp,
		callerInfo,
		line,
		strings.ToUpper(entry.Level.String()),
		srcIP,
		entry.Message,
	)

	return []byte(logLine), nil
}

// Get preferred outbound ip of this machine
func getOutboundIP() {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	srcIP = conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func InitLogger() {
	Logger = logrus.New()

	getOutboundIP()

	// Enable reporting the caller information.
	Logger.SetReportCaller(true)

	// Set our custom formatter.
	Logger.SetFormatter(new(myFormatter))
}

// LogEvent logs an event with info level
func LogEvent(traceID, eventType, entity, level, message string) {
	Logger.WithFields(logrus.Fields{
		"trace_id":   traceID,
		"event_type": eventType,
		"entity":     entity,
	}).Info(message)
}

// LogErrorEvent logs an event with error level
func LogErrorEvent(traceID, eventType, entity, level, message string) {
	Logger.WithFields(logrus.Fields{
		"trace_id":   traceID,
		"event_type": eventType,
		"entity":     entity,
	}).Error(message)
}

// LogWarnEvent logs an event with warning level
func LogWarnEvent(traceID, eventType, entity, level, message string) {
	Logger.WithFields(logrus.Fields{
		"trace_id":   traceID,
		"event_type": eventType,
		"entity":     entity,
	}).Warn(message)
}
