package ui

import (
	"go-password-manager/internal/domain"
	"go-password-manager/internal/logger"
	"go-password-manager/internal/service"
	"go-password-manager/ui/molecules"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var secretsService = service.NewSecretsService("1.0.0", "jack.branch")

// CreateMainContent creates the main content of the window, including the secret form and secret list
func CreateMainContent(window fyne.Window) fyne.CanvasObject {
	logger.Debug("Creating main content for the UI")
	fileData, _ := secretsService.LoadLatestSecrets()
	var selectedIdx int = -1

	listBox := container.NewVBox()
	detailBox := container.NewVBox(widget.NewLabel("Select a secret"))

	// Declare updateList before using it in closures
	var updateDetail func()
	var updateList func()

	refreshDetail := func() {
		// Reload the data to get the latest version
		fileData, _ = secretsService.LoadLatestSecrets()
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
		fileData, _ = secretsService.LoadLatestSecrets()
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
	delBtn := widget.NewButton("🗑", onDelete)
	return container.NewHBox(nameBtn, delBtn)
}
