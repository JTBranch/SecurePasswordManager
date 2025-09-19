package molecules

import (
	"context"
	"fmt"
	"go-password-manager/internal/service"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/transport"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ShareModalState string

const (
	ShareIdle        ShareModalState = "idle"
	ShareDiscovering ShareModalState = "discovering"
	ShareSelecting   ShareModalState = "selecting"
	ShareSending     ShareModalState = "sending"
	ShareSuccess     ShareModalState = "success"
	ShareError       ShareModalState = "error"
)

type ShareModalProps struct {
	Win     fyne.Window
	Service *service.SharingTransferService
	Secrets []sharing.ExportSecret // pre-selected for first iteration
	OnClose func()
}

// buildDevicesList creates the devices list widget and returns it with backing slice pointer
func buildDevicesList(deviceEntries *[]transport.DeviceDescriptor) *widget.List {
	return widget.NewList(
		func() int { return len(*deviceEntries) },
		func() fyne.CanvasObject { return widget.NewLabel("device") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i >= 0 && i < len(*deviceEntries) {
				d := (*deviceEntries)[i]
				o.(*widget.Label).SetText(d.DeviceName + " (" + d.DeviceID + ")")
			}
		},
	)
}

// buildSecretsList constructs a multi-select secrets list returning selection map reference
func buildSecretsList(secrets []sharing.ExportSecret) (*widget.List, []string, map[int]bool) {
	secretNames := make([]string, 0, len(secrets))
	for _, s := range secrets {
		secretNames = append(secretNames, s.Name)
	}
	selected := map[int]bool{}
	lst := widget.NewList(
		func() int { return len(secretNames) },
		func() fyne.CanvasObject { return widget.NewCheck("", nil) },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			chk := o.(*widget.Check)
			chk.Text = secretNames[i]
			chk.Checked = selected[i]
			chk.OnChanged = func(b bool) { selected[i] = b }
		},
	)
	return lst, secretNames, selected
}

func ShareModal(props ShareModalProps) {
	state := ShareIdle
	statusLabel := widget.NewLabel("")
	deviceEntries := []transport.DeviceDescriptor{}
	devicesList := buildDevicesList(&deviceEntries)

	secretsList, secretNames, selectedSecrets := buildSecretsList(props.Secrets)

	addrEntry := widget.NewEntry()
	addrEntry.SetPlaceHolder("host:port target (required)")
	sendBtn := widget.NewButton("Send", func() {
		if props.Service == nil {
			statusLabel.SetText("Send not wired: service nil")
			return
		}
		if addrEntry.Text == "" {
			statusLabel.SetText("Enter target host:port")
			return
		}
		// Collect selected secrets (names only placeholder)
		chosen := []string{}
		for idx, ok := range selectedSecrets {
			if ok {
				chosen = append(chosen, secretNames[idx])
			}
		}
		if len(chosen) == 0 && len(secretNames) > 0 { // default to first
			chosen = append(chosen, secretNames[0])
		}
		state = ShareSending
		statusLabel.SetText("Sending...")
		go func(chosen []string, addr string) {
			// Build placeholder bundle (minimal fields) until full export path wired
			payload := sharing.SecretExportPayload{ID: fmt.Sprintf("ui-%d", time.Now().UnixNano()), Timestamp: time.Now().Unix(), Name: "UI Share", SenderInfo: sharing.SenderMetadata{DeviceName: "UI"}}
			bundle := &sharing.SecretExportBundle{Payload: payload}
			target := transport.DeviceDescriptor{DeviceID: "manual-target", DeviceName: addr, LastAddr: addr}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			progCh, err := props.Service.SendBundle(ctx, "lan", bundle, target)
			if err != nil {
				statusLabel.SetText("Send error: " + err.Error())
				state = ShareError
				return
			}
			for p := range progCh {
				fyne.CurrentApp().SendNotification(&fyne.Notification{Title: "Share", Content: string(p.State)})
				if p.State == service.ShareFlowSucceeded {
					statusLabel.SetText("Send succeeded")
					state = ShareSuccess
				}
				if p.State == service.ShareFlowFailed {
					fyne.DoAndWait(func() {
						statusLabel.SetText("Send failed: " + fmt.Sprintf("%v", p.Error))
					})
					state = ShareError
				}
			}
		}(chosen, addrEntry.Text)
	})

	discoverBtn := widget.NewButton("Discover", func() {
		if state != ShareIdle && state != ShareSelecting {
			return
		}
		state = ShareDiscovering
		statusLabel.SetText("Discovering devices...")
		ctx, cancel := context.WithTimeout(context.Background(), 700*time.Millisecond)
		defer cancel()
		if props.Service == nil {
			state = ShareSelecting
			statusLabel.SetText("No service wired")
			return
		}
		devs, err := props.Service.DiscoverDevices(ctx, "lan", 25)
		if err != nil {
			state = ShareError
			statusLabel.SetText("Discovery error: " + err.Error())
			return
		}
		deviceEntries = devs
		devicesList.Refresh()
		state = ShareSelecting
		statusLabel.SetText(fmt.Sprintf("Discovered %d device(s)", len(devs)))
	})

	var pop *widget.PopUp
	closeBtn := widget.NewButton("Close", func() {
		if pop != nil {
			pop.Hide()
		}
		if props.OnClose != nil {
			props.OnClose()
		}
	})

	secretCount := len(props.Secrets)
	statusLabel.SetText(fmt.Sprintf("Ready to share %d secrets", secretCount))

	leftCol := container.NewVBox(widget.NewLabel("Secrets"), container.NewStack(secretsList))
	rightCol := container.NewVBox(widget.NewLabel("Devices"), container.NewStack(devicesList))
	topRow := container.NewHBox(discoverBtn, addrEntry, sendBtn, closeBtn)
	split := container.NewHSplit(leftCol, rightCol)
	split.SetOffset(0.5)

	content := container.NewBorder(
		container.NewVBox(widget.NewLabel("Share Secrets"), statusLabel, topRow),
		nil, nil, nil,
		split,
	)
	// Enlarge modal via a MinSize wrapper
	wrapper := container.NewStack(content)
	wrapper.Resize(fyne.NewSize(1500, 1000))
	pop = widget.NewModalPopUp(wrapper, props.Win.Canvas())
	pop.Show()
}

// fmtInt removed (inlined fmt.Sprintf usage to reduce symbols)
