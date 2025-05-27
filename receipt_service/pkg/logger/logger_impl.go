package logger

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/sirupsen/logrus"
)

type customLogger struct {
	instance *logrus.Logger
	srcIP    string
}

type myFormatter struct {
	srcIP string
}

func (f *myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	if len(timestamp) > 19 {
		timestamp = timestamp[:19] + "," + timestamp[20:]
	}

	var line int
	if entry.HasCaller() {
		line = entry.Caller.Line
	}

	callerInfo := entry.Caller.Function
	if callerInfo == "" {
		callerInfo = "unknown"
	}

	logLine := fmt.Sprintf("%s %s %d %s : %s %s\n",
		timestamp,
		callerInfo,
		line,
		strings.ToUpper(entry.Level.String()),
		f.srcIP,
		entry.Message,
	)

	return []byte(logLine), nil
}

// Public constructor
func NewLogger() Logger {
	srcIP := getOutboundIP()

	l := logrus.New()
	l.SetReportCaller(true)
	l.SetFormatter(&myFormatter{srcIP: srcIP})

	return &customLogger{
		instance: l,
		srcIP:    srcIP,
	}
}

// Helper to get IP
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

// Logger methods
func (l *customLogger) Info(args ...interface{}) {
	l.instance.Info(args...)
}

func (l *customLogger) Error(args ...interface{}) {
	l.instance.Error(args...)
}

func (l *customLogger) Debug(args ...interface{}) {
	l.instance.Debug(args...)
}

func (l *customLogger) Warn(args ...interface{}) {
	l.instance.Warn(args...)
}

func (l *customLogger) WithFields(fields logrus.Fields) Logger {
	entry := l.instance.WithFields(fields)
	return &customLogger{instance: entry.Logger, srcIP: l.srcIP}
}
