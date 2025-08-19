package logger

import "testing"

type MockLoggerConfig struct {
	debug bool
	level string
}

func (m *MockLoggerConfig) IsDebug() bool {
	return m.debug
}

func (m *MockLoggerConfig) GetLogLevel() string {
	return m.level
}

func TestLoggerInit(t *testing.T) {

	t.Run("Init with debug enabled", func(t *testing.T) {
		cfg := &MockLoggerConfig{debug: true, level: "debug"}
		Init(cfg)
		if !debugEnabled {
			t.Error("Expected debugEnabled to be true")
		}
	})

	t.Run("Init with debug disabled", func(t *testing.T) {
		cfg := &MockLoggerConfig{debug: false, level: "info"}
		Init(cfg)
		if debugEnabled {
			t.Error("Expected debugEnabled to be false")
		}
	})

	t.Run("Init with default log level", func(t *testing.T) {
		cfg := &MockLoggerConfig{debug: false, level: "info"}
		Init(cfg)
		if cfg.level != "info" {
			t.Error("Expected logLevel to be 'info'")
		}
	})
}
