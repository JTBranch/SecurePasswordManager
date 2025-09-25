package share

import (
	"context"
	"fmt"
	"go-password-manager/internal/logger"
	"go-password-manager/internal/service"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/transport"
	"go-password-manager/ui/molecules"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ActionCallbacks struct {
	OnStatus   func(string)
	OnState    func(ShareModalState)
	OnSent     func()
	OnError    func(error)
	OnDiscover func([]transport.DeviceDescriptor, error)
}

func ActionBar(props Props, vm *ViewModel, devicesList *widget.List) (*fyne.Container, *widget.Entry) {
	addrEntry := widget.NewEntry()
	addrEntry.SetPlaceHolder("host:port (e.g. 192.168.1.42:8080)")
	// transportSelect chooses which transport to use for discover/send.
	// Hide Bluetooth option on Windows since we don't support a Windows adapter.
	options := []string{"lan", "bluetooth"}
	if runtime.GOOS == "windows" {
		options = []string{"lan"}
	}
	// We don't need to react immediately when user toggles it; selection is read by actions.
	transportSelect := widget.NewSelect(options, func(_ string) {
		// intentionally no-op; selection read on action invocation
	})
	transportSelect.SetSelected("lan")
	spinner := widget.NewProgressBarInfinite()
	spinner.Hide()
	sendBtn := widget.NewButton("Send", func() {
		transportID := transportSelect.Selected
		if transportID == "" {
			transportID = "lan"
		}
		addr := addrEntry.Text
		if addr == "" { // try selected device
			for i, sel := range vm.SelectedDevice {
				if sel && i < len(vm.Devices) && vm.Devices[i].LastAddr != "" {
					addr = vm.Devices[i].LastAddr
					break
				}
			}
		}
		if !validateInputs(props, addr) {
			return
		}
		runSendAsync(props, vm, transportID, addr, collectSecrets(vm))
	})

	// Inline spinner next to Discover button for compactness
	discoverBtn := widget.NewButton("Discover", func() {
		spinner.Show()
		transportID := transportSelect.Selected
		if transportID == "" {
			transportID = "lan"
		}
		runDiscover(props, vm, devicesList, spinner, transportID)
	})
	fallbackChk := widget.NewCheck("Show fallback", func(b bool) {
		vm.ShowFallback = b
		// Reapply filter to last discovery without triggering new discovery
		if len(vm.LastDiscovery) > 0 {
			filtered := filterFallback(vm.LastDiscovery, vm.ShowFallback)
			SetDevices(vm, filtered, devicesList)
		}
	})
	fallbackChk.SetChecked(vm.ShowFallback)
	closeBtn := widget.NewButton("Close", func() {
		if props.OnClose != nil {
			props.OnClose()
		}
	})
	help := widget.NewLabel("Discover scans LAN or Bluetooth. Or manually enter host:port (e.g. 192.168.1.42:8080)")
	help.Wrapping = fyne.TextWrapWord
	buttons := container.NewHBox(sendBtn, closeBtn)
	// place transport selector above address entry; show discover + spinner inline
	discoverRow := container.NewHBox(discoverBtn, spinner)
	leftControls := container.NewVBox(discoverRow, transportSelect, fallbackChk)
	row := container.NewBorder(nil, nil, leftControls, buttons, addrEntry)
	return container.NewVBox(row, help), addrEntry
}

// Helpers extracted to reduce cognitive complexity.
func validateInputs(props Props, addr string) bool {
	if props.Service == nil {
		if props.Win != nil {
			molecules.ErrorModal(props.Win, molecules.ErrorModalProps{Title: "Share", Message: "Service not available"})
		}
		return false
	}
	if addr == "" {
		if props.Win != nil {
			molecules.ErrorModal(props.Win, molecules.ErrorModalProps{Title: "Send", Message: "Please enter a host:port"})
		}
		return false
	}
	return true
}
func collectSecrets(vm *ViewModel) []string {
	chosen := make([]string, 0, len(vm.SelectedSecret))
	for i, sel := range vm.SelectedSecret {
		if sel {
			chosen = append(chosen, vm.SecretNames[i])
		}
	}
	if len(chosen) == 0 && len(vm.SecretNames) > 0 {
		chosen = append(chosen, vm.SecretNames[0])
	}
	return chosen
}
func runSendAsync(props Props, vm *ViewModel, transportID string, addr string, picked []string) {
	vm.State = ShareSending
	go func() { executeSend(props, vm, transportID, addr, picked) }()
}

func executeSend(props Props, vm *ViewModel, transportID string, addr string, picked []string) {
	bundle := buildExportBundle(picked)
	target := transport.DeviceDescriptor{DeviceID: "manual-target", DeviceName: addr, LastAddr: addr}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	progCh, err := props.Service.SendBundle(ctx, transportID, bundle, target)
	if err != nil {
		showErr(props, "Send", err.Error())
		return
	}
	go func() {
		handleProgress(props, vm, progCh)
		if ctx.Err() == context.DeadlineExceeded {
			logger.Debug("ui.send: context deadline exceeded during send")
		}
	}()
}

func buildExportBundle(picked []string) *sharing.SecretExportBundle {
	payload := sharing.SecretExportPayload{ID: fmt.Sprintf("ui-%d", time.Now().UnixNano()), Timestamp: time.Now().Unix(), Name: fmt.Sprintf("Share %d secret(s)", len(picked)), SenderInfo: sharing.SenderMetadata{DeviceName: "UI"}}
	return &sharing.SecretExportBundle{Payload: payload}
}

func handleProgress(props Props, vm *ViewModel, progCh <-chan service.TransferSendProgress) {
	for p := range progCh {
		switch p.State {
		case service.ShareFlowSucceeded:
			if props.OnClose != nil {
				props.OnClose()
			}
		case service.ShareFlowFailed:
			showErr(props, "Send Failed", fmt.Sprintf("%v", p.Error))
		}
	}
}

func showErr(props Props, title, msg string) {
	if props.Win != nil {
		molecules.ErrorModal(props.Win, molecules.ErrorModalProps{Title: title, Message: msg})
	}
}

// runDiscover performs async discovery with spinner & caching.
func runDiscover(props Props, vm *ViewModel, devicesList *widget.List, spinner *widget.ProgressBarInfinite, transportID string) {
	if props.Service == nil {
		return
	}
	if len(vm.LastDiscovery) > 0 && time.Since(vm.LastDiscoveryAt) < 5*time.Second {
		logger.Debug("ui.discover: using cached device list")
		SetDevices(vm, vm.LastDiscovery, devicesList)
		return
	}
	vm.State = ShareDiscovering
	vm.StatusText = "Discovering..."
	// Clear list during discovery (UI shows dedicated spinner widget)
	SetDevices(vm, []transport.DeviceDescriptor{}, devicesList)

	// Use channel to marshal back results; ensures main thread update via ticker loop.
	resultCh := make(chan []transport.DeviceDescriptor, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		raw, _ := props.Service.DiscoverDevices(ctx, transportID, 50)
		deduped := dedupeDevices(raw, vm.LastDiscovery)
		// Sort for stable UI (by DeviceName then ID)
		sort.Slice(deduped, func(i, j int) bool {
			if deduped[i].DeviceName == deduped[j].DeviceName {
				return deduped[i].DeviceID < deduped[j].DeviceID
			}
			return deduped[i].DeviceName < deduped[j].DeviceName
		})
		resultCh <- deduped
	}()

	// Single-use goroutine to pick up result and update UI.
	go func() {
		res := <-resultCh
		// Filter fallback entries if toggle off
		res = filterFallback(res, vm.ShowFallback)
		vm.LastDiscovery = res
		vm.LastDiscoveryAt = time.Now()
		vm.State = ShareSelecting
		vm.StatusText = fmt.Sprintf("%d device(s) found", len(res))
		logger.Debug(fmt.Sprintf("ui.discover: %d devices after dedupe", len(res)))
		fyne.CurrentApp().SendNotification(&fyne.Notification{Title: "Discovery", Content: fmt.Sprintf("%d device(s) found", len(res))})
		// schedule on main with small delay; protect with mutex if rapid consecutive calls.
		time.AfterFunc(15*time.Millisecond, func() {
			SetDevices(vm, res, devicesList)
			if spinner != nil {
				go func() {
					fyne.DoAndWait(func() {
						spinner.Hide()
					})
				}()
			}
		})
	}()
}

var dedupeMu sync.Mutex

// dedupeDevices merges newly discovered with previous (for fallback merging) and removes duplicates.
func dedupeDevices(current, _ []transport.DeviceDescriptor) []transport.DeviceDescriptor {
	dedupeMu.Lock()
	defer dedupeMu.Unlock()
	index := map[string]transport.DeviceDescriptor{}
	for _, d := range current {
		if ex, ok := index[d.DeviceID]; ok {
			// Replace placeholder (no address) with addressed entry
			if ex.LastAddr == "" && d.LastAddr != "" {
				index[d.DeviceID] = d
			}
			continue
		}
		index[d.DeviceID] = d
	}
	// Drop any entries still without address when there exists at least one addressed entry overall
	hasAddress := false
	for _, v := range index {
		if v.LastAddr != "" {
			hasAddress = true
			break
		}
	}
	out := make([]transport.DeviceDescriptor, 0, len(index))
	for _, v := range index {
		if hasAddress && v.LastAddr == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

// filterFallback removes entries that appear to be fallback-sourced when show==false.
// We currently detect fallback peers as those with LastAddr == "" (no resolved address) OR DeviceName contains "[fallback]".
func filterFallback(devs []transport.DeviceDescriptor, show bool) []transport.DeviceDescriptor {
	if show {
		return devs
	}
	out := make([]transport.DeviceDescriptor, 0, len(devs))
	for _, d := range devs {
		if d.LastAddr == "" || containsFallbackMarker(d.DeviceName) {
			continue
		}
		out = append(out, d)
	}
	return out
}

func containsFallbackMarker(name string) bool { return strings.Contains(name, "[fallback]") }
