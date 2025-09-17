package buildconfig

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Set the environment to a known value that will use the default config
	os.Setenv("GO_PASSWORD_MANAGER_ENV", "test-default")
	defer os.Unsetenv("GO_PASSWORD_MANAGER_ENV")

	config, err := Load()
	require.NoError(t, err, "Failed to load config")

	// Test default values
	assert.Equal(t, "GoPasswordManager", config.Application.Name, "Expected app name 'GoPasswordManager'")
	assert.Equal(t, 1600, config.UI.Window.Width, "Expected window width 1600")
	assert.Equal(t, 900, config.UI.Window.Height, "Expected window height 900")
	assert.Equal(t, 32, config.Security.Encryption.KeySize, "Expected encryption key size 32")
}

func TestEnvironmentOverrides(t *testing.T) {
	// Set environment variables
	os.Setenv("APP_NAME", "TestApp")
	os.Setenv("DEFAULT_WINDOW_WIDTH", "1200")
	os.Setenv("DEBUG_LOGGING", "false")
	defer func() {
		os.Unsetenv("APP_NAME")
		os.Unsetenv("DEFAULT_WINDOW_WIDTH")
		os.Unsetenv("DEBUG_LOGGING")
	}()

	// Clear global config to force reload
	globalConfig = nil

	// Change to project root for test
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	if err := os.Chdir("../../"); err != nil {
		t.Skipf("Could not change to project root: %v", err)
	}

	config, err := Load()
	require.NoError(t, err, "Failed to load config")

	// Test environment overrides
	assert.Equal(t, "TestApp", config.Application.Name, "Expected app name 'TestApp'")
	assert.Equal(t, 1200, config.UI.Window.Width, "Expected window width 1200")
	assert.Equal(t, false, config.Logging.Debug, "Expected debug logging false")
}

func TestEnvironmentDetection(t *testing.T) {
	tests := []struct {
		env    string
		isDev  bool
		isProd bool
		isTest bool
	}{
		{"development", true, false, false},
		{"dev", true, false, false},
		{"production", false, true, false},
		{"prod", false, true, false},
		{"test", false, false, true},
		{"integration-test", false, false, true},
		{"e2e-test", false, false, true},
	}

	for _, tt := range tests {
		config := &Config{
			Application: ApplicationConfig{
				Environment: tt.env,
			},
		}

		assert.Equal(t, tt.isDev, config.IsDevelopment(), "Environment %s: expected IsDevelopment() %v", tt.env, tt.isDev)
		assert.Equal(t, tt.isProd, config.IsProduction(), "Environment %s: expected IsProduction() %v", tt.env, tt.isProd)
		assert.Equal(t, tt.isTest, config.IsTest(), "Environment %s: expected IsTest() %v", tt.env, tt.isTest)
	}
}

func TestDynamicConfigMerging(t *testing.T) {
	t.Run("Test Config Merging", (func(t *testing.T) {
		// Set working directory to configDir for test
		oldWd, _ := os.Getwd()
		os.Chdir("")
		defer os.Chdir(oldWd)

		// Set env so Load() picks up development.yaml
		os.Setenv("GO_PASSWORD_MANAGER_ENV", "development")
		defer os.Unsetenv("GO_PASSWORD_MANAGER_ENV")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "GoPasswordManager", cfg.Application.Name)
		assert.Equal(t, "development", cfg.Application.Environment)
		assert.Equal(t, 1600, cfg.UI.Window.Width)
		assert.Equal(t, 900, cfg.UI.Window.Height) // fallback from default
		assert.Equal(t, "default", cfg.UI.Theme)   // fallback from default
		assert.Equal(t, true, cfg.Logging.Debug)
		assert.Equal(t, "debug", cfg.Logging.Level)
		assert.Equal(t, "console", cfg.Logging.Format)
	}))
}
