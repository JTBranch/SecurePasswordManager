package logger

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockLoggerConfig struct {
	debug            bool
	level            string
	appName          string
	useUserConfigDir bool
}

func (m *MockLoggerConfig) IsDebug() bool {
	return m.debug
}

func (m *MockLoggerConfig) GetLogLevel() string {
	return m.level
}

func (m *MockLoggerConfig) UseUserConfigDir() bool {
	return m.useUserConfigDir
}

func (m *MockLoggerConfig) GetAppName() string {
	return m.appName
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

func TestGetLogDir(t *testing.T) {
	t.Run("creates and returns correct path", func(t *testing.T) {
		cfg := &MockLoggerConfig{appName: "GoPasswordManagerTest", useUserConfigDir: false}
		logDir, err := GetLogDir(cfg.GetAppName(), cfg)
		if err != nil {
			require.NoError(t, err, "GetLogDir returned error")
		}
		if logDir == "" {
			require.NotEmpty(t, logDir, "GetLogDir returned empty path")
		}
		// Check directory exists
		info, err := os.Stat(logDir)
		if err != nil {
			require.NoError(t, err, "Log directory does not exist")
		}
		if !info.IsDir() {
			require.DirExists(t, logDir, "Log path is not a directory")
		}
		expected := filepath.Join("..", "..", "tmp", "output")
		if logDir != expected {
			t.Errorf("Log directory path does not match expected: got %s, want %s", logDir, expected)
		}
	})
}

func TestLoggerAppendsToFile(t *testing.T) {
	cfg := &MockLoggerConfig{appName: "GoPasswordManagerTest", useUserConfigDir: false}
	Init(cfg)
	logFile := filepath.Join(GetLogDirPath(), "application.log")
	_ = os.Remove(logFile) // Ensure clean file
	f, err := os.Create(logFile)
	if err != nil {
		require.NoError(t, err, "Failed to create log file")
	}
	f.Close()
	Info("test log entry 1")
	Info("test log entry 2")
	data, err := os.ReadFile(logFile)
	if err != nil {
		require.NoError(t, err, "Failed to read log file")
	}
	content := string(data)
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 2 {
		require.GreaterOrEqual(t, len(lines), 2, "Expected at least 2 log lines")
	}
	logRegex := regexp.MustCompile(`^INFO: test log entry (1|2)$`)
	for _, line := range lines {
		if !logRegex.MatchString(line) {
			t.Errorf("Log line does not match expected format: %s", line)
		}
	}
}
