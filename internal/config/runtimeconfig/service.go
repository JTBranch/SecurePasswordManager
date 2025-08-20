package config

import (
	"encoding/json"
	"fmt"
	"os"

	buildconfig "go-password-manager/internal/config/buildconfig"

	"github.com/google/uuid"
)

// AppConfig represents the application configuration that is persisted on disk.
type AppConfig struct {
	KeyUUID      string `json:"keyUUID"`
	AppVersion   string `json:"appVersion"`
	WindowWidth  int    `json:"windowWidth"`
	WindowHeight int    `json:"windowHeight"`
}

// ConfigService manages application configuration
type ConfigService struct {
	Config *AppConfig
	path   string
}

// BuildConfigProvider defines an interface for accessing necessary build-time configuration.
// This allows for decoupling the runtime config service from the concrete build config implementation.
type BuildConfigProvider interface {
	GetConfigFilePath() (string, error)
	GetAppVersion() string
	GetUiConfig() buildconfig.UIConfig
}

// NewConfigService creates a new ConfigService.
// It loads an existing configuration from disk if one is found, otherwise it creates a new default config.
func NewConfigService(buildCfg BuildConfigProvider) (*ConfigService, error) {
	configPath, err := buildCfg.GetConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("could not determine config file path: %w", err)
	}

	// Try to load existing config
	loadedConfig, err := loadConfigFromFile(configPath)
	if err != nil {
		// If file doesn't exist or is corrupt, create a new default config
		uiConfig := buildCfg.GetUiConfig()
		newConfig := &AppConfig{
			AppVersion:   buildCfg.GetAppVersion(),
			WindowWidth:  uiConfig.Window.Width,
			WindowHeight: uiConfig.Window.Height,
		}

		// Generate a new KeyUUID
		newUUID, uuidErr := uuid.NewRandom()
		if uuidErr != nil {
			return nil, fmt.Errorf("failed to generate KeyUUID: %w", uuidErr)
		}
		newConfig.KeyUUID = newUUID.String()

		loadedConfig = newConfig
	}

	s := &ConfigService{
		Config: loadedConfig,
		path:   configPath,
	}

	return s, nil
}

// loadConfigFromFile loads the configuration from the specified file
func loadConfigFromFile(path string) (*AppConfig, error) {
	var cfg AppConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}

// Save saves the configuration to disk
func (cs *ConfigService) Save() error {
	data, err := json.MarshalIndent(cs.Config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cs.path, data, 0600)
}

// SetWindowSize sets the window dimensions in the configuration
func (cs *ConfigService) SetWindowSize(width, height int) error {
	cs.Config.WindowWidth = width
	cs.Config.WindowHeight = height
	return cs.Save()
}

// GetWindowSize returns the window dimensions from the configuration
func (cs *ConfigService) GetWindowSize() (int, int) {
	return cs.Config.WindowWidth, cs.Config.WindowHeight
}

// GetKeyUUID returns the key UUID from the configuration
func (cs *ConfigService) GetKeyUUID() string {
	return cs.Config.KeyUUID
}
