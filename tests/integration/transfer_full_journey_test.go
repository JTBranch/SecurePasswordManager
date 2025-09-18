package integration

import (
	"context"
	"go-password-manager/internal/domain"
	"go-password-manager/internal/service"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/transport"
	"go-password-manager/tests/helpers"
	"go-password-manager/tests/reporting"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// helper to map first secret to ExportSecret slice
func buildExportSecrets(secs []domain.Secret, svc *service.SecretsService) ([]sharing.ExportSecret, error) {
	out := make([]sharing.ExportSecret, 0, len(secs))
	for _, s := range secs {
		v, err := svc.GetSecretValue(&s)
		if err != nil {
			return nil, err
		}
		cur := s.GetCurrentVersion()
		out = append(out, sharing.ExportSecret{Name: s.SecretName, Type: s.Type, Value: v, UpdatedAt: cur.UpdatedAt, Version: cur.Version})
	}
	return out, nil
}

// TestFullTransferJourney exercises export via export provider, LAN send, receive + import using SharingTransferService with real dependencies.
func TestFullTransferJourney(t *testing.T) {
	reporting.WithReporting(t, "TestFullTransferJourney", func(reporter *reporting.TestWrapper) {
		senderSuite := helpers.NewIntegrationTestSuite(reporter)
		receiverSuite := helpers.NewIntegrationTestSuite(reporter)
		senderSuite.SetupTestEnvironment()
		receiverSuite.SetupTestEnvironment()
		// Create a secret on sender
		// Create a secret
		err := senderSuite.SecretsService.SaveNewSecret("journey-secret", "super-value")
		require.NoError(t, err)
		file, err := senderSuite.SecretsService.LoadAllSecrets()
		require.NoError(t, err)
		// Generate recipient X25519 keys (receiver)
		pubPEM, privPEM, err := receiverSuite.CryptoService.GenerateX25519KeyPairPEM()
		require.NoError(t, err)
		// Build export provider & import provider
		exportProv := senderSuite.ExportService
		importProv := receiverSuite.ImportService
		// Build transfer services (stateless)
		senderLocal := transport.DeviceDescriptor{DeviceID: "sender-device", DeviceName: "Sender", LastAddr: ""}
		receiverLocal := transport.DeviceDescriptor{DeviceID: "receiver-device", DeviceName: "Receiver", LastAddr: ""}
		deps := transport.Dependencies{Registry: transport.NewInMemoryRegistry()}
		receiverSvc := service.NewSharingTransferService(receiverLocal, deps, nil, importProv, nil, nil)
		senderSvc := service.NewSharingTransferService(senderLocal, deps, nil, nil, nil, nil)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		seclist, err := buildExportSecrets(file.Secrets, senderSuite.SecretsService)
		require.NoError(t, err)
		exportBundle, err := exportProv.ExportSecrets(seclist, pubPEM, 3600, sharing.SenderMetadata{DeviceName: "Sender"})
		require.NoError(t, err)
		// Start receiver goroutine waiting for inbound
		recvCtx, recvCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer recvCancel()
		listenAddr := "127.0.0.1:39001"
		evCh, err := receiverSvc.ReceiveOnceWithKeysAt(recvCtx, "lan", true, listenAddr, privPEM, pubPEM)
		require.NoError(t, err)
		// Allow listener to bind before sending
		time.Sleep(150 * time.Millisecond)
		// Build target descriptor (receiver listens on dynamic address; transport handles dialing via LastAddr embed from advertisement - simplified here by manual loopback address)
		target := transport.DeviceDescriptor{DeviceID: "receiver-device", DeviceName: "Receiver", LastAddr: listenAddr}
		sendCh, err := senderSvc.SendBundle(ctx, "lan", exportBundle, target)
		require.NoError(t, err)
		// Drain send progress ensure no failure
		for sp := range sendCh {
			if sp.State == service.ShareFlowFailed {
				t.Fatalf("send failed: %v", sp.Error)
			}
		}
		// Collect receiver events
		var gotBundle, gotImport bool
		for ev := range evCh {
			if ev.Type == service.InboundBundleReceived {
				gotBundle = true
			}
			if ev.Type == service.InboundImportSucceeded {
				gotImport = true
			}
		}
		require.True(t, gotBundle, "receiver did not get bundle event")
		require.True(t, gotImport, "receiver did not import bundle")
	})
}
