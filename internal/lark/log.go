package lark

import (
	"context"

	"github.com/chyroc/lark"
	"github.com/sirupsen/logrus"
)

type larkLogger struct {
	logger *logrus.Logger
}

var (
	levelMap = map[lark.LogLevel]logrus.Level{
		lark.LogLevelTrace: logrus.TraceLevel,
		lark.LogLevelDebug: logrus.DebugLevel,
		lark.LogLevelInfo:  logrus.InfoLevel,
		lark.LogLevelWarn:  logrus.WarnLevel,
		lark.LogLevelError: logrus.ErrorLevel,
	}
)

func (l larkLogger) Log(ctx context.Context, larkLevel lark.LogLevel, msg string, args ...interface{}) {
	if level, ok := levelMap[larkLevel]; ok {
		l.logger.Logf(level, msg, args...)
		return
	}
	l.logger.Logf(logrus.InfoLevel, msg, args...)
}

func getLarkLogLevel(level string) lark.LogLevel {
	switch level {
	case "trace":
		return lark.LogLevelTrace
	case "debug":
		return lark.LogLevelDebug
	case "info":
		return lark.LogLevelInfo
	case "warn":
		return lark.LogLevelWarn
	case "error":
		return lark.LogLevelError
	}
	return lark.LogLevelInfo
}
