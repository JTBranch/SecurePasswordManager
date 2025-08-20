package pages

import (
	buildconfig "go-password-manager/internal/config/buildConfig"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/service"
	"go-password-manager/internal/storage"
	"go-password-manager/ui/atoms"
	"go-password-manager/ui/molecules"
	"go-password-manager/ui/themes"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var defaultSecretsService *service.SecretsService

func init() {
	buildCfg, err := buildconfig.Load()
	if err != nil {
		log.Fatalf("Failed to load build config: %v", err)
	}
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
	storageService := storage.NewFileStorage(secretsPath, buildCfg.Application.Version, "default-user")
	defaultSecretsService = service.NewSecretsService(cryptoService, storageService)
}

func MainPage(win fyne.Window) fyne.CanvasObject {
	return MainPageWithService(win, defaultSecretsService)
}

func MainPageWithService(win fyne.Window, secretsService *service.SecretsService) fyne.CanvasObject {
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
		println("Theme changed to:", themeName)
		switch themeName {
		case "light":
			fyne.CurrentApp().Settings().SetTheme(&themes.LightTheme{})
		case "dark":
			fyne.CurrentApp().Settings().SetTheme(&themes.DarkTheme{})
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
