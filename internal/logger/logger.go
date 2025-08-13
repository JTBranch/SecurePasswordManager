package logger

import (
    "go.uber.org/zap"
    "strings"
)

var log *zap.Logger

func init() {
    l, _ := zap.NewDevelopment()
    log = l
}

func Debug(msgs ...string) {
    log.WithOptions(zap.AddCallerSkip(1)).Debug(strings.Join(msgs, " | "))
}

func Info(msgs ...string) {
    log.WithOptions(zap.AddCallerSkip(1)).Info(strings.Join(msgs, " | "))
}

func Warn(msgs ...string) {
    log.WithOptions(zap.AddCallerSkip(1)).Warn(strings.Join(msgs, " | "))
}

func Error(msgs ...string) {
    log.WithOptions(zap.AddCallerSkip(1)).Error(strings.Join(msgs, " | "))
}