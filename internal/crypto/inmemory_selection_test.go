package crypto

import (
	buildconfig "go-password-manager/internal/config/buildconfig"
	"go-password-manager/internal/config/devicekeys"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/config/secretkeymetadata"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// helper to set env and reset build config cache
func withEnv(t *testing.T, key, value string, fn func()) {
	orig := os.Getenv(key)
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
	buildconfig.ResetCacheForTest()
	t.Cleanup(func() {
		if orig == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, orig)
		}
		buildconfig.ResetCacheForTest()
	})
	fn()
}

// ensure test config path exists for test environment
func ensureTestDataDir(t *testing.T) string {
	dir := t.TempDir()
	os.Setenv("TEST_DATA_DIR", dir)
	t.Cleanup(func() { os.Unsetenv("TEST_DATA_DIR") })
	return dir
}

func TestSecretsEncryptionKeyManagerUsesInMemoryWhenConfigured(t *testing.T) {
	ensureTestDataDir(t)
	withEnv(t, "GO_PASSWORD_MANAGER_ENV", "test", func() {
		// Force in_memory true via test.yaml (already true) or env override if needed
		cfg, err := buildconfig.Load()
		require.NoError(t, err, "failed load: %v", err)
		require.True(t, cfg.Security.Keyring.InMemory, "expected in_memory true in test config")
		cfgSvc, err := config.NewConfigService(cfg)
		require.NoError(t, err, "cfg svc err: %v", err)
		metaSvc, err := secretkeymetadata.NewSecretKeyMetadataFileService(cfg)
		require.NoError(t, err, "meta svc err: %v", err)
		mgr, err := NewSecretsEncryptionKeyManager(cfgSvc, metaSvc)
		require.NoError(t, err, "manager err: %v", err)
		require.True(t, mgr.UsingInMemoryKeyring(), "expected in-memory keyring provider")
	})
}

func TestDeviceKeyManagerUsesInMemoryWhenConfigured(t *testing.T) {
	ensureTestDataDir(t)
	withEnv(t, "GO_PASSWORD_MANAGER_ENV", "test", func() {
		cfg, err := buildconfig.Load()
		require.NoError(t, err, "failed load: %v", err)
		require.True(t, cfg.Security.Keyring.InMemory, "expected in_memory true in test config")
		cfgSvc, err := config.NewConfigService(cfg)
		require.NoError(t, err, "cfg svc err: %v", err)
		metaSvc, err := secretkeymetadata.NewSecretKeyMetadataFileService(cfg)
		require.NoError(t, err, "meta svc err: %v", err)
		sekMgr, err := NewSecretsEncryptionKeyManager(cfgSvc, metaSvc)
		require.NoError(t, err, "sek mgr err: %v", err)
		cryptoSvc, err := NewCryptoServiceDefault(cfgSvc, sekMgr)
		require.NoError(t, err, "crypto svc err: %v", err)
		deviceKeyFileSvc := devicekeys.NewDeviceKeyFileService(cfg)
		devMgr, err := NewDeviceKeyManager(cryptoSvc, &PemUtils{}, deviceKeyFileSvc)
		require.NoError(t, err, "dev mgr err: %v", err)
		require.True(t, devMgr.UsingInMemoryProviders(), "expected in-memory providers for device manager")
	})
}
