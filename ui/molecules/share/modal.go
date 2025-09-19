package share

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Modal composes subcomponents and manages popup lifecycle.
func Modal(props Props) {
	secretNames := make([]string, 0, len(props.Secrets))
	for _, s := range props.Secrets {
		secretNames = append(secretNames, s.Name)
	}
	vm := NewViewModel(secretNames)
	devicesList := DevicesList(vm)
	secretsList := SecretsList(vm, props.Secrets)
	bar := ActionBar(props, vm, devicesList)
	status := widget.NewLabel(fmt.Sprintf("Ready to share %d secrets", len(props.Secrets)))
	header := container.NewVBox(widget.NewLabel("Share Secrets"), status, bar)
	cols := container.NewHSplit(
		container.NewVBox(widget.NewLabel("Secrets"), container.NewStack(secretsList)),
		container.NewVBox(widget.NewLabel("Devices"), container.NewStack(devicesList)),
	)
	cols.SetOffset(0.5)
	content := container.NewBorder(header, nil, nil, nil, cols)
	wrap := container.NewStack(content)
	wrap.Resize(fyne.NewSize(1400, 900))
	pop := widget.NewModalPopUp(wrap, props.Win.Canvas())
	pop.Show()
	// simple tick to refresh status if state changes
	go func() {
		prev := vm.State
		for pop.Visible() {
			if vm.State != prev {
				status.SetText(string(vm.State))
				prev = vm.State
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()
}
