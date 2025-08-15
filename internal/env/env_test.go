package env_test

import (
	"go-password-manager/internal/env"
	"go-password-manager/tests/helpers"
	"os"
	"testing"
	"time"
)

const testDataDir = "/tmp/test"

func TestEnv(t *testing.T) {
	helpers.WithUnitTestCase(t, "LoadFromFile", func(tc *helpers.UnitTestCase) {
		// Create temporary test environment
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		tc.Require.NoError(err)
		defer func() {
			err := os.Chdir(oldWd)
			tc.Require.NoError(err)
		}()
		err = os.Chdir(tempDir)
		tc.Require.NoError(err)

		// Create test .env file
		envContent := `GO_PASSWORD_MANAGER_ENV=test
APP_NAME=TestApp
DEFAULT_WINDOW_WIDTH=800
DEFAULT_WINDOW_HEIGHT=600
DEBUG_LOGGING=false
TEST_DATA_DIR=` + testDataDir + `
E2E_TEST_TIMEOUT=15s
ENCRYPTION_KEY_SIZE=16`

		err = os.WriteFile(".env.local", []byte(envContent), 0644)
		tc.Require.NoError(err, "Failed to create test .env file")

		config, err := env.Load()
		tc.Require.NoError(err, "Load() failed")

		// Test values are loaded correctly
		tc.Assert.Equal("test", config.Environment, "Expected Environment=test")
		tc.Assert.Equal("TestApp", config.AppName, "Expected AppName=TestApp")
		tc.Assert.Equal(800, config.DefaultWindowWidth, "Expected DefaultWindowWidth=800")
		tc.Assert.Equal(600, config.DefaultWindowHeight, "Expected DefaultWindowHeight=600")
		tc.Assert.False(config.DebugLogging, "Expected DebugLogging=false")
		tc.Assert.Equal(testDataDir, config.TestDataDir, "Expected TestDataDir")
		tc.Assert.Equal(15*time.Second, config.E2ETestTimeout, "Expected E2ETestTimeout=15s")
		tc.Assert.Equal(16, config.EncryptionKeySize, "Expected EncryptionKeySize=16")
	})

	helpers.WithUnitTestCase(t, "DefaultValues", func(tc *helpers.UnitTestCase) {
		// Clear any existing env vars
		envVars := []string{
			"GO_PASSWORD_MANAGER_ENV",
			"APP_NAME",
			"DEFAULT_WINDOW_WIDTH",
			"DEFAULT_WINDOW_HEIGHT",
			"DEBUG_LOGGING",
			"TEST_DATA_DIR",
			"E2E_TEST_TIMEOUT",
			"ENCRYPTION_KEY_SIZE",
		}

		originalValues := make(map[string]string)
		for _, e := range envVars {
			originalValues[e] = os.Getenv(e)
			os.Unsetenv(e)
		}
		defer func() {
			// Restore original values
			for e, value := range originalValues {
				if value != "" {
					os.Setenv(e, value)
				}
			}
		}()

		// Create temporary directory without .env files
		tempDir := t.TempDir()
		oldWd, err := os.Getwd()
		tc.Require.NoError(err)
		defer func() {
			err := os.Chdir(oldWd)
			tc.Require.NoError(err)
		}()
		err = os.Chdir(tempDir)
		tc.Require.NoError(err)

		// Load config without .env file
		config, err := env.Load()
		tc.Require.NoError(err, "Load() with no .env file should not fail")

		// Test default values
		tc.Assert.Equal("dev", config.Environment, "Expected default Environment")
		tc.Assert.Equal("GoPasswordManager", config.AppName, "Expected default AppName")
		tc.Assert.Equal(1600, config.DefaultWindowWidth, "Expected default window width")
		tc.Assert.Equal(900, config.DefaultWindowHeight, "Expected default window height")
		tc.Assert.True(config.DebugLogging, "Expected default debug logging")
		tc.Assert.Equal("", config.TestDataDir, "Expected default test data dir")
		tc.Assert.Equal(30*time.Second, config.E2ETestTimeout, "Expected default E2E timeout")
		tc.Assert.Equal(32, config.EncryptionKeySize, "Expected default encryption key size")
	})

	helpers.WithUnitTestCase(t, "Get", func(tc *helpers.UnitTestCase) {
		// The Get() function uses a singleton pattern.
		// We can't easily reset the singleton, but we can test that it returns a valid config.
		config := env.Get()
		tc.Assert.NotNil(config, "Expected Get() to return a non-nil config")
		// A simple check to ensure it's a plausible config object
		tc.Assert.NotEmpty(config.AppName, "Expected AppName to be non-empty")
	})
}
