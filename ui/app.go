package ui

import (
	"fmt"
	"go-password-manager/internal/config/buildconfig"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/logger"
	"go-password-manager/internal/service"
	"go-password-manager/ui/pages"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// App represents the main application
type App struct {
	fyneApp        fyne.App
	window         fyne.Window
	configService  *config.ConfigService
	buildConfig    *buildconfig.Config
	secretsService *service.SecretsService
}

// NewApp creates a new application instance
func NewApp(buildCfg *buildconfig.Config, secretsService *service.SecretsService) *App {
	fyneApp := app.New()
	window := fyneApp.NewWindow(buildCfg.Application.Name)

	// Load legacy config service for window size persistence
	configService, err := config.NewConfigService(buildCfg)
	if err != nil {
		// Use environment config defaults
		window.Resize(fyne.NewSize(
			float32(buildCfg.UI.Window.Width),
			float32(buildCfg.UI.Window.Height),
		))
	} else {
		width, height := configService.GetWindowSize()
		if buildCfg.Logging.Debug {
			logger.Debug(fmt.Sprintf("Loaded window size from config, width: %d, height: %d", width, height))
		}
		window.Resize(fyne.NewSize(float32(width), float32(height)))
	}

	return &App{
		fyneApp:        fyneApp,
		window:         window,
		configService:  configService,
		buildConfig:    buildCfg,
		secretsService: secretsService,
	}
}

// Run starts the application
func (a *App) Run() {
	a.window.SetContent(pages.MainPageWithService(a.window, a.secretsService))

	// Save window size on close
	a.window.SetOnClosed(func() {
		if a.configService != nil {
			size := a.window.Canvas().Size()
			_ = a.configService.SetWindowSize(int(size.Width), int(size.Height))
		}
	})

	a.window.ShowAndRun()
}
