package sharing_test

import (
	"go-password-manager/internal/sharing"
	tmocks "go-password-manager/internal/testHelpers/mocks"
	"path/filepath"
	"testing"
	"time"
)

func TestNewImportServiceWithReplayConfigFileMode(t *testing.T) {
	cryptoProvider := tmocks.NewCryptoProvider(t)
	deviceKeyProvider := tmocks.NewDeviceKeyProvider(t)
	secretsProvider := tmocks.NewSecretsProvider(t)
	dir := t.TempDir()
	cfg := sharing.ReplayConfig{Mode: sharing.ReplayModeFile, FilePath: filepath.Join(dir, "store.json"), TTL: time.Hour, MaxEntries: 10}
	svc, err := sharing.NewImportServiceWithReplayConfig(cryptoProvider, deviceKeyProvider, secretsProvider, cfg)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if svc.GetReplayStore() == nil {
		t.Fatalf("expected replay store")
	}
}
