//go:build !windows
// +build !windows

package adapters

import (
	"testing"

	"github.com/go-ble/ble"
)

func TestFindCharacteristicFromProfile(t *testing.T) {
	// build a profile with the expected service and characteristic
	char := &ble.Characteristic{UUID: charUUID}
	srv := &ble.Service{UUID: serviceUUID, Characteristics: []*ble.Characteristic{char}}
	p := &ble.Profile{Services: []*ble.Service{srv}}

	found := findCharacteristicFromProfile(p)
	if found == nil {
		t.Fatalf("expected characteristic, got nil")
	}
	if !found.UUID.Equal(char.UUID) {
		t.Fatalf("characteristic UUID mismatch")
	}
}
