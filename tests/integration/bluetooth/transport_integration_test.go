package integration

import (
	"context"
	"testing"
	"time"

	"go-password-manager/internal/sharing"
	"go-password-manager/internal/transport"
	bt "go-password-manager/internal/transport/bluetooth"
	rep "go-password-manager/tests/reporting"

	"github.com/stretchr/testify/require"
)

func setupPairedAdapters() (*bt.PlatformAdapter, transport.DeviceDescriptor, transport.DeviceDescriptor) {
	devA := transport.DeviceDescriptor{DeviceID: "devA", DeviceName: "A"}
	devB := transport.DeviceDescriptor{DeviceID: "devB", DeviceName: "B"}
	platform := bt.NewPlatformAdapter()
	platform.CreateDevice(devA.DeviceID)
	platform.CreateDevice(devB.DeviceID)
	return platform, devA, devB
}

// TestBluetoothTransport groups transport-oriented bluetooth scenarios.
func TestBluetoothTransport(t *testing.T) {
	rep.WithReporting(t, "TestBluetoothTransport", func(r *rep.TestWrapper) {
		t.Run("SendReceive", func(t *testing.T) {
			mock, devA, devB := setupPairedAdapters()

			depsA := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: mock}
			depsB := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: mock}

			trA, err := transport.Build(context.Background(), "bluetooth", nil, devA, depsA)
			require.NoError(t, err)
			defer trA.Close()
			trB, err := transport.Build(context.Background(), "bluetooth", nil, devB, depsB)
			require.NoError(t, err)
			defer trB.Close()

			bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: "bundle-1", Timestamp: time.Now().Unix()}}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			recvCh := make(chan *transport.InboundBundle, 1)
			go func() {
				ib, err := trB.Receive(ctx)
				if err == nil && ib != nil {
					recvCh <- ib
				}
			}()

			receipt, err := trA.Send(ctx, bundle, devB)
			require.NoError(t, err)
			require.Equal(t, bundle.Payload.ID, receipt.BundleID)

			select {
			case ib := <-recvCh:
				require.Equal(t, bundle.Payload.ID, ib.Bundle.Payload.ID)
			case <-ctx.Done():
				require.FailNow(t, "timeout waiting for receive")
			}
		})

		t.Run("Chunking", func(t *testing.T) {
			mock, devA, devB := setupPairedAdapters()
			mock.SetMTU(16) // force fragmentation

			depsA := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: mock}
			depsB := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: mock}

			trA, err := transport.Build(context.Background(), "bluetooth", nil, devA, depsA)
			require.NoError(t, err)
			defer trA.Close()
			trB, err := transport.Build(context.Background(), "bluetooth", nil, devB, depsB)
			require.NoError(t, err)
			defer trB.Close()

			longID := "bundle-chunk-" + string(make([]byte, 200))
			bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: longID, Timestamp: time.Now().Unix()}}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			recvCh := make(chan *transport.InboundBundle, 1)
			go func() {
				ib, err := trB.Receive(ctx)
				if err == nil && ib != nil {
					recvCh <- ib
				}
			}()

			receipt, err := trA.Send(ctx, bundle, devB)
			require.NoError(t, err)
			require.Equal(t, bundle.Payload.ID, receipt.BundleID)

			select {
			case ib := <-recvCh:
				require.Equal(t, bundle.Payload.ID, ib.Bundle.Payload.ID)
			case <-ctx.Done():
				require.FailNow(t, "timeout waiting for receive")
			}
		})

		t.Run("Retry", func(t *testing.T) {
			mock, devA, devB := setupPairedAdapters()
			mock.FailNextWrites(2) // first writes fail

			depsA := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: mock}
			depsB := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: mock}

			trA, err := transport.Build(context.Background(), "bluetooth", nil, devA, depsA)
			require.NoError(t, err)
			defer trA.Close()
			trB, err := transport.Build(context.Background(), "bluetooth", nil, devB, depsB)
			require.NoError(t, err)
			defer trB.Close()

			bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: "bundle-retry", Timestamp: time.Now().Unix()}}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			recvCh := make(chan *transport.InboundBundle, 1)
			go func() {
				ib, err := trB.Receive(ctx)
				if err == nil && ib != nil {
					recvCh <- ib
				}
			}()

			receipt, err := trA.Send(ctx, bundle, devB)
			require.NoError(t, err)
			require.GreaterOrEqual(t, receipt.Attempt, 1)

			select {
			case ib := <-recvCh:
				require.Equal(t, bundle.Payload.ID, ib.Bundle.Payload.ID)
			case <-ctx.Done():
				require.FailNow(t, "timeout waiting for receive")
			}
		})
	})
}
