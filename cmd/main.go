package main

import (
	"flag"
	"fmt"
	"log"

	buildconfig "go-password-manager/internal/config/buildconfig"
	"go-password-manager/internal/logger"
	"go-password-manager/ui"
	"go-password-manager/ui/di"

	// Side-effect import to register LAN transport
	_ "go-password-manager/internal/transport/lan"
)

var (
	commit = "none"
	date   = "unknown"
)

func runApp() error {
	// Handle version flag
	var showVersion = flag.Bool("version", false, "Show version information")
	flag.Parse()

	buildCfg, err := buildconfig.Load()
	if err != nil {
		return fmt.Errorf("load build config: %w", err)
	}

	if *showVersion {
		fmt.Printf("Go Password Manager %s\n", buildCfg.Application.Version)
		fmt.Printf("Commit: %s\n", commit)
		fmt.Printf("Built: %s\n", date)
		return nil
	}

	logger.Init(buildCfg)

	bundle, err := di.BuildSharing(buildCfg)
	if err != nil {
		return fmt.Errorf("build sharing: %w", err)
	}
	app := ui.NewApp(buildCfg, bundle.SecretsService, bundle.TransferService)
	app.Run()
	return nil
}

func main() {
	if err := runApp(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}
