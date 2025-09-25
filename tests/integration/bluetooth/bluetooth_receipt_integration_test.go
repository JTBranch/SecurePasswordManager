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

// TestBluetoothReceiptMetadata asserts receipt timing and envelope metadata.
func TestBluetoothReceiptMetadata(t *testing.T) {
	rep.WithReporting(t, "TestBluetoothReceiptMetadata", func(r *rep.TestWrapper) {
		devA := transport.DeviceDescriptor{DeviceID: "devA", DeviceName: "A"}
		devB := transport.DeviceDescriptor{DeviceID: "devB", DeviceName: "B"}

		platform := bt.NewPlatformAdapter()
		platform.CreateDevice("devA")
		platform.CreateDevice("devB")
		depsA := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: platform}
		depsB := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: platform}

		trA, err := transport.Build(context.Background(), "bluetooth", nil, devA, depsA)
		require.NoError(t, err)
		defer trA.Close()
		trB, err := transport.Build(context.Background(), "bluetooth", nil, devB, depsB)
		require.NoError(t, err)
		defer trB.Close()

		bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: "bundle-rcpt", Timestamp: time.Now().Unix()}}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		recvCh := make(chan *transport.InboundBundle, 1)
		go func() {
			ib, err := trB.Receive(ctx)
			if err == nil && ib != nil {
				recvCh <- ib
			}
		}()

		started := time.Now()
		receipt, err := trA.Send(ctx, bundle, devB)
		require.NoError(t, err)
		require.Equal(t, bundle.Payload.ID, receipt.BundleID)
		require.GreaterOrEqual(t, receipt.Attempt, 1)
		require.True(t, !receipt.StartedAt.IsZero(), "StartedAt should be set")
		require.True(t, !receipt.CompletedAt.IsZero(), "CompletedAt should be set")
		require.True(t, receipt.StartedAt.After(started.Add(-time.Second)), "StartedAt reasonable")

		select {
		case ib := <-recvCh:
			require.Equal(t, bundle.Payload.ID, ib.Bundle.Payload.ID)
			// ensure envelope metadata exists (our test adapter sets bluetooth:true)
			v, ok := ib.Envelope["bluetooth"]
			require.True(t, ok, "expected bluetooth envelope metadata")
			require.Equal(t, true, v)
		case <-ctx.Done():
			require.FailNow(t, "timeout waiting for receive")
		}
	})
}
