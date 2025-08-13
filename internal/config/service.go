package config

import (
    "encoding/json"
    "os"
)

type ConfigService struct {
    Config *AppConfig
    path   string
}

func NewConfigService() (*ConfigService, error) {
    path, err := configFilePath()
    if err != nil {
        return nil, err
    }
    var cfg AppConfig
    if _, err := os.Stat(path); os.IsNotExist(err) {
        cfg = AppConfig{
            AppVersion:   "1.0.0",
            WindowWidth:  1600,
            WindowHeight: 900,
        }
        data, _ := json.MarshalIndent(cfg, "", "  ")
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

func (cs *ConfigService) Save() error {
    data, err := json.MarshalIndent(cs.Config, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(cs.path, data, 0600)
}

func (cs *ConfigService) SetWindowSize(width, height int) error {
    cs.Config.WindowWidth = width
    cs.Config.WindowHeight = height
    return cs.Save()
}

func (cs *ConfigService) GetWindowSize() (int, int) {
    return cs.Config.WindowWidth, cs.Config.WindowHeight
}