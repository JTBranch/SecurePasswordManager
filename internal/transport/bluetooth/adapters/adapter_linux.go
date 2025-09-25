//go:build linux
// +build linux

package adapters

import (
	"context"

	"github.com/go-ble/ble"
	belinux "github.com/go-ble/ble/linux"
)

func init() {
	RegisterAdapter("linux", linuxFactory)
	RegisterConnector("linux", connectToDeviceLinux)
	RegisterScanner("linux", func(serviceUUID string, limit int) []string {
		// We reuse any platform scanner available on the default device. The
		// go-ble library exposes Scan via package-level helpers; adapters may
		// instead provide their own Scan implementation if necessary. For
		// now, use ble.Scan which returns ADV addresses.
		return scanWithBLE(serviceUUID, limit)
	})
}

func linuxFactory(name string) (interface{}, error) {
	d, err := belinux.NewDevice()
	if err != nil {
		return nil, err
	}
	ble.SetDefaultDevice(d)
	a := &struct{}{}
	go func() {
		_ = ble.AdvertiseNameAndServices(context.Background(), "VPM", serviceUUID)
	}()
	return a, nil
}

// ConnectToDevice dials a remote device and returns a minimal Conn. The
// returned value is an implementation-specific type and must be type-asserted
// by the caller to the expected `bluetooth.Conn` interface.
func connectToDeviceLinux(ctx context.Context, deviceID string) (interface{}, error) {
	return connectWithDial(ctx, deviceID)
}
