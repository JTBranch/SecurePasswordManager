package di

import (
	"context"
	"go-password-manager/internal/config/buildconfig"
	"go-password-manager/internal/config/devicekeys"
	config "go-password-manager/internal/config/runtimeconfig"
	"go-password-manager/internal/config/secretkeymetadata"
	"go-password-manager/internal/crypto"
	"go-password-manager/internal/logger"
	"go-password-manager/internal/service"
	"go-password-manager/internal/sharing"
	"go-password-manager/internal/storage"
	"go-password-manager/internal/transport"
	"strconv"
	"time"
)

// SharingBundle aggregates constructed sharing-related services.
type SharingBundle struct {
	SecretsService   *service.SecretsService
	DeviceKeyManager *crypto.DeviceKeyManager
	ExportService    *sharing.ExportService
	ImportService    *sharing.ImportService
	TransferService  *service.SharingTransferService
	DeviceDescriptor transport.DeviceDescriptor
	AdvertiseCloser  func() // optional cleanup for advertising transport
}

// internal adapter replicating main.go pattern
type exportAdapter struct{ *sharing.ExportService }

func (a exportAdapter) Export(secrets []sharing.ExportSecret, recipientPubKey []byte, expiryMinutes int, meta sharing.SenderMetadata) (*sharing.SecretExportBundle, error) {
	return a.ExportService.ExportSecrets(secrets, recipientPubKey, expiryMinutes, meta)
}

// BuildSharing constructs sharing related services and transfer orchestrator.
func BuildSharing(buildCfg *buildconfig.Config) (*SharingBundle, error) {
	cfgSvc, err := config.NewConfigService(buildCfg)
	if err != nil {
		return nil, err
	}
	secretMetaSvc, err := secretkeymetadata.NewSecretKeyMetadataFileService(buildCfg)
	if err != nil {
		return nil, err
	}
	sekMgr, err := crypto.NewSecretsEncryptionKeyManager(cfgSvc, secretMetaSvc)
	if err != nil {
		return nil, err
	}
	cryptoSvc, err := crypto.NewCryptoServiceDefault(cfgSvc, sekMgr)
	if err != nil {
		return nil, err
	}
	secretsPath, err := buildCfg.GetSecretsFilePath()
	if err != nil {
		return nil, err
	}
	storageSvc := storage.NewFileStorage(secretsPath, buildCfg.Application.Version, "e2e-user")
	secretsSvc := service.NewSecretsService(cryptoSvc, storageSvc)
	deviceKeyFileSvc := devicekeys.NewDeviceKeyFileService(buildCfg)
	deviceKeyMgr, err := crypto.NewDeviceKeyManager(cryptoSvc, &crypto.PemUtils{}, deviceKeyFileSvc)
	if err != nil {
		return nil, err
	}
	// Determine if we should operate entirely in-memory (no keychain, no device_keys.json persistence).
	inMem := buildCfg.Security.Keyring.InMemory || buildCfg.IsTest()
	if inMem {
		logger.Info("keyring mode: in-memory (no host keychain writes, no device_keys.json persistence)")
		deviceKeyMgr.SetKeyringProvider(crypto.NewInMemoryKeyring())
		// Replace device key file provider with in-memory variant so rotations & saves don't touch disk.
		deviceKeyMgr.SetDeviceKeyFileProvider(crypto.NewInMemoryDeviceKeyFileProvider())
	} else {
		logger.Info("keyring mode: system keychain + disk persistence")
	}
	// Build descriptor lazily (ignore errors retrieving keys until needed)
	encKey, encErr := deviceKeyMgr.GetEncryptionDeviceKey()
	sigKey, sigErr := deviceKeyMgr.GetSigningDeviceKey()
	var desc transport.DeviceDescriptor
	if encErr == nil && sigErr == nil {
		desc = transport.DeviceDescriptor{DeviceID: encKey.ID, UserID: deviceKeyMgr.GetAppUser(), DeviceName: deviceKeyMgr.GetDeviceName(), Ed25519Pub: sigKey.PublicKey, X25519Pub: encKey.PublicKey}
	}
	exp := sharing.NewExportService(cryptoSvc, deviceKeyMgr)
	imp := sharing.NewImportService(cryptoSvc, deviceKeyMgr, secretsSvc)
	deps := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), Crypto: cryptoSvc, KeyGen: cryptoSvc}

	// Start a persistent advertising transport (discovery=true) so this instance is visible.
	advCtx, cancel := context.WithCancel(context.Background())
	lanTr, err := transport.Build(advCtx, "lan", map[string]any{"listen_addr": ":0", "discovery": true}, desc, deps)
	if err != nil {
		cancel() // fallback silently; continue without advertisement
		lanTr = nil
	}
	var discSession service.DiscoverySession
	if dt, ok := lanTr.(transport.DiscoverableTransport); ok {
		discSession = &lanDiscoverySession{tr: dt}
	}
	transfer := service.NewSharingTransferService(desc, deps, exportAdapter{exp}, imp, discSession, nil)
	var closer func()
	if lanTr != nil {
		closer = func() { _ = lanTr.Close(); cancel() }
	} else {
		cancel()
	}
	return &SharingBundle{SecretsService: secretsSvc, DeviceKeyManager: deviceKeyMgr, ExportService: exp, ImportService: imp, TransferService: transfer, DeviceDescriptor: desc, AdvertiseCloser: closer}, nil
}

// lanDiscoverySession adapts a running LAN transport to the DiscoverySession interface.
type lanDiscoverySession struct{ tr transport.BundleTransport }

func (s *lanDiscoverySession) Devices() []transport.DeviceDescriptor {
	if s == nil || s.tr == nil {
		return nil
	}
	if d, ok := s.tr.(transport.DiscoverableTransport); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		logger.Debug("lan.discoverySession: calling transport.Discover (2s timeout)")
		res, err := d.Discover(ctx, 50)
		if err != nil {
			logger.Debug("lan.discoverySession: discover error: " + err.Error())
		}
		logger.Debug("lan.discoverySession: devices returned=" + strconv.Itoa(len(res)))
		return res
	}
	return nil
}
