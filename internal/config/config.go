package config

import (
    "path/filepath"
    "os"
)

type AppConfig struct {
    KeyUUID      string `json:"keyUUID"`
    AppVersion   string `json:"appVersion"`
    WindowWidth  int    `json:"windowWidth"`
    WindowHeight int    `json:"windowHeight"`
}

const configFileName = "app.config"
const appName = "GoPasswordManager"

func configFilePath() (string, error) {
    configDir, err := os.UserConfigDir()
    if err != nil {
        return "", err
    }
    appConfigDir := filepath.Join(configDir, appName)
    if err := os.MkdirAll(appConfigDir, 0700); err != nil {
        return "", err
    }
    return filepath.Join(appConfigDir, configFileName), nil
}