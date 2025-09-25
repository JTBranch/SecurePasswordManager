package service

import (
	"context"
	"errors"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/testHelpers/mocks"
	"go-password-manager/internal/transport"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockTransport struct {
	mock.Mock
	inbound  *transport.InboundBundle
	received bool
}

func (m *MockTransport) ID() string { return "lan" }
func (m *MockTransport) Send(ctx context.Context, bundle *sharing.SecretExportBundle, target transport.DeviceDescriptor) (*transport.TransportReceipt, error) {
	args := m.Called(ctx, bundle, target)
	var r *transport.TransportReceipt
	if v := args.Get(0); v != nil {
		r = v.(*transport.TransportReceipt)
	}
	return r, args.Error(1)
}
func (m *MockTransport) Receive(ctx context.Context) (*transport.InboundBundle, error) {
	if m.received {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	m.received = true
	if m.inbound == nil {
		return nil, nil
	}
	return m.inbound, nil
}
func (m *MockTransport) Close() error { return nil }

func TestTransferServicePrepareExportSuccess(t *testing.T) {
	exp := &mocks.ExportProvider{}
	exp.On("Export", mock.Anything, mock.Anything, 60, mock.Anything).Return(&sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: "export-123", Timestamp: time.Now().Unix()}}, nil)
	svc := NewSharingTransferService(transport.DeviceDescriptor{DeviceID: "L"}, transport.Dependencies{Registry: transport.NewInMemoryRegistry()}, exp, nil, nil, nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, progCh, err := svc.PrepareExport(ctx, nil, 60, sharing.SenderMetadata{DeviceName: "dev"})
	require.NoError(t, err, "PrepareExport error: %v", err)
	states := []ExportFlowState{}
	var final ExportProgress
	for p := range progCh {
		states = append(states, p.State)
		final = p
	}
	t.Logf("Export progress: %#v", states)
	require.Equal(t, 2, len(states), "unexpected states: %#v", states)
	require.Equal(t, ExportStatePreparing, states[0], "unexpected states: %#v", states)
	require.Equal(t, ExportStateReady, states[1], "unexpected states: %#v", states)
	require.Equal(t, "export-123", final.BundleID)
}

func TestTransferServiceSendBundleSuccess(t *testing.T) {
	bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: "b1", Timestamp: time.Now().Unix()}}
	tr := &MockTransport{}
	tr.On("Send", mock.Anything, bundle, mock.Anything).Return(&transport.TransportReceipt{TransportID: "lan", BundleID: "b1", StartedAt: time.Now(), CompletedAt: time.Now(), Attempt: 1}, nil)
	builder := &mocks.TransportBuilder{}
	builder.On("Build", mock.Anything, "lan", mock.Anything, mock.Anything, mock.Anything).Return(tr, nil)
	svc := NewSharingTransferService(transport.DeviceDescriptor{DeviceID: "S"}, transport.Dependencies{Registry: transport.NewInMemoryRegistry()}, nil, nil, nil, builder)
	target := transport.DeviceDescriptor{DeviceID: "T", LastAddr: "127.0.0.1:9999"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ch, err := svc.SendBundle(ctx, "lan", bundle, target)
	require.NoError(t, err, "SendBundle error: %v", err)
	seenQueued, seenSucceeded := false, false
	for p := range ch {
		switch p.State {
		case ShareFlowQueued:
			seenQueued = true
		case ShareFlowSucceeded:
			seenSucceeded = true
		case ShareFlowFailed:
			require.FailNowf(t, "unexpected failure", "%v", p.Error)
		}
	}
	assert.True(t, seenQueued, "expected queued state not observed")
	assert.True(t, seenSucceeded, "expected succeeded state not observed")
}

func TestTransferServiceReceiveOnceAutoImport(t *testing.T) {
	bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: "rx1", Timestamp: time.Now().Unix()}}
	inbound := &transport.InboundBundle{From: transport.DeviceDescriptor{DeviceID: "Sender"}, Bundle: bundle, ReceivedAt: time.Now()}
	tr := &MockTransport{inbound: inbound}
	builder := &mocks.TransportBuilder{}
	builder.On("Build", mock.Anything, "lan", mock.Anything, mock.Anything, mock.Anything).Return(tr, nil)
	imp := &mocks.ImportProvider{}
	imp.On("ImportSecrets", bundle, mock.Anything, mock.Anything).Return(&sharing.SecretImportResult{ImportedSecretsCount: 1, Success: true}, nil)
	svc := NewSharingTransferService(transport.DeviceDescriptor{DeviceID: "R"}, transport.Dependencies{Registry: transport.NewInMemoryRegistry()}, nil, imp, nil, builder)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	evCh, err := svc.ReceiveOnce(ctx, "lan", true)
	require.NoError(t, err, "ReceiveOnce error: %v", err)
	gotBundle, gotImport := false, false
	for ev := range evCh {
		switch ev.Type {
		case InboundBundleReceived:
			gotBundle = true
		case InboundImportSucceeded:
			gotImport = true
		case InboundImportFailed:
			require.FailNowf(t, "import failed", "%v", ev.Error)
		}
	}
	assert.True(t, gotBundle, "expected bundle event not observed")
	assert.True(t, gotImport, "expected import event not observed")
}

func TestTransferServiceSendBundleError(t *testing.T) {
	bundle := &sharing.SecretExportBundle{Payload: sharing.SecretExportPayload{ID: "b2"}}
	tr := &MockTransport{}
	tr.On("Send", mock.Anything, bundle, mock.Anything).Return(&transport.TransportReceipt{TransportID: "lan", BundleID: "b2", StartedAt: time.Now(), Attempt: 1, CompletedAt: time.Now(), Error: errors.New("boom")}, errors.New("boom"))
	builder := &mocks.TransportBuilder{}
	builder.On("Build", mock.Anything, "lan", mock.Anything, mock.Anything, mock.Anything).Return(tr, nil)
	svc := NewSharingTransferService(transport.DeviceDescriptor{DeviceID: "S"}, transport.Dependencies{Registry: transport.NewInMemoryRegistry()}, nil, nil, nil, builder)
	target := transport.DeviceDescriptor{DeviceID: "T", LastAddr: "127.0.0.1:9999"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ch, _ := svc.SendBundle(ctx, "lan", bundle, target)
	failed := false
	for p := range ch {
		if p.State == ShareFlowFailed {
			failed = true
		}
	}
	assert.True(t, failed, "expected failure state not observed")
}
