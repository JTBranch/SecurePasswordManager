package molecules

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
	boldMsg := widget.NewLabelWithStyle(
		fmt.Sprintf("Are you sure you want to delete '%s'?", props.SecretName),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	warning := widget.NewLabelWithStyle(
		"This action cannot be undone. All versions of this secret will be permanently deleted.",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	bgColor := fyne.CurrentApp().Settings().Theme().Color("overlayBackground", fyne.CurrentApp().Settings().ThemeVariant())
	gap := canvas.NewRectangle(bgColor)
	gap.SetMinSize(fyne.NewSize(0, 24)) // 24px tall gap

	warning.Importance = widget.DangerImportance

	content := container.NewVBox(
		boldMsg,
		warning,
		gap, // Adds space below the warnings
	)

	paddedContent := container.NewPadded(content)

	dialog.NewCustomConfirm(
		"",
		"Delete",
		"Cancel",
		paddedContent,
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
