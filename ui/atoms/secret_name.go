package atoms

import (
	"go-password-manager/internal/domain"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func SecretName(secret domain.Secret, onClick func(), onDelete func()) fyne.CanvasObject {
	nameBtn := widget.NewButton(secret.SecretName, onClick)
	delBtn := widget.NewButton("🗑", onDelete)

	// Make name button expand to fill available space
	nameBtn.Resize(fyne.NewSize(300, nameBtn.MinSize().Height)) // Set a larger width

	return container.NewBorder(nil, nil, nil, delBtn, nameBtn)
}
