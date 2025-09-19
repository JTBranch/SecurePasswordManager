package pages

import (
	"fmt"
	"go-password-manager/internal/service"
	"go-password-manager/internal/sharing"
	"go-password-manager/ui/molecules/share"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// SharePage renders a full-page sharing workspace replacing the previous modal.
func SharePage(win fyne.Window, secretsService *service.SecretsService, transferSvc *service.SharingTransferService, onBack func()) fyne.CanvasObject {
	// collect secret names (minimal for now; could map service layer types to sharing.ExportSecret later)
	fileData, _ := secretsService.LoadAllSecrets()
	secrets := make([]sharing.ExportSecret, 0, len(fileData.Secrets))
	for _, s := range fileData.Secrets {
		secrets = append(secrets, sharing.ExportSecret{Name: s.SecretName})
	}
	props := share.Props{Win: win, Service: transferSvc, Secrets: secrets, OnClose: func() {
		if onBack != nil {
			onBack()
		}
	}}
	// build composing components similar to modal but using full window layout
	secretNames := make([]string, 0, len(props.Secrets))
	for _, s := range props.Secrets {
		secretNames = append(secretNames, s.Name)
	}
	vm := share.NewViewModel(secretNames)
	devicesList := share.DevicesList(vm)
	secretsList := share.SecretsList(vm, props.Secrets)
	bar := share.ActionBar(props, vm, devicesList)
	status := widget.NewLabel(fmt.Sprintf("Ready to share %d secrets", len(props.Secrets)))
	backBtn := widget.NewButton("Back", func() {
		if onBack != nil {
			onBack()
		}
	})
	header := container.NewHBox(backBtn, widget.NewLabelWithStyle("Share Secrets", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), status)
	cols := container.NewHSplit(
		container.NewBorder(widget.NewLabel("Secrets"), nil, nil, nil, secretsList),
		container.NewBorder(widget.NewLabel("Devices"), nil, nil, nil, devicesList),
	)
	cols.SetOffset(0.5)
	content := container.NewBorder(container.NewVBox(header, bar), nil, nil, nil, cols)
	// background refresh loop for status updates
	go func() {
		prev := vm.State
		for {
			if vm.State != prev {
				fyne.DoAndWait(func() {
					status.SetText(string(vm.State))
					prev = vm.State
				})
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()
	return content
}
