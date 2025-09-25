//go:build !windows
// +build !windows

package adapters

import (
	"context"
	"errors"
	"time"

	"github.com/go-ble/ble"
)

// connectWithDial handles dialing a BLE address, discovering the profile and
// returning a Conn via newConnFromClient.
func connectWithDial(ctx context.Context, deviceID string) (interface{}, error) {
	ctxDial, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cln, err := ble.Dial(ctxDial, ble.NewAddr(deviceID))
	if err != nil {
		return nil, err
	}
	p, err := cln.DiscoverProfile(true)
	if err != nil {
		_ = cln.CancelConnection()
		return nil, err
	}
	foundChar := findCharacteristicFromProfile(p)
	if foundChar == nil {
		_ = cln.CancelConnection()
		return nil, errors.New("characteristic not found on remote device")
	}
	return newConnFromClient(cln, foundChar)
}

// findCharacteristicFromProfile scans a profile for the configured
// service/characteristic and returns the characteristic or nil.
func findCharacteristicFromProfile(p *ble.Profile) *ble.Characteristic {
	if p == nil {
		return nil
	}
	for _, srv := range p.Services {
		if srv.UUID.Equal(serviceUUID) {
			for _, c := range srv.Characteristics {
				if c.UUID.Equal(charUUID) {
					return c
				}
			}
		}
	}
	return nil
}
