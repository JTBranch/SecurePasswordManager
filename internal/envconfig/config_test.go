package envconfig

import (
	"os"
	"testing"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Change to project root for test
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Change to project root (go up two levels from internal/envconfig)
	if err := os.Chdir("../../"); err != nil {
		t.Skipf("Could not change to project root: %v", err)
	}

	config, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test default values
	if config.Application.Name != "GoPasswordManager" {
		t.Errorf("Expected app name 'GoPasswordManager', got '%s'", config.Application.Name)
	}

	if config.UI.Window.Width != 1600 {
		t.Errorf("Expected window width 1600, got %d", config.UI.Window.Width)
	}

	if config.UI.Window.Height != 900 {
		t.Errorf("Expected window height 900, got %d", config.UI.Window.Height)
	}

	if config.Security.Encryption.KeySize != 32 {
		t.Errorf("Expected encryption key size 32, got %d", config.Security.Encryption.KeySize)
	}
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
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test environment overrides
	if config.Application.Name != "TestApp" {
		t.Errorf("Expected app name 'TestApp', got '%s'", config.Application.Name)
	}

	if config.UI.Window.Width != 1200 {
		t.Errorf("Expected window width 1200, got %d", config.UI.Window.Width)
	}

	if config.Logging.Debug != false {
		t.Errorf("Expected debug logging false, got %v", config.Logging.Debug)
	}
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

		if config.IsDevelopment() != tt.isDev {
			t.Errorf("Environment %s: expected IsDevelopment() %v, got %v", tt.env, tt.isDev, config.IsDevelopment())
		}

		if config.IsProduction() != tt.isProd {
			t.Errorf("Environment %s: expected IsProduction() %v, got %v", tt.env, tt.isProd, config.IsProduction())
		}

		if config.IsTest() != tt.isTest {
			t.Errorf("Environment %s: expected IsTest() %v, got %v", tt.env, tt.isTest, config.IsTest())
		}
	}
}
