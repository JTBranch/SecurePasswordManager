package ui

import (
	"fmt"
	buildconfig "go-password-manager/internal/config/buildconfig"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/logger"
	"go-password-manager/internal/service"
	pages "go-password-manager/ui/pages"
	"go-password-manager/ui/themes"
	"os"
	"path/filepath"
	"runtime"

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
	// load application icon if available; prefer platform-specific formats
	var iconBytes []byte
	var iconName string
	findAsset := func(name string) ([]byte, error) {
		// try relative path
		p := filepath.Join("ui", "assets", name)
		if b, err := os.ReadFile(p); err == nil {
			return b, nil
		}
		// try executable dir (app bundle resources)
		if ex, err := os.Executable(); err == nil {
			dir := filepath.Dir(ex)
			tryPaths := []string{
				filepath.Join(dir, name),
				filepath.Join(dir, "Contents", "Resources", name),
				filepath.Join(dir, "ui", "assets", name),
			}
			for _, tp := range tryPaths {
				if b, err := os.ReadFile(tp); err == nil {
					return b, nil
				}
			}
		}
		return os.ReadFile(name)
	}

	// prefer ICO for Windows/Linux; prefer ICNS for macOS if provided during packaging
	if runtime.GOOS == "darwin" {
		if b, err := findAsset("main-icon.icns"); err == nil {
			iconBytes = b
			iconName = "main-icon.icns"
		}
	}
	if iconBytes == nil {
		if b, err := findAsset("main-icon.ico"); err == nil {
			iconBytes = b
			iconName = "main-icon.ico"
		} else if b, err := findAsset("main-icon.png"); err == nil {
			iconBytes = b
			iconName = "main-icon.png"
		}
	}
	// Fyne expects image data (PNG/ICO). Prefer using the PNG variant at runtime so
	// the window and app icon display consistently across platforms. Keep ICNS for
	// packaging (macOS bundles) but don't pass raw .icns bytes to Fyne.
	if b, err := findAsset("main-icon.png"); err == nil {
		res := fyne.NewStaticResource("main-icon.png", b)
		fyneApp.SetIcon(res)
	} else if iconBytes != nil {
		res := fyne.NewStaticResource(iconName, iconBytes)
		fyneApp.SetIcon(res)
	}
	fyneApp.Settings().SetTheme(&themes.LightTheme{})
	var window fyne.Window
	if win == nil {
		window = fyneApp.NewWindow(buildCfg.Application.Name)
		// set window icon to app icon if available
		if ic := fyneApp.Icon(); ic != nil {
			window.SetIcon(ic)
		}
	} else {
		window = win
	}

	// Load legacy config service for window size persistence
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
		a.window.SetContent(pages.MainPageWithService(a.window, a.secretsService, a.transferSvc, a.configService, showShare, showMain))
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
