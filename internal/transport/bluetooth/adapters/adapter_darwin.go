//go:build darwin
// +build darwin

package adapters

import (
	"context"

	"github.com/go-ble/ble"
	bledarwin "github.com/go-ble/ble/darwin"
)

// GetSystemAdapter returns a Darwin adapter implementation backed by CoreBluetooth.
func init() {
	RegisterAdapter("darwin", darwinFactory)
	RegisterConnector("darwin", connectToDeviceDarwin)
	RegisterScanner("darwin", func(serviceUUID string, limit int) []string {
		return scanWithBLE(serviceUUID, limit)
	})
}

func darwinFactory(name string) (interface{}, error) {
	d, err := bledarwin.NewDevice()
	if err != nil {
		return nil, err
	}
	ble.SetDefaultDevice(d)
	go func() { _ = ble.AdvertiseNameAndServices(context.Background(), "VPM", serviceUUID) }()
	return &struct{}{}, nil
}

// connectToDeviceDarwin dials a remote device and returns a minimal Conn.
func connectToDeviceDarwin(ctx context.Context, deviceID string) (interface{}, error) {
	return connectWithDial(ctx, deviceID)
}
