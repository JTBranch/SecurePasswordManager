package ui

import (
	"fmt"
	"go-password-manager/internal/logger"
	"go-password-manager/internal/service"
	"go-password-manager/internal/domain"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var secretsService = service.NewSecretsService("1.0.0", "jack.branch")

// CreateMainContent creates the main content of the window, including the secret form and secret list
func CreateMainContent() fyne.CanvasObject {
    logger.Debug("Creating main content for the UI")
    fileData, _ := secretsService.LoadAllSecrets()
    var selectedIdx int = -1

    listBox := container.NewVBox()
    detailBox := container.NewVBox(widget.NewLabel("Select a secret"))

    // Declare updateList before using it in closures
    var updateList func()

    // Helper to update the detail panel
    updateDetail := func() {
        detailBox.Objects = nil
        if selectedIdx >= 0 && selectedIdx < len(fileData.Secrets) {
            detailBox.Add(SecretDetail(fileData.Secrets[selectedIdx]))
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
	delBtn := widget.NewButton("🗑", onDelete)
	return container.NewHBox(nameBtn, delBtn)
}

// SecretDetail renders the selected secret with hide/show functionality
func SecretDetail(secret domain.Secret) fyne.CanvasObject {
	revealed := false
	label := widget.NewLabel(fmt.Sprintf("%s: ******* [%s]", secret.SecretName, secret.Type))
	revealBtn := widget.NewButton("👁", func() {
		revealed = !revealed
		if revealed {
			plain, err := secretsService.DisplaySecret(secret)
			if err == nil {
				label.SetText(fmt.Sprintf("%s: %s [%s]", secret.SecretName, plain, secret.Type))
			}
		} else {
			label.SetText(fmt.Sprintf("%s: ******* [%s]", secret.SecretName, secret.Type))
		}
	})
	return container.NewVBox(label, revealBtn)
}
