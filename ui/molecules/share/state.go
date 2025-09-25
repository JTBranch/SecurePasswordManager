package share

import (
	"go-password-manager/internal/service"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/transport"
	"time"

	"fyne.io/fyne/v2"
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

type Props struct {
	Win        fyne.Window
	Service    *service.SharingTransferService
	Secrets    []sharing.ExportSecret
	OnClose    func()
	OnExported func(bundle *sharing.SecretExportBundle) // future real export hook
}

type ViewModel struct {
	State           ShareModalState
	Devices         []transport.DeviceDescriptor
	SelectedDevice  map[int]bool
	SelectedSecret  map[int]bool
	SecretNames     []string
	StatusText      string
	LastDiscovery   []transport.DeviceDescriptor
	LastDiscoveryAt time.Time
	ShowFallback    bool
	OnDeviceAction  func(idx int, action string)
}

func NewViewModel(secretNames []string) *ViewModel {
	return &ViewModel{State: ShareIdle, Devices: []transport.DeviceDescriptor{}, SelectedSecret: map[int]bool{}, SelectedDevice: map[int]bool{}, SecretNames: secretNames, StatusText: "", ShowFallback: false, OnDeviceAction: nil}
}
