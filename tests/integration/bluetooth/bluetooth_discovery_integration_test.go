package integration

import (
	"context"
	"testing"

	"go-password-manager/internal/transport"
	bt "go-password-manager/internal/transport/bluetooth"
	rep "go-password-manager/tests/reporting"

	"github.com/stretchr/testify/require"
)

// TestBluetoothDiscovery validates advertise/scan discovery flow using the test adapter.
func TestBluetoothDiscovery(t *testing.T) {
	rep.WithReporting(t, "TestBluetoothDiscovery", func(r *rep.TestWrapper) {
		platform := bt.NewPlatformAdapter()
		// register a local mailbox and build transport with platform adapter injected
		local := transport.DeviceDescriptor{DeviceID: "local", DeviceName: "local"}
		platform.CreateDevice(local.DeviceID)
		deps := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: platform}
		tr, err := transport.Build(context.Background(), "bluetooth", nil, local, deps)
		require.NoError(t, err)

		// ensure transport exposes Discover via the DiscoverableTransport interface
		dt, ok := tr.(transport.DiscoverableTransport)
		require.True(t, ok, "transport should implement DiscoverableTransport")

		// advertise a device via the platform adapter
		platform.CreateDevice("devX")
		platform.AdvertiseDevice("devX")

		// Discover should surface the advertised device
		list, err := dt.Discover(context.Background(), 10)
		require.NoError(t, err)
		var ids []string
		for _, d := range list {
			ids = append(ids, d.DeviceID)
		}
		require.Contains(t, ids, "devX")

		// stop advertise and ensure Discover no longer returns it
		platform.StopAdvertise("devX")
		list2, err := dt.Discover(context.Background(), 10)
		require.NoError(t, err)
		var ids2 []string
		for _, d := range list2 {
			ids2 = append(ids2, d.DeviceID)
		}
		require.NotContains(t, ids2, "devX")
	})
}
