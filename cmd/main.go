package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go-password-manager/internal/config/buildconfig"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/logger"
	"go-password-manager/internal/service"
	"go-password-manager/internal/storage"
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

	buildCfg, err := buildconfig.Load()
	if err != nil {
		log.Fatalf("Failed to load build config: %v", err)
	}

	if *showVersion {
		fmt.Printf("Go Password Manager %s\n", buildCfg.Application.Version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		os.Exit(0)
	}

	logger.Init(buildCfg.Logging.Debug)

	// Setup services
	configService, err := config.NewConfigService(buildCfg)
	if err != nil {
		log.Fatalf("Failed to create config service: %v", err)
	}

	cryptoService, err := crypto.NewCryptoService(configService)
	if err != nil {
		log.Fatalf("Failed to create crypto service: %v", err)
	}

	secretsPath, err := buildCfg.GetSecretsFilePath()
	if err != nil {
		log.Fatalf("Failed to get secrets file path: %v", err)
	}
	storageService := storage.NewFileStorage(secretsPath, buildCfg.Application.Version, "e2e-user")

	secretsService := service.NewSecretsService(cryptoService, storageService)

	// Pass services to the UI
	app := ui.NewApp(buildCfg, secretsService)
	app.Run()
}
