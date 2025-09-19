package sharing_test

import (
	"go-password-manager/internal/sharing"
	tmocks "go-password-manager/internal/testHelpers/mocks"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewImportServiceWithReplayConfigFileMode(t *testing.T) {
	cryptoProvider := tmocks.NewCryptoProvider(t)
	deviceKeyProvider := tmocks.NewDeviceKeyProvider(t)
	secretsProvider := tmocks.NewSecretsProvider(t)
	dir := t.TempDir()
	cfg := sharing.ReplayConfig{Mode: sharing.ReplayModeFile, FilePath: filepath.Join(dir, "store.json"), TTL: time.Hour, MaxEntries: 10}
	svc, err := sharing.NewImportServiceWithReplayConfig(cryptoProvider, deviceKeyProvider, secretsProvider, cfg)
	require.NoError(t, err, "unexpected err: %v", err)
	require.NotNil(t, svc.GetReplayStore(), "expected replay store")
}
