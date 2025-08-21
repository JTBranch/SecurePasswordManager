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

func MainPageWithService(win fyne.Window, secretsService *service.SecretsService, configService *config.ConfigService) fyne.CanvasObject {
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
	props := molecules.AppHeaderProps{
		OnSearch: func(query string) {
			fileData, _ = secretsService.LoadAllSecrets()
			listBox.Objects = nil
			for i, s := range fileData.Secrets {
				if query == "" || containsIgnoreCase(s.SecretName, query) {
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
			}
			listBox.Refresh()
		},
		OnCreateSecret: func() {
			molecules.NewSecretModal(win, secretsService, func() {
				updateList()
			})
		},
		OnMenuAction: func() {
			// TODO: Implement menu functionality
			// This will be used for importing secrets from browser, etc.
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
	}

	header := molecules.AppHeader(props, win)
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

// Helper for case-insensitive substring search
func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
