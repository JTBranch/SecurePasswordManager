package integration

import (
	buildconfig "go-password-manager/internal/config/buildconfig"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/config/secretkeymetadata"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/service"
	"go-password-manager/internal/storage"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNoKeychainWriteInTestEnv ensures that creating a secret in test env does not touch OS keychain
func TestNoKeychainWriteInTestEnv(t *testing.T) {
	os.Setenv("GO_PASSWORD_MANAGER_ENV", "test")
	t.Cleanup(func() { os.Unsetenv("GO_PASSWORD_MANAGER_ENV") })
	// Provide temp test dir (do not export global env var to avoid clobbering other tests)
	testDir := t.TempDir()

	buildconfig.ResetCacheForTest()
	cfg, err := buildconfig.Load()
	require.NoError(t, err, "config load")
	require.Equal(t, "test", cfg.Application.Environment, "expected test env")

	// Ensure this test uses its own data dir without mutating global environment
	cfg.Testing.DataDir = testDir

	cfgSvc, err := config.NewConfigService(cfg)
	require.NoError(t, err, "config svc")
	metaSvc, err := secretkeymetadata.NewSecretKeyMetadataFileService(cfg)
	require.NoError(t, err, "meta svc")

	sekMgr, err := crypto.NewSecretsEncryptionKeyManager(cfgSvc, metaSvc)
	require.NoError(t, err, "sek mgr")
	require.True(t, sekMgr.UsingInMemoryKeyring(), "expected in-memory keyring from constructor in test env")

	cryptoSvc, err := crypto.NewCryptoServiceDefault(cfgSvc, sekMgr)
	require.NoError(t, err, "crypto svc")

	secretsPath, err := cfg.GetSecretsFilePath()
	require.NoError(t, err, "secrets path")
	storageSvc := storage.NewFileStorage(secretsPath, cfg.Application.Version, "it-user")
	secretsSvc := service.NewSecretsService(cryptoSvc, storageSvc)

	// Create a secret (forces key creation & encryption usage)
	err = secretsSvc.SaveNewSecret("integration-keychain-check", "value")
	require.NoError(t, err, "create secret")

	// Attempt to read from real keyring using default provider; should fail with not exist
	sysProvider := &crypto.DefaultkeyringProvider{}
	// Use the same key UUID to attempt retrieval
	keyUUID := cfgSvc.GetKeyUUID()
	_, getErr := sysProvider.Get(crypto.KeyringSecretsEncryption, keyUUID)
	require.Error(t, getErr, "expected no system keyring entry for %s / %s", crypto.KeyringSecretsEncryption, keyUUID)
}
