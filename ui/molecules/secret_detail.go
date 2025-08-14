package molecules

import (
	"go-password-manager/internal/domain"
	"go-password-manager/internal/service"
	"go-password-manager/ui/atoms"
	"go-password-manager/ui/helpers"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func SecretDetail(secret domain.Secret, secretsService *service.SecretsService, window fyne.Window, onUpdate func()) fyne.CanvasObject {

	revealed := false
	editMode := false
	var currentPlainValue string

	// Secret name label - full width at the top
	nameLabel := widget.NewLabel(secret.SecretName)
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Secret value entry for edit mode
	valueEntry := widget.NewEntry()
	valueEntry.Hide() // Initially hidden

	// Edit/Save button
	editBtn := widget.NewButton("✏️", nil)

	// Header with name and edit button
	header := container.NewBorder(nil, nil, nil, editBtn, nameLabel)

	// Create a container that will hold the main secret value atom
	mainValueContainer := container.NewVBox()

	// Function to update the main value display
	var updateMainValueDisplay func()
	updateMainValueDisplay = func() {
		mainValueContainer.Objects = nil
		secretValueAtom := atoms.SecretValue(atoms.SecretValueProps{
			Value:      currentPlainValue,
			IsRevealed: revealed,
			OnRevealClick: func() {
				if !editMode { // Only allow reveal/hide when not in edit mode
					revealed = !revealed
					if revealed {
						plain, err := secretsService.DisplaySecret(secret)
						if err == nil {
							currentPlainValue = plain
						}
					}
					updateMainValueDisplay()
				}
			},
			OnValueClick: func() {
				if revealed && currentPlainValue != "" {
					helpers.CopyToClipboard(currentPlainValue, window)
				}
			},
		})
		mainValueContainer.Objects = append(mainValueContainer.Objects, secretValueAtom)
		mainValueContainer.Refresh()
	}

	// Initial display
	updateMainValueDisplay()

	// Set the edit button callback
	editBtn.OnTapped = func() {
		if !editMode {
			// Enter edit mode
			editMode = true
			editBtn.SetText("💾") // Save icon

			// Get current value and show in entry
			plain, err := secretsService.DisplaySecret(secret)
			if err == nil {
				valueEntry.SetText(plain)
				currentPlainValue = plain
			}

			// Hide main value container, show entry
			mainValueContainer.Hide()
			valueEntry.Show()
		} else {
			// Save mode
			newValue := valueEntry.Text
			if newValue != "" {
				// Update the secret using EditSecret method
				err := secretsService.EditSecret(secret.SecretName, newValue)
				if err == nil {
					// Exit edit mode
					editMode = false
					revealed = false
					editBtn.SetText("✏️") // Edit icon
					currentPlainValue = ""

					// Hide entry, show main value container
					valueEntry.Hide()
					updateMainValueDisplay() // Refresh to reset state
					mainValueContainer.Show()

					// Trigger parent update to refresh with new data
					if onUpdate != nil {
						onUpdate()
					}
				}
			}
		}
	}

	// Value row with main value container and entry (stacked)
	valueRow := container.NewStack(mainValueContainer, valueEntry)

	// History component
	historyComponent := SecretHistory(secret.SecretName, secretsService, window)

	return container.NewVBox(header, valueRow, historyComponent)
}
