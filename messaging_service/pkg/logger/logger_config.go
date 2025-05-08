package logger

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
	srcIP  string = "0.0.0.0"
)

// Get the preferred outbound IP of this machine
func getOutboundIP() {
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	srcIP = conn.LocalAddr().(*net.UDPAddr).IP.String()
}

// InitLogger initializes the logger
func InitLogger() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var err error
	Logger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build logger: %v", err))
	}

	getOutboundIP()
	fmt.Println("Logger initialized")
}

// Error logs an error message
func Error(message string, err error) {
	if Logger == nil {
		fmt.Println("Logger not initialized")
		return
	}

	errorMsg := message
	if err != nil {
		errorMsg = fmt.Sprintf("%s: %v", message, err)
	}

	Logger.Error(errorMsg,
		zap.String("ipAddress", srcIP),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
	)
}

// LogEvent logs an event with requestId
func LogEvent(RequestID string, EventType string, UserID string, Status string, Message string) {
	if Logger == nil {
		fmt.Println("Logger not initialized")
		return
	}

	Logger.Info(Message,
		zap.String("requestId", RequestID),
		zap.String("eventType", EventType),
		zap.String("userId", UserID),
		zap.String("status", Status),
		zap.String("ipAddress", srcIP),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
	)
}

// LogErrorEvent logs an error event
func LogErrorEvent(RequestID string, EventType string, UserID string, Status string, Message string) {
	if Logger == nil {
		fmt.Println("Logger not initialized")
		return
	}

	Logger.Error(Message,
		zap.String("requestId", RequestID),
		zap.String("eventType", EventType),
		zap.String("userId", UserID),
		zap.String("status", Status),
		zap.String("ipAddress", srcIP),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
	)
}

// LogWarnEvent logs a warning event
func LogWarnEvent(RequestID string, EventType string, UserID string, Status string, Message string) {
	if Logger == nil {
		fmt.Println("Logger not initialized")
		return
	}

	Logger.Warn(Message,
		zap.String("requestId", RequestID),
		zap.String("eventType", EventType),
		zap.String("userId", UserID),
		zap.String("status", Status),
		zap.String("ipAddress", srcIP),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
	)
}

// LogDebugEvent logs a debug event
func LogDebugEvent(RequestID string, EventType string, UserID string, Status string, Message string) {
	if Logger == nil {
		fmt.Println("Logger not initialized")
		return
	}

	Logger.Debug(Message,
		zap.String("requestId", RequestID),
		zap.String("eventType", EventType),
		zap.String("userId", UserID),
		zap.String("status", Status),
		zap.String("ipAddress", srcIP),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
	)
}

// LogWarnEvent logs a warning event
func LogInfo(EventType string, Message string) {
	if Logger == nil {
		fmt.Println("Logger not initialized", EventType, Message)
		return
	}

	Logger.Warn(Message,
		zap.String("eventType", EventType),
		zap.String("ipAddress", srcIP),
		zap.String("timestamp", time.Now().Format(time.RFC3339)),
	)
}
