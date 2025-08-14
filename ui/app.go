package ui

import (
	"fmt"
	"go-password-manager/internal/config"
	"go-password-manager/internal/logger"
	"go-password-manager/ui/pages"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

type App struct {
	fyneApp       fyne.App
	window        fyne.Window
	configService *config.ConfigService
}

func NewApp() *App {
	fyneApp := app.New()
	window := fyneApp.NewWindow("Password Manager")

	// Load config and set window size
	configService, err := config.NewConfigService()
	if err != nil {
		window.Resize(fyne.NewSize(667, 675))
	} else {
		width, height := configService.GetWindowSize()
		logger.Debug(fmt.Sprintf("Loaded window size from config, width: %d, height: %d", width, height))
		window.Resize(fyne.NewSize(float32(width), float32(height)))
	}

	return &App{fyneApp: fyneApp, window: window, configService: configService}
}

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
