package pages

import (
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/logger"
	"go-password-manager/internal/service"
	"go-password-manager/ui/atoms"
	"go-password-manager/ui/molecules"
	"go-password-manager/ui/themes"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func MainPageWithService(win fyne.Window, secretsService *service.SecretsService, transferSvc *service.SharingTransferService, configService *config.ConfigService, onNavigateShare func(), onThemeApplied func()) fyne.CanvasObject {
	fileData, _ := secretsService.LoadAllSecrets()
	var selectedIdx int = -1
	listBox := container.NewVBox()
	detailBox := container.NewVBox(widget.NewLabel("Select a secret"))

	var updateList func()
	var updateDetail func()

	refreshDetail := func() {
		// Reload the data to get the latest version
		fileData, _ = secretsService.LoadAllSecrets()
		if selectedIdx >= 0 && selectedIdx < len(fileData.Secrets) {
			updateDetail()
		}
	}

	updateDetail = func() {
		detailBox.Objects = nil
		if selectedIdx >= 0 && selectedIdx < len(fileData.Secrets) {
			detailBox.Add(molecules.SecretDetail(fileData.Secrets[selectedIdx], secretsService, win, refreshDetail))
		} else {
			detailBox.Add(widget.NewLabel("Select a secret"))
		}
		detailBox.Refresh()
	}

	updateList = func() {
		fileData, _ = secretsService.LoadAllSecrets()
		listBox.Objects = nil
		for i, s := range fileData.Secrets {
			listBox.Add(atoms.SecretName(s, func(idx int) func() {
				return func() {
					selectedIdx = idx
					updateDetail()
				}
			}(i), func(secretName string) func() {
				return func() {
					// Show delete confirmation modal
					molecules.DeleteConfirmationModal(win, molecules.DeleteConfirmationModalProps{
						SecretName: secretName,
						OnConfirm: func() {
							_ = secretsService.DeleteSecret(secretName)
							selectedIdx = -1
							updateList()
							updateDetail()
						},
						OnCancel: func() {
							// Do nothing on cancel
						},
					})
				}
			}(s.SecretName)))
		}
		listBox.Refresh()
	}

	// --- AppHeader logic moved to component ---
	filterAndRender := func(query string) {
		fileData, _ = secretsService.LoadAllSecrets()
		listBox.Objects = nil
		for i, s := range fileData.Secrets {
			if query == "" || containsIgnoreCase(s.SecretName, query) {
				listBox.Add(atoms.SecretName(s, func(idx int) func() { return func() { selectedIdx = idx; updateDetail() } }(i), func(secretName string) func() {
					return func() {
						molecules.DeleteConfirmationModal(win, molecules.DeleteConfirmationModalProps{SecretName: secretName, OnConfirm: func() { _ = secretsService.DeleteSecret(secretName); selectedIdx = -1; updateList(); updateDetail() }})
					}
				}(s.SecretName)))
			}
		}
		listBox.Refresh()
	}
	props := molecules.AppHeaderProps{
		OnSearch: filterAndRender,
		OnCreateSecret: func() {
			molecules.NewSecretModal(win, secretsService, func() {
				updateList()
			})
		},
		OnMenuAction: func() {
			if onNavigateShare != nil {
				onNavigateShare()
			}
		},
		OnExport: func() {
			if onNavigateShare != nil {
				onNavigateShare()
			}
		},
	}
	props.OnThemeChange = func(themeName string) {
		logger.Debug("Theme changed to:", themeName)
		switch themeName {
		case "light":
			fyne.CurrentApp().Settings().SetTheme(&themes.LightTheme{})
			configService.SetTheme(themeName)

		case "dark":
			fyne.CurrentApp().Settings().SetTheme(&themes.DarkTheme{})
			configService.SetTheme(themeName)
		}
		// keep application icon constant; UI uses light/dark PNGs for header only
		// Wire theme-applied to navigate back to main if caller provided a handler
		props.OnThemeApplied = func() {
			if onThemeApplied != nil {
				onThemeApplied()
			}
		}
		// After applying a theme, navigate back to the main page by invoking onNavigateShare's inverse.
		// If onNavigateShare is provided, simply ensure we're on the main page by doing nothing; callers
		// that want explicit navigation can wrap this handler. The app wiring uses a showMain/ showShare
		// closure so applying a theme should not unexpectedly navigate. For safety, if a caller exposed
		// a navigation callback via props.OnMenuAction, prefer calling it to surface UI changes.
		if onNavigateShare != nil {
			// no-op here; the app has direct control to navigate. keep behaviour minimal.
		}
	}

	// Set header starting theme from config service and build header
	if configService != nil {
		props.CurrentTheme = configService.GetTheme()
	}
	header := molecules.AppHeader(props, win)
	// register the create button from the header so tests can find it deterministically
	if c := findCreateButtonInHeader(header); c != nil {
		registerCreateButton(c)
	}
	updateList()

	split := container.NewHSplit(listBox, detailBox)
	split.SetOffset(0.3) // This sets the split ratio, not a fixed size

	content := container.NewBorder(
		header, // top
		nil,    // bottom
		nil,    // left
		nil,    // right
		container.NewHSplit(listBox, detailBox),
	)
	return content
}

// test registration for create button
var registeredCreateButton *widget.Button

func registerCreateButton(b *widget.Button) {
	registeredCreateButton = b
}

func GetRegisteredCreateButton() *widget.Button {
	return registeredCreateButton
}

// findCreateButtonInHeader attempts to locate the create button inside header container
func findCreateButtonInHeader(header fyne.CanvasObject) *widget.Button {
	if header == nil {
		return nil
	}
	var found *widget.Button
	// local traversal helper
	var traverseLocal func(fyne.CanvasObject, func(fyne.CanvasObject))
	traverseLocal = func(obj fyne.CanvasObject, fn func(fyne.CanvasObject)) {
		if obj == nil {
			return
		}
		fn(obj)
		type hasObjects interface{ Objects() []fyne.CanvasObject }
		if c, ok := obj.(hasObjects); ok {
			for _, child := range c.Objects() {
				traverseLocal(child, fn)
			}
		}
	}

	traverseLocal(header, func(o fyne.CanvasObject) {
		if found != nil {
			return
		}
		if b, ok := o.(*widget.Button); ok {
			if containsIgnoreCase(b.Text, "Create") || containsIgnoreCase(b.Text, "New") {
				found = b
			}
		}
	})
	return found
}

// Helper for case-insensitive substring search
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
