package main

import (
	"log"

	"go-password-manager/internal/env"
	"go-password-manager/ui"
)

func main() {
	// Load environment configuration
	config, err := env.Load()
	if err != nil {
		log.Fatalf("Failed to load environment configuration: %v", err)
	}

	// Pass config to the UI
	app := ui.NewApp(config)
	app.Run()
}
