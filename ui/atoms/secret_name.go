package atoms

import (
	"go-password-manager/internal/domain"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SecretName creates a UI component for displaying a secret name with actions
func SecretName(secret domain.Secret, onClick func(), onDelete func()) fyne.CanvasObject {
	nameBtn := widget.NewButton(secret.SecretName, onClick)
	deleteBtn := widget.NewButtonWithIcon("Delete", theme.CancelIcon(), onDelete)
	deleteBtn.Importance = widget.DangerImportance

	// Make name button expand to fill available space
	nameBtn.Resize(fyne.NewSize(300, nameBtn.MinSize().Height)) // Set a larger width

	return container.NewBorder(nil, nil, nil, deleteBtn, nameBtn)
}
