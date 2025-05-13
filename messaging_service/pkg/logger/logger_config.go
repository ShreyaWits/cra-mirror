package logger

import (
	"log"
	"net"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var srcIP string

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
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var err error
	Logger, err = config.Build()
	if err != nil {
		log.Fatal(err)
	}

	getOutboundIP()
}

// LogEvent logs an event with the given parameters
func LogEvent(requestID, eventType, userID, status, message string) {
	if Logger == nil {
		return
	}
	Logger.Info(message,
		zap.String("request_id", requestID),
		zap.String("event_type", eventType),
		zap.String("user_id", userID),
		zap.String("status", status),
		zap.String("source_ip", srcIP),
	)
}

// LogErrorEvent logs an error event with the given parameters
func LogErrorEvent(requestID, eventType, userID, status, message string) {
	if Logger == nil {
		return
	}
	Logger.Error(message,
		zap.String("request_id", requestID),
		zap.String("event_type", eventType),
		zap.String("user_id", userID),
		zap.String("status", status),
		zap.String("source_ip", srcIP),
	)
}

// LogWarnEvent logs a warning event with the given parameters
func LogWarnEvent(requestID, eventType, userID, status, message string) {
	if Logger == nil {
		return
	}
	Logger.Warn(message,
		zap.String("request_id", requestID),
		zap.String("event_type", eventType),
		zap.String("user_id", userID),
		zap.String("status", status),
		zap.String("source_ip", srcIP),
	)
}

// LogDebugEvent logs a debug event with the given parameters
func LogDebugEvent(requestID, eventType, userID, status, message string) {
	if Logger == nil {
		return
	}
	Logger.Debug(message,
		zap.String("request_id", requestID),
		zap.String("event_type", eventType),
		zap.String("user_id", userID),
		zap.String("status", status),
		zap.String("source_ip", srcIP),
	)
}

// Error logs an error message with optional error
func Error(message string, err error) {
	if Logger == nil {
		return
	}
	if err != nil {
		Logger.Error(message,
			zap.Error(err),
			zap.String("source_ip", srcIP),
		)
	} else {
		Logger.Error(message,
			zap.String("source_ip", srcIP),
		)
	}
}
