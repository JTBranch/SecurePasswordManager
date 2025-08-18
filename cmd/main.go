package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go-password-manager/internal/env"
	"go-password-manager/internal/envconfig"
	"go-password-manager/internal/logger"
	"go-password-manager/ui"
)

var (
	version = "development"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Handle version flag
	var showVersion = flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Go Password Manager %s\n", env.GetVersion())
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		os.Exit(0)
	}

	// Load environment configuration (both old and new systems)
	config, err := env.Load()
	if err != nil {
		log.Fatalf("Failed to load environment configuration: %v", err)
	}

	// Load YAML-based environment configuration
	yamlConfig, err := envconfig.Load()
	if err != nil {
		// If YAML config fails to load, log warning but continue
		log.Printf("Warning: Failed to load YAML configuration: %v", err)
	}

	// Initialize logger - use YAML config if available, otherwise fall back to env detection
	if yamlConfig != nil {
		logger.Init(yamlConfig.Logging.Debug)
	} else {
		logger.Init(env.IsDevMode())
	}

	// Set version in config if available
	if envVersion := env.GetVersion(); envVersion != "" {
		config.AppVersion = envVersion
	} else if version != "dev" {
		config.AppVersion = version
	}

	// Pass config to the UI
	app := ui.NewApp(config)
	app.Run()
}
