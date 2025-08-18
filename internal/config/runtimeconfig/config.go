package config

// AppConfig represents the application configuration that is persisted on disk.
type AppConfig struct {
	KeyUUID      string `json:"keyUUID"`
	AppVersion   string `json:"appVersion"`
	WindowWidth  int    `json:"windowWidth"`
	WindowHeight int    `json:"windowHeight"`
}
