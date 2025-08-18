package molecules

import (
	"go-password-manager/internal/service"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func NewSecretModal(win fyne.Window, secretsService *service.SecretsService, onSuccess func()) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Secret name")
	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("Secret value")

	formItems := []*widget.FormItem{
		widget.NewFormItem("Name", nameEntry),
		widget.NewFormItem("Value", valueEntry),
	}

	dialog.ShowForm("Create Secret", "Create", "Cancel", formItems, func(ok bool) {
		if ok {
			name := nameEntry.Text
			value := valueEntry.Text
			if name != "" && value != "" {
				_ = secretsService.SaveNewSecret(name, value)
				onSuccess()
			}
		}
	}, win)
}
