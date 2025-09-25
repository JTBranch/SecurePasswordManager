package integration

import (
	"go-password-manager/internal/transport"
	bt "go-password-manager/internal/transport/bluetooth"
)

// SetupPairedPlatform creates a PlatformAdapter with two devices registered and
// returns the adapter and the corresponding DeviceDescriptors.
func SetupPairedPlatform(aID, bID string) (*bt.PlatformAdapter, transport.DeviceDescriptor, transport.DeviceDescriptor) {
	platform := bt.NewPlatformAdapter()
	platform.CreateDevice(aID)
	platform.CreateDevice(bID)
	return platform, transport.DeviceDescriptor{DeviceID: aID, DeviceName: aID}, transport.DeviceDescriptor{DeviceID: bID, DeviceName: bID}
}

// SetupAdvertisedDevice registers and advertises the given device id on the adapter.
func SetupAdvertisedDevice(platform *bt.PlatformAdapter, id string) transport.DeviceDescriptor {
	platform.CreateDevice(id)
	platform.AdvertiseDevice(id)
	return transport.DeviceDescriptor{DeviceID: id, DeviceName: id}
}
