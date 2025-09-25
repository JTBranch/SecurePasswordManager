package share

import (
	"fmt"
	"go-password-manager/internal/transport"
	clip "go-password-manager/ui/helpers"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// DevicesList renders the list of discovered devices with per-item actions.
func DevicesList(vm *ViewModel) *widget.List {
	return widget.NewList(
		func() int { return len(vm.Devices) },
		func() fyne.CanvasObject {
			chk := widget.NewCheck("", nil)
			title := widget.NewLabel("")
			subtitle := widget.NewLabel("")
			subtitle.TextStyle = fyne.TextStyle{Italic: true}
			addrLabel := widget.NewLabel("")

			copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), nil)
			connectBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), nil)
			detailsBtn := widget.NewButtonWithIcon("", theme.InfoIcon(), nil)

			actions := container.NewHBox(copyBtn, connectBtn, detailsBtn)
			info := container.NewVBox(title, subtitle)
			rightCol := container.NewVBox(addrLabel, actions)
			row := container.New(layout.NewHBoxLayout(), chk, info, layout.NewSpacer(), rightCol)
			return row
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < 0 || i >= len(vm.Devices) {
				return
			}
			d := vm.Devices[i]
			box := o.(*fyne.Container)
			chk, _ := box.Objects[0].(*widget.Check)
			info, _ := box.Objects[1].(*fyne.Container)
			rightCol, _ := box.Objects[3].(*fyne.Container)

			title, _ := info.Objects[0].(*widget.Label)
			subtitle, _ := info.Objects[1].(*widget.Label)
			addrLabel, _ := rightCol.Objects[0].(*widget.Label)
			actions, _ := rightCol.Objects[1].(*fyne.Container)

			copyBtn, _ := actions.Objects[0].(*widget.Button)
			connectBtn, _ := actions.Objects[1].(*widget.Button)
			detailsBtn, _ := actions.Objects[2].(*widget.Button)

			title.SetText(d.DeviceName)
			subtitleText := d.DeviceID
			if d.LastAddr != "" {
				subtitleText = fmt.Sprintf("%s — %s", d.DeviceID, d.LastAddr)
			}
			if d.LastAddr == "" && !strings.Contains(d.DeviceName, "[fallback]") {
				title.SetText(d.DeviceName + " [fallback]")
			}
			subtitle.SetText(subtitleText)

			if d.LastAddr != "" {
				addrLabel.SetText(d.LastAddr)
			} else {
				addrLabel.SetText("")
			}

			idx := i
			copyBtn.OnTapped = func() {
				if d.LastAddr != "" {
					clip.CopyToClipboard(d.LastAddr, nil)
				}
			}
			connectBtn.OnTapped = func() {
				if vm.OnDeviceAction != nil {
					vm.OnDeviceAction(idx, "connect")
				}
			}
			detailsBtn.OnTapped = func() {
				if vm.OnDeviceAction != nil {
					vm.OnDeviceAction(idx, "details")
				}
			}

			// guard map access
			if vm.SelectedDevice == nil {
				vm.SelectedDevice = map[int]bool{}
			}
			chk.SetChecked(vm.SelectedDevice[i])
			chk.OnChanged = func(b bool) { vm.SelectedDevice[idx] = b }
		},
	)
}

// SetDevices updates the view model devices and refreshes the list on the UI thread.
func SetDevices(vm *ViewModel, devices []transport.DeviceDescriptor, list *widget.List) {
	vm.Devices = devices
	newSel := map[int]bool{}
	for i := range devices {
		if vm.SelectedDevice[i] {
			newSel[i] = true
		}
	}
	vm.SelectedDevice = newSel
	go func() {
		fyne.DoAndWait(func() {
			list.Refresh()
		})
	}()
}
