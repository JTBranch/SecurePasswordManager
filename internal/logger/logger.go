package logger

import (
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger
var debugEnabled bool

type LoggerConfig interface {
	IsDebug() bool
	GetLogLevel() string
}

// Init initializes the logger with the specified debug mode
func Init(cfg LoggerConfig) {
	debugEnabled = cfg.IsDebug()

	config := zap.NewProductionConfig()
	if debugEnabled {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		// Production mode: only show INFO and above
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		config.DisableCaller = true
		config.DisableStacktrace = true
	}

	l, _ := config.Build()
	log = l
}

// Debug logs debug messages only if debug is enabled
func Debug(msgs ...string) {
	if debugEnabled && log != nil {
		log.WithOptions(zap.AddCallerSkip(1)).Debug(strings.Join(msgs, " | "))
	}
}

// Info logs info messages
func Info(msgs ...string) {
	if log != nil {
		log.WithOptions(zap.AddCallerSkip(1)).Info(strings.Join(msgs, " | "))
	}
}

// Warn logs warning messages
func Warn(msgs ...string) {
	if log != nil {
		log.WithOptions(zap.AddCallerSkip(1)).Warn(strings.Join(msgs, " | "))
	}
}

// Error logs error messages
func Error(msgs ...string) {
	if log != nil {
		log.WithOptions(zap.AddCallerSkip(1)).Error(strings.Join(msgs, " | "))
	}
}
