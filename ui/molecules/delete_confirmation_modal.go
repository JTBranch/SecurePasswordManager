package molecules

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// DeleteConfirmationModalProps holds the properties for the delete confirmation modal
type DeleteConfirmationModalProps struct {
	SecretName string
	OnConfirm  func()
	OnCancel   func()
}

// DeleteConfirmationModal creates a confirmation dialog for secret deletion
func DeleteConfirmationModal(window fyne.Window, props DeleteConfirmationModalProps) {
	message := fmt.Sprintf("Are you sure you want to delete '%s'?\n\nThis action cannot be undone. All versions of this secret will be permanently deleted.", props.SecretName)

	dialog.NewConfirm(
		"",
		message,
		func(confirm bool) {
			if confirm {
				if props.OnConfirm != nil {
					props.OnConfirm()
				}
			} else {
				if props.OnCancel != nil {
					props.OnCancel()
				}
			}
		},
		window,
	).Show()
}
