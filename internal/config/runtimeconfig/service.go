package config

import (
	"encoding/json"
	"go-password-manager/internal/config/buildconfig"
	"os"
	"path/filepath"
)

// ConfigService manages application configuration
type ConfigService struct {
	Config *AppConfig
	path   string
}

// NewConfigService creates a new configuration service
func NewConfigService(buildCfg *buildconfig.Config) (*ConfigService, error) {
	path, err := buildCfg.GetConfigFilePath()
	if err != nil {
		return nil, err
	}

	var cfg AppConfig
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg = AppConfig{
			AppVersion:   buildCfg.Application.Version,
			WindowWidth:  buildCfg.UI.Window.Width,
			WindowHeight: buildCfg.UI.Window.Height,
		}
		data, _ := json.MarshalIndent(cfg, "", "  ")

		// Ensure directory exists
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}

		_ = os.WriteFile(path, data, 0600)
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(data, &cfg)
	}
	return &ConfigService{Config: &cfg, path: path}, nil
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
