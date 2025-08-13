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
    return container.NewHBox(nameBtn, delBtn)
}