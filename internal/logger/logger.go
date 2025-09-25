package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger
var debugEnabled bool
var logDir string

type LoggerConfig interface {
	IsDebug() bool
	GetLogLevel() string
	UseUserConfigDir() bool
	GetAppName() string
}

// Init initializes the logger with the specified debug mode and sets up the log directory
func Init(cfg LoggerConfig) {
	dir, err := GetLogDir(cfg.GetAppName(), cfg)
	if err != nil {
		panic("Failed to initialize log directory: " + err.Error())
	}
	logDir = dir

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

// GetLogDirPath returns the current log directory path
func GetLogDirPath() string {
	return logDir
}

func GetLogDir(appName string, cfg LoggerConfig) (string, error) {
	var logDir string
	if cfg.UseUserConfigDir() {
		base, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		logDir = filepath.Join(base, appName, "logs")
	} else {
		logDir = filepath.Join("..", "..", "tmp", "output") // Use project root tmp/output
	}
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return "", err
	}
	// Create application log file if it does not exist
	logFile := filepath.Join(logDir, "application.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		f, err := os.Create(logFile)
		if err != nil {
			return "", err
		}
		f.Close()
	}
	return logDir, nil
}

func appendToLogFile(level string, msgs ...string) {
	if logDir == "" {
		return
	}
	logFile := filepath.Join(logDir, "application.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	for _, msg := range msgs {
		fmt.Fprintln(f, level+": "+msg)
	}
}

// Debug logs debug messages only if debug is enabled
func Debug(msgs ...string) {
	if debugEnabled && log != nil {
		log.WithOptions(zap.AddCallerSkip(1)).Debug(strings.Join(msgs, " | "))
		appendToLogFile("DEBUG", msgs...)
	}
}

// Info logs info messages
func Info(msgs ...string) {
	if log != nil {
		log.WithOptions(zap.AddCallerSkip(1)).Info(strings.Join(msgs, " | "))
		appendToLogFile("INFO", msgs...)
	}
}

// Warn logs warning messages
func Warn(msgs ...string) {
	if log != nil {
		log.WithOptions(zap.AddCallerSkip(1)).Warn(strings.Join(msgs, " | "))
		appendToLogFile("WARN", msgs...)
	}
}

// Error logs error messages
func Error(msg string, err error) {
	if log != nil {
		log.WithOptions(zap.AddCallerSkip(1)).Error(msg, zap.Error(err))
		appendToLogFile("ERROR", msg, err.Error())
	}
}
