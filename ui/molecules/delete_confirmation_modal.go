package molecules

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// DeleteConfirmationModalProps holds the properties for the delete confirmation modal
type DeleteConfirmationModalProps struct {
	SecretName string
	OnConfirm  func()
	OnCancel   func()
}

// DeleteConfirmationModal creates a confirmation dialog for secret deletion
func DeleteConfirmationModal(window fyne.Window, props DeleteConfirmationModalProps) {
	// Create the warning message
	warningIcon := widget.NewIcon(nil)
	warningIcon.SetResource(fyne.CurrentApp().Icon())

	titleLabel := widget.NewLabel("Delete Secret")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	messageLabel := widget.NewLabel(fmt.Sprintf("Are you sure you want to delete '%s'?", props.SecretName))
	messageLabel.Wrapping = fyne.TextWrapWord

	warningLabel := widget.NewLabel("This action cannot be undone. All versions of this secret will be permanently deleted.")
	warningLabel.TextStyle = fyne.TextStyle{Italic: true}
	warningLabel.Wrapping = fyne.TextWrapWord

	// Layout the content
	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		messageLabel,
		widget.NewSeparator(),
		warningLabel,
		widget.NewSeparator(),
	)

	// Create the dialog first so we can reference it in button callbacks
	modal := dialog.NewCustom("", "", content, window)

	// Create buttons with modal reference
	cancelBtn := widget.NewButton("Cancel", func() {
		modal.Hide()
		if props.OnCancel != nil {
			props.OnCancel()
		}
	})

	deleteBtn := widget.NewButton("Delete", func() {
		modal.Hide()
		if props.OnConfirm != nil {
			props.OnConfirm()
		}
	})
	deleteBtn.Importance = widget.DangerImportance

	// Add buttons to the content
	buttonRow := container.NewBorder(nil, nil, cancelBtn, deleteBtn, widget.NewLabel(""))
	content.Add(buttonRow)

	// Make the modal wider and show it
	modal.Resize(fyne.NewSize(450, 200))
	modal.SetOnClosed(func() {
		if props.OnCancel != nil {
			props.OnCancel()
		}
	})

	modal.Show()
}
