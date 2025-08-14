package config

import (
	"os"
	"path/filepath"
	"testing"
)

const errCreateConfigService = "Expected no error creating ConfigService, got: %v"

func TestConfigServiceBasic(t *testing.T) {
	// Test NewConfigService
	service, err := NewConfigService()
	if err != nil {
		t.Fatalf("Expected no error creating ConfigService, got: %v", err)
	}

	if service == nil {
		t.Fatal("Expected NewConfigService to return non-nil service")
	}

	if service.Config == nil {
		t.Fatal("Expected ConfigService to have non-nil Config")
	}

	// Check that default values are reasonable
	if service.Config.WindowWidth <= 0 {
		t.Errorf("Expected positive window width, got: %d", service.Config.WindowWidth)
	}

	if service.Config.WindowHeight <= 0 {
		t.Errorf("Expected positive window height, got: %d", service.Config.WindowHeight)
	}

	if service.Config.AppVersion == "" {
		t.Error("Expected non-empty app version")
	}
}

func TestConfigServiceSetWindowSize(t *testing.T) {
	service, err := NewConfigService()
	if err != nil {
		t.Fatalf(errCreateConfigService, err)
	}

	// Test SetWindowSize
	testWidth := 1024
	testHeight := 768

	err = service.SetWindowSize(testWidth, testHeight)
	if err != nil {
		t.Fatalf("Expected no error setting window size, got: %v", err)
	}

	// Verify values were updated
	if service.Config.WindowWidth != testWidth {
		t.Errorf("Expected window width %d, got: %d", testWidth, service.Config.WindowWidth)
	}

	if service.Config.WindowHeight != testHeight {
		t.Errorf("Expected window height %d, got: %d", testHeight, service.Config.WindowHeight)
	}
}

func TestConfigServiceSave(t *testing.T) {
	service, err := NewConfigService()
	if err != nil {
		t.Fatalf(errCreateConfigService, err)
	}

	// Modify config
	service.Config.WindowWidth = 1200
	service.Config.WindowHeight = 900

	// Test Save
	err = service.Save()
	if err != nil {
		t.Fatalf("Expected no error saving config, got: %v", err)
	}

	// Create new service to verify persistence
	newService, err := NewConfigService()
	if err != nil {
		t.Fatalf("Expected no error creating new ConfigService, got: %v", err)
	}

	// Verify values were persisted
	if newService.Config.WindowWidth != 1200 {
		t.Errorf("Expected persisted window width 1200, got: %d", newService.Config.WindowWidth)
	}

	if newService.Config.WindowHeight != 900 {
		t.Errorf("Expected persisted window height 900, got: %d", newService.Config.WindowHeight)
	}
}

func TestConfigFilePath(t *testing.T) {
	// Test that configFilePath function works
	path, err := configFilePath()
	if err != nil {
		t.Fatalf("Expected no error getting config file path, got: %v", err)
	}

	if path == "" {
		t.Error("Expected non-empty config file path")
	}

	// Test that the directory exists after calling configFilePath
	if _, err := os.Stat(path); err != nil {
		// The file might not exist yet, but the directory should
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Error("Expected config directory to be created")
		}
	}
}
