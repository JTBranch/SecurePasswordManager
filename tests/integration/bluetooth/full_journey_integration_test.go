package integration

import (
	"context"
	"testing"
	"time"

	"go-password-manager/tests/helpers"
	"go-password-manager/tests/testdata"

	"go-password-manager/internal/transport"
	bt "go-password-manager/internal/transport/bluetooth"
	rep "go-password-manager/tests/reporting"

	"github.com/stretchr/testify/require"
)

// Full journey: export on sender -> send via bluetooth -> receive on receiver -> import -> assert secrets present
func TestBluetoothFullJourney(t *testing.T) {
	rep.WithReporting(t, "TestBluetoothFullJourney", func(r *rep.TestWrapper) {
		// Setup sender and receiver test suites (isolated data dirs)
		sender := helpers.NewIntegrationTestSuite(r)
		sender.SetupTestEnvironment()
		defer sender.Cleanup()

		receiver := helpers.NewIntegrationTestSuite(r)
		receiver.SetupTestEnvironment()
		defer receiver.Cleanup()

		td := testdata.NewTestDataManager()
		// Create two secrets on sender
		require.NoError(r.T(), td.CreateTestSecret(sender.SecretsService, testdata.TestSecrets.Simple.Name))
		require.NoError(r.T(), td.CreateTestSecret(sender.SecretsService, testdata.TestSecrets.Complex.Name))

		file, err := sender.SecretsService.LoadAllSecrets()
		require.NoError(r.T(), err)

		// Generate recipient key pair on receiver (PEM encoded)
		pubPEM, privPEM, err := receiver.CryptoService.GenerateX25519KeyPairPEM()
		require.NoError(r.T(), err)

		// Sender exports bundle targeted at receiver's public key
		exportBundle, err := sender.SharingService.ExportSecrets(file.Secrets, pubPEM, 300)
		require.NoError(r.T(), err)
		require.NotNil(r.T(), exportBundle)

		// Setup in-memory platform adapter and create two devices
		devA := transport.DeviceDescriptor{DeviceID: "sender", DeviceName: "sender"}
		devB := transport.DeviceDescriptor{DeviceID: "receiver", DeviceName: "receiver"}
		platform := bt.NewPlatformAdapter()
		// Register both devices' mailboxes and advertise receiver
		platform.CreateDevice(devA.DeviceID)
		platform.CreateDevice(devB.DeviceID)
		platform.AdvertiseDevice(devB.DeviceID)

		depsA := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: platform}
		depsB := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), BluetoothAdapter: platform}

		trA, err := transport.Build(context.Background(), "bluetooth", nil, devA, depsA)
		require.NoError(r.T(), err)
		defer trA.Close()
		trB, err := transport.Build(context.Background(), "bluetooth", nil, devB, depsB)
		require.NoError(r.T(), err)
		defer trB.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Receiver goroutine: receive bundle, import into receiver.SecretsService, and signal result
		done := make(chan error, 1)
		go func() {
			ib, err := trB.Receive(ctx)
			if err != nil {
				done <- err
				return
			}
			if ib == nil {
				done <- nil
				return
			}
			// Import using receiver's import service (real import path)
			res, impErr := receiver.SharingService.ImportSecrets(ib.Bundle, privPEM, pubPEM)
			if impErr != nil {
				done <- impErr
				return
			}
			if res == nil || !res.Success {
				if res != nil && res.Error != nil {
					done <- res.Error
				} else {
					done <- nil
				}
				return
			}
			// Verify secrets now exist on receiver
			s1, err := receiver.SecretsService.GetSecret(testdata.TestSecrets.Simple.Name)
			if err != nil {
				done <- err
				return
			}
			if s1 == nil {
				done <- nil
				return
			}
			_, err = receiver.SecretsService.GetSecretValue(s1)
			if err != nil {
				done <- err
				return
			}
			s2, err := receiver.SecretsService.GetSecret(testdata.TestSecrets.Complex.Name)
			if err != nil {
				done <- err
				return
			}
			if s2 == nil {
				done <- nil
				return
			}
			_, err = receiver.SecretsService.GetSecretValue(s2)
			if err != nil {
				done <- err
				return
			}
			done <- nil
		}()

		// Send bundle from sender to receiver
		receipt, err := trA.Send(ctx, exportBundle, devB)
		require.NoError(r.T(), err)
		require.Equal(r.T(), exportBundle.Payload.ID, receipt.BundleID)

		select {
		case err := <-done:
			require.NoError(r.T(), err, "import/verification failed on receiver")
		case <-ctx.Done():
			require.FailNow(r.T(), "timeout waiting for full journey to complete")
		}
	})
}
