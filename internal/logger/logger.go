package logger

import (
	"strings"

	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	l, _ := zap.NewDevelopment()
	log = l
}

// Debug logs debug messages
func Debug(msgs ...string) {
	log.WithOptions(zap.AddCallerSkip(1)).Debug(strings.Join(msgs, " | "))
}

// Info logs info messages
func Info(msgs ...string) {
	log.WithOptions(zap.AddCallerSkip(1)).Info(strings.Join(msgs, " | "))
}

// Warn logs warning messages
func Warn(msgs ...string) {
	log.WithOptions(zap.AddCallerSkip(1)).Warn(strings.Join(msgs, " | "))
}

// Error logs error messages
func Error(msgs ...string) {
	log.WithOptions(zap.AddCallerSkip(1)).Error(strings.Join(msgs, " | "))
}
