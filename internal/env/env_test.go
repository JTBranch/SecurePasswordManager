package env

import (
	"os"
	"testing"
	"time"
)

const testDataDir = "/tmp/test"

func TestLoad(t *testing.T) {
	// Create temporary test environment
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create test .env file
	envContent := `GO_PASSWORD_MANAGER_ENV=test
APP_NAME=TestApp
DEFAULT_WINDOW_WIDTH=800
DEFAULT_WINDOW_HEIGHT=600
DEBUG_LOGGING=false
TEST_DATA_DIR=` + testDataDir + `
E2E_TEST_TIMEOUT=15s
ENCRYPTION_KEY_SIZE=16`

	err := os.WriteFile(".env.local", []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Test values are loaded correctly
	if config.Environment != "test" {
		t.Errorf("Expected Environment=test, got %s", config.Environment)
	}

	if config.AppName != "TestApp" {
		t.Errorf("Expected AppName=TestApp, got %s", config.AppName)
	}

	if config.DefaultWindowWidth != 800 {
		t.Errorf("Expected DefaultWindowWidth=800, got %d", config.DefaultWindowWidth)
	}

	if config.DefaultWindowHeight != 600 {
		t.Errorf("Expected DefaultWindowHeight=600, got %d", config.DefaultWindowHeight)
	}

	if config.DebugLogging != false {
		t.Errorf("Expected DebugLogging=false, got %t", config.DebugLogging)
	}

	if config.TestDataDir != testDataDir {
		t.Errorf("Expected TestDataDir=%s, got %s", testDataDir, config.TestDataDir)
	}

	if config.E2ETestTimeout != 15*time.Second {
		t.Errorf("Expected E2ETestTimeout=15s, got %v", config.E2ETestTimeout)
	}

	if config.EncryptionKeySize != 16 {
		t.Errorf("Expected EncryptionKeySize=16, got %d", config.EncryptionKeySize)
	}
}

func TestDefaultValues(t *testing.T) {
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
	for _, env := range envVars {
		originalValues[env] = os.Getenv(env)
		os.Unsetenv(env)
	}
	defer func() {
		// Restore original values
		for env, value := range originalValues {
			if value != "" {
				os.Setenv(env, value)
			}
		}
	}()

	// Create temporary directory without .env files
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Test default values
	if config.Environment != "dev" {
		t.Errorf("Expected default Environment=dev, got %s", config.Environment)
	}

	if config.AppName != "GoPasswordManager" {
		t.Errorf("Expected default AppName=GoPasswordManager, got %s", config.AppName)
	}

	if config.DefaultWindowWidth != 1600 {
		t.Errorf("Expected default DefaultWindowWidth=1600, got %d", config.DefaultWindowWidth)
	}

	if config.DebugLogging != true {
		t.Errorf("Expected default DebugLogging=true, got %t", config.DebugLogging)
	}
}

func TestEnvironmentMethods(t *testing.T) {
	tests := []struct {
		env           string
		isProduction  bool
		isDevelopment bool
		isTest        bool
	}{
		{"prod", true, false, false},
		{"production", true, false, false},
		{"dev", false, true, false},
		{"development", false, true, false},
		{"test", false, false, true},
		{"staging", false, false, false},
	}

	for _, test := range tests {
		config := &Config{Environment: test.env}

		if config.IsProduction() != test.isProduction {
			t.Errorf("Environment %s: expected IsProduction()=%t, got %t",
				test.env, test.isProduction, config.IsProduction())
		}

		if config.IsDevelopment() != test.isDevelopment {
			t.Errorf("Environment %s: expected IsDevelopment()=%t, got %t",
				test.env, test.isDevelopment, config.IsDevelopment())
		}

		if config.IsTest() != test.isTest {
			t.Errorf("Environment %s: expected IsTest()=%t, got %t",
				test.env, test.isTest, config.IsTest())
		}
	}
}

func TestGetSecretsFilePath(t *testing.T) {
	testCases := []struct {
		name        string
		environment string
		testDataDir string
		expected    string
	}{
		{
			name:        "test environment with test data dir",
			environment: "test",
			testDataDir: testDataDir,
			expected:    testDataDir + "/" + DefaultSecretsFileName,
		},
		{
			name:        "development environment",
			environment: "dev",
			testDataDir: "",
			expected:    DefaultSecretsFileName,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &Config{
				Environment: tc.environment,
				TestDataDir: tc.testDataDir,
				AppName:     "TestApp",
			}

			result := config.GetSecretsFilePath()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestGetSecretsFilePathProduction(t *testing.T) {
	config := &Config{
		Environment: "prod",
		AppName:     "TestApp",
	}

	result := config.GetSecretsFilePath()
	if result == "" {
		t.Error("Production secrets file path should not be empty")
	}
}

func TestGetConfigFilePath(t *testing.T) {
	config := &Config{
		Environment: "test",
		TestDataDir: testDataDir,
		AppName:     "TestApp",
	}

	result := config.GetConfigFilePath()
	expected := testDataDir + "/" + DefaultConfigFileName

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestGlobalConfig(t *testing.T) {
	// Reset global config
	globalConfig = nil

	// First call should auto-load
	config1 := Get()
	if config1 == nil {
		t.Error("Get() should auto-load config")
	}

	// Second call should return same instance
	config2 := Get()
	if config1 != config2 {
		t.Error("Get() should return same instance on subsequent calls")
	}
}
