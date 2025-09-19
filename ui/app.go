package ui

import (
	"fmt"
	buildconfig "go-password-manager/internal/config/buildconfig"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/logger"
	"go-password-manager/internal/service"
	pages "go-password-manager/ui/pages"
	"go-password-manager/ui/themes"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// App represents the main application
type App struct {
	fyneApp        fyne.App
	window         fyne.Window
	configService  *config.ConfigService
	buildconfig    *buildconfig.Config
	secretsService *service.SecretsService
	transferSvc    *service.SharingTransferService
	currentPage    string
}

const (
	FALLBACK_WINDOW_WIDTH  = 750
	FALLBACK_WINDOW_HEIGHT = 1100
)

// NewApp creates a new application instance
func NewApp(buildCfg *buildconfig.Config, secretsService *service.SecretsService, transferSvc *service.SharingTransferService) *App {

	return NewAppWithFyne(app.New(), nil, buildCfg, secretsService, transferSvc)
}

// NewAppWithFyne creates a new application instance using the provided fyne.App and optional window.
// Pass a test app and window when running headless tests.
func NewAppWithFyne(fyneApp fyne.App, win fyne.Window, buildCfg *buildconfig.Config, secretsService *service.SecretsService, transferSvc *service.SharingTransferService) *App {
	if fyneApp == nil {
		fyneApp = app.New()
	}
	fyneApp.Settings().SetTheme(&themes.LightTheme{})
	var window fyne.Window
	if win == nil {
		window = fyneApp.NewWindow(buildCfg.Application.Name)
	} else {
		window = win
	}

	// Load legacy config service for window size persistence
	configService, err := config.NewConfigService(buildCfg)
	if err != nil {
		// Use environment config defaults
		window.Resize(fyne.NewSize(
			float32(FALLBACK_WINDOW_WIDTH),
			float32(FALLBACK_WINDOW_HEIGHT),
		))
	} else {
		width, height := configService.GetWindowSize()
		if buildCfg.Logging.Debug {
			logger.Debug(fmt.Sprintf("Loaded window size from config, width: %d, height: %d", width, height))
		}
		if width == 0 || height == 0 {
			window.Resize(fyne.NewSize(750, 1100))
		} else {
			window.Resize(fyne.NewSize(float32(width), float32(height)))
		}
		configTheme := configService.GetTheme()
		if configTheme != "" {
			fyneApp.Settings().SetTheme(themes.ThemeFromName(configTheme))
		}
	}

	return &App{fyneApp: fyneApp, window: window, configService: configService, buildconfig: buildCfg, secretsService: secretsService, transferSvc: transferSvc}
}

// Start shows the window and initializes the UI without blocking (useful for tests).
func (a *App) Start() {
	var showMain func()
	showShare := func() {
		a.currentPage = "share"
		a.window.SetContent(pages.SharePage(a.window, a.secretsService, a.transferSvc, func() { showMain() }))
	}
	showMain = func() {
		a.currentPage = "main"
		a.window.SetContent(pages.MainPageWithService(a.window, a.secretsService, a.transferSvc, a.configService, showShare))
	}
	showMain()

	// Save window size on close
	a.window.SetOnClosed(func() {
		if a.configService != nil {
			size := a.window.Canvas().Size()
			_ = a.configService.SetWindowSize(int(size.Width), int(size.Height))
		}
	})

	a.window.Show()
}

// Run starts the application
func (a *App) Run() {
	a.Start()
	a.window.ShowAndRun()
}
