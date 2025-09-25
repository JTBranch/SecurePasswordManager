package share

import (
	"fmt"
	"time"

	"go-password-manager/internal/transport"

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
	bar, addrEntry := ActionBar(props, vm, devicesList)
	status := widget.NewLabel(fmt.Sprintf("Ready to share %d secrets", len(props.Secrets)))

	// wire device actions: connect -> populate addrEntry, details -> show modal
	vm.OnDeviceAction = func(idx int, action string) {
		if idx < 0 || idx >= len(vm.Devices) {
			return
		}
		d := vm.Devices[idx]
		switch action {
		case "connect":
			if d.LastAddr != "" {
				addrEntry.SetText(d.LastAddr)
				props.Win.Canvas().Focus(addrEntry)
			}
		case "details":
			DeviceDetailsModal(props.Win, d)
		}
	}

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

// DeviceDetailsModal shows a simple read-only view of a DeviceDescriptor.
func DeviceDetailsModal(win fyne.Window, d transport.DeviceDescriptor) {
	title := widget.NewLabel("Device Details")
	name := widget.NewLabel("Name: " + d.DeviceName)
	id := widget.NewLabel("ID: " + d.DeviceID)
	addr := widget.NewLabel("Last Address: " + d.LastAddr)
	// adv/raw fields are optional; show timestamp
	seen := widget.NewLabel("Last Seen: " + d.LastSeenAt.String())
	closeBtn := widget.NewButton("Close", func() { /* handler assigned after popup created */ })

	content := container.NewVBox(title, name, id, addr, seen, closeBtn)
	wrap := container.NewStack(content)
	wrap.Resize(fyne.NewSize(400, 300))
	pop := widget.NewModalPopUp(wrap, win.Canvas())
	closeBtn.OnTapped = func() { pop.Hide() }
	pop.Show()
}
