package molecules

import (
	"go-password-manager/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func NewSecretModal(win fyne.Window, secretsService *service.SecretsService, onSuccess func()) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Secret name")
	nameRow := container.NewGridWrap(fyne.NewSize(500, nameEntry.MinSize().Height), nameEntry)

	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("Secret value")
	valueRow := container.NewGridWrap(fyne.NewSize(500, valueEntry.MinSize().Height), valueEntry)

	form := widget.NewForm(
		widget.NewFormItem("Name", nameRow),
		widget.NewFormItem("Value", valueRow),
	)

	bgColor := fyne.CurrentApp().Settings().Theme().Color("overlayBackground", fyne.CurrentApp().Settings().ThemeVariant())
	// Rectangle background for the form area
	paddedForm := container.NewPadded(form)

	// Add a gap below the form, above the buttons
	gap := canvas.NewRectangle(bgColor)
	gap.SetMinSize(fyne.NewSize(0, 24)) // 24px tall gap

	content := container.NewVBox(
		paddedForm,
		gap,
	)

	dialog.NewCustomConfirm(
		"Create Secret",
		"Create",
		"Cancel",
		content,
		func(ok bool) {
			if ok {
				name := nameEntry.Text
				value := valueEntry.Text
				if name != "" && value != "" {
					_ = secretsService.SaveNewSecret(name, value)
					onSuccess()
				}
			}
		},
		win,
	).Show()
}
