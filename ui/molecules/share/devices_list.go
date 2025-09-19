package share

import (
	"go-password-manager/internal/transport"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func DevicesList(vm *ViewModel) *widget.List {
	return widget.NewList(
		func() int { return len(vm.Devices) },
		func() fyne.CanvasObject {
			chk := widget.NewCheck("device", nil)
			box := container.NewHBox(chk)
			return box
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(vm.Devices) {
				return
			}
			d := vm.Devices[i]
			// first child is Check
			box := o.(*fyne.Container)
			chk, _ := box.Objects[0].(*widget.Check)
			label := d.DeviceName + " (" + d.DeviceID + ")"
			if d.LastAddr != "" {
				label += " - " + d.LastAddr
			}
			if d.LastAddr == "" && !strings.Contains(label, "[fallback]") {
				label += " [fallback]"
			}
			chk.SetText(label)
			chk.SetChecked(vm.SelectedDevice[i])
			idx := i
			chk.OnChanged = func(b bool) { vm.SelectedDevice[idx] = b }
		},
	)
}

func SetDevices(vm *ViewModel, devices []transport.DeviceDescriptor, list *widget.List) {
	vm.Devices = devices
	// ensure selection map sized; preserve existing flags where possible
	newSel := map[int]bool{}
	for i := range devices {
		if vm.SelectedDevice[i] {
			newSel[i] = true
		}
	}
	vm.SelectedDevice = newSel
	fyne.DoAndWait(func() {
		list.Refresh()
	})
}
