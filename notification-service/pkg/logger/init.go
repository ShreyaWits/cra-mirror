package logger

import (
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger(lokiUrl string) {
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetLevel(logrus.InfoLevel)
	// send log to loki

	// Configure the Loki hook
	// opts := lokirus.NewLokiHookOptions().
	// 	WithLevelMap(lokirus.LevelMap{logrus.PanicLevel: "critical"}).
	// 	WithFormatter(&logrus.JSONFormatter{}).
	// 	WithStaticLabels(lokirus.Labels{
	// 		"app":         "notification-service",
	// 		"environment": "prod",
	// 	})

	// hook := lokirus.NewLokiHookWithOpts(
	// 	lokiUrl,
	// 	opts,
	// 	logrus.InfoLevel,
	// 	logrus.WarnLevel,
	// 	logrus.ErrorLevel,
	// 	logrus.FatalLevel)

	// Log.AddHook(hook)

}
