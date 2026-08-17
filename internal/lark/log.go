package lark

import (
	"context"

	"github.com/chyroc/lark"
	"github.com/sirupsen/logrus"
)

type larkLogger struct {
	logger *logrus.Logger
}

var levelMap = map[lark.LogLevel]logrus.Level{
	lark.LogLevelInfo:  logrus.InfoLevel,
	lark.LogLevelWarn:  logrus.WarnLevel,
	lark.LogLevelError: logrus.ErrorLevel,
}

func (l larkLogger) Log(ctx context.Context, larkLevel lark.LogLevel, msg string, args ...interface{}) {
	level, ok := levelMap[larkLevel]
	if !ok {
		return
	}
	l.logger.Logf(level, msg, args...)
}

func getLarkLogLevel(level string) lark.LogLevel {
	switch level {
	case "trace", "debug":
		return lark.LogLevelInfo
	case "info":
		return lark.LogLevelInfo
	case "warn":
		return lark.LogLevelWarn
	case "error":
		return lark.LogLevelError
	}
	return lark.LogLevelInfo
}
