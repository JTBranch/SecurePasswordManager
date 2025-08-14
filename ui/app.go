package ui

import (
	"fmt"
	"go-password-manager/internal/config"
	"go-password-manager/internal/env"
	"go-password-manager/internal/logger"
	"go-password-manager/ui/pages"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// App represents the main application
type App struct {
	fyneApp       fyne.App
	window        fyne.Window
	configService *config.ConfigService
	envConfig     *env.Config
}

// NewApp creates a new application instance
func NewApp(envConfig *env.Config) *App {
	fyneApp := app.New()
	window := fyneApp.NewWindow(envConfig.AppName)

	// Load legacy config service for window size persistence
	configService, err := config.NewConfigService()
	if err != nil {
		// Use environment config defaults
		window.Resize(fyne.NewSize(
			float32(envConfig.DefaultWindowWidth),
			float32(envConfig.DefaultWindowHeight),
		))
	} else {
		width, height := configService.GetWindowSize()
		if envConfig.DebugLogging {
			logger.Debug(fmt.Sprintf("Loaded window size from config, width: %d, height: %d", width, height))
		}
		window.Resize(fyne.NewSize(float32(width), float32(height)))
	}

	return &App{
		fyneApp:       fyneApp,
		window:        window,
		configService: configService,
		envConfig:     envConfig,
	}
}

// Run starts the application
func (a *App) Run() {
	a.window.SetContent(pages.MainPage(a.window))

	// Save window size on close
	a.window.SetOnClosed(func() {
		if a.configService != nil {
			size := a.window.Canvas().Size()
			_ = a.configService.SetWindowSize(int(size.Width), int(size.Height))
		}
	})

	a.window.ShowAndRun()
}
