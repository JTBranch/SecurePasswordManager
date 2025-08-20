package config_test

import (
	buildconfig "go-password-manager/internal/config/buildconfig"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/tests/helpers"
	"os"
	"path/filepath"
	"testing"
)

// mockBuildConfig is a mock implementation of the BuildConfigProvider interface for testing.
type mockBuildConfig struct {
	ConfigPath string
	PathError  error
	AppVersion string
	ui         buildconfig.UIConfig
}

// GetConfigFilePath returns a mock config path or an error.
func (m *mockBuildConfig) GetConfigFilePath() (string, error) {
	if m.PathError != nil {
		return "", m.PathError
	}
	return m.ConfigPath, nil
}

// GetAppVersion returns a mock app version.
func (m *mockBuildConfig) GetAppVersion() string {
	return m.AppVersion
}

// GetWindowSize returns mock window dimensions.
func (m *mockBuildConfig) GetUiConfig() buildconfig.UIConfig {
	return m.ui
}

func TestNewConfigServiceInitialization(t *testing.T) {
	helpers.WithUnitTestCase(t, "Initializes with default values when config file does not exist", func(tc *helpers.UnitTestCase) {
		// Create a temporary directory to ensure the path is valid but the file is empty
		tempDir, err := os.MkdirTemp("", "config-init-test-")
		tc.Require.NoError(err, "Failed to create temp dir")
		defer os.RemoveAll(tempDir)

		configFilePath := filepath.Join(tempDir, "non-existent-config.json")

		// Setup mock to return a valid path to a non-existent file.
		// This will cause `loadConfigFromFile` to fail, triggering the default creation logic.
		mockProvider := &mockBuildConfig{
			ConfigPath: configFilePath,
			PathError:  nil, // No error getting the path
			AppVersion: "v1.2.3-mock",
			ui: buildconfig.UIConfig{
				Window: buildconfig.WindowConfig{
					Width:  1280,
					Height: 720,
				},
			},
		}

		service, err := config.NewConfigService(mockProvider)
		tc.Require.NoError(err, "NewConfigService should not fail when config file doesn't exist")
		tc.Require.NotNil(service, "Service should not be nil")

		// Assert that the config was initialized with values from our mock provider
		tc.Assert.NotEmpty(service.Config.KeyUUID, "A new KeyUUID should be generated")
		tc.Assert.Equal("v1.2.3-mock", service.Config.AppVersion, "AppVersion should match the mock provider")
		tc.Assert.Equal(1280, service.Config.WindowWidth, "WindowWidth should match the mock provider")
		tc.Assert.Equal(720, service.Config.WindowHeight, "WindowHeight should match the mock provider")
	})
}

func TestConfigServiceSaveAndLoad(t *testing.T) {
	helpers.WithUnitTestCase(t, "Saves and loads configuration correctly", func(tc *helpers.UnitTestCase) {
		tempDir, err := os.MkdirTemp("", "config-test-")
		tc.Require.NoError(err, "Failed to create temp dir")
		defer os.RemoveAll(tempDir)

		configFilePath := filepath.Join(tempDir, "app.config.json")

		// 1. Create the first service and save its config
		mockProvider1 := &mockBuildConfig{ConfigPath: configFilePath}
		service1, err := config.NewConfigService(mockProvider1)
		tc.Require.NoError(err, "Creating service 1 should not fail")

		service1.Config.WindowWidth = 1024
		err = service1.Save()
		tc.Require.NoError(err, "Saving config should not produce an error")

		// 2. Create a second service that should load the file saved by the first
		mockProvider2 := &mockBuildConfig{ConfigPath: configFilePath}
		service2, err := config.NewConfigService(mockProvider2)
		tc.Require.NoError(err, "Creating service 2 should not fail")

		tc.Assert.Equal(1024, service2.Config.WindowWidth, "WindowWidth should be loaded from the saved file")
	})
}
