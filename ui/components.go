package ui

import (
	buildconfig "go-password-manager/internal/config/buildconfig"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/logger"
	"go-password-manager/internal/service"
	"go-password-manager/internal/storage"
	"go-password-manager/ui/molecules"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var secretsService *service.SecretsService

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
	secretsService = service.NewSecretsService(cryptoService, storageService)
}

// CreateMainContent creates the main content of the window, including the secret form and secret list
func CreateMainContent(window fyne.Window) fyne.CanvasObject {
	logger.Debug("Creating main content for the UI")
	fileData, _ := secretsService.LoadAllSecrets()
	var selectedIdx int = -1

	listBox := container.NewVBox()
	detailBox := container.NewVBox(widget.NewLabel("Select a secret"))

	// Declare updateList before using it in closures
	var updateDetail func()
	var updateList func()

	refreshDetail := func() {
		// Reload the data to get the latest version
		fileData, _ = secretsService.LoadAllSecrets()
		if selectedIdx >= 0 && selectedIdx < len(fileData.Secrets) {
			updateDetail()
		}
	}

	// Helper to update the detail panel
	updateDetail = func() {
		detailBox.Objects = nil
		if selectedIdx >= 0 && selectedIdx < len(fileData.Secrets) {
			detailBox.Add(molecules.SecretDetail(fileData.Secrets[selectedIdx], secretsService, window, refreshDetail))
		} else {
			detailBox.Add(widget.NewLabel("Select a secret"))
		}
		detailBox.Refresh()
	}

	// Helper to update the secret list
	updateList = func() {
		fileData, _ = secretsService.LoadAllSecrets()
		listBox.Objects = nil
		for i, s := range fileData.Secrets {
			listBox.Add(SecretName(s, func(idx int) func() {
				return func() {
					selectedIdx = idx
					updateDetail()
				}
			}(i), func() {
				_ = secretsService.DeleteSecret(s.SecretName)
				selectedIdx = -1
				updateList()
				updateDetail()
			}))
		}
		listBox.Refresh()
	}

	updateList() // Initial population

	content := container.NewHSplit(
		listBox,
		detailBox,
	)
	content.SetOffset(0.3)
	return content
}

// SecretList renders all secret names and handles selection/deletion
func SecretList(secrets []domain.Secret, onSelect func(idx int), onDelete func(name string)) fyne.CanvasObject {
	list := container.NewVBox()
	for i, s := range secrets {
		list.Add(SecretName(s, func() { onSelect(i) }, func() { onDelete(s.SecretName) }))
	}
	return list
}

// SecretName renders a single secret name with delete icon
func SecretName(secret domain.Secret, onClick func(), onDelete func()) fyne.CanvasObject {
	nameBtn := widget.NewButton(secret.SecretName, onClick)
	return container.NewHBox(nameBtn)
}
