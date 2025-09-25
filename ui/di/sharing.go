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
	"go-password-manager/internal/transport/bluetooth"
	"os"
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
	env := os.Getenv("GO_PASSWORD_MANAGER_ENV")
	inMem := buildCfg.Security.Keyring.InMemory || buildCfg.IsTest() || env == "test"
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
		// Allow environment overrides for containerized runs so advertised address is reachable.
		advertised := os.Getenv("VPM_ADVERTISED_ADDR")
		instanceName := os.Getenv("VPM_INSTANCE_NAME")
		deviceName := deviceKeyMgr.GetDeviceName()
		if instanceName != "" {
			deviceName = instanceName
		}
		desc = transport.DeviceDescriptor{DeviceID: encKey.ID, UserID: deviceKeyMgr.GetAppUser(), DeviceName: deviceName, Ed25519Pub: sigKey.PublicKey, X25519Pub: encKey.PublicKey}
		if advertised != "" {
			desc.LastAddr = advertised
		}
	}
	exp := sharing.NewExportService(cryptoSvc, deviceKeyMgr)
	imp := sharing.NewImportService(cryptoSvc, deviceKeyMgr, secretsSvc)
	deps := transport.Dependencies{Registry: transport.NewInMemoryRegistry(), Crypto: cryptoSvc, KeyGen: cryptoSvc}
	// Attempt to obtain a system Bluetooth adapter and inject it if present.
	if adapt, err := bluetooth.GetSystemAdapter(""); err == nil && adapt != nil {
		deps.BluetoothAdapter = adapt
		logger.Debug("bluetooth: system adapter available, injected into transport dependencies")
	} else if err != nil {
		logger.Warn("bluetooth: failed to get system adapter", err.Error())
	} else {
		logger.Debug("bluetooth: no system adapter registered for this platform")
	}

	// Start persistent advertising transports so this instance is visible for supported transports.
	advCtx, cancel := context.WithCancel(context.Background())
	// Track started transports and their closers so we can clean up on exit.
	var started []transport.BundleTransport
	var discables []transport.DiscoverableTransport

	lanTr, err := transport.Build(advCtx, "lan", map[string]any{"listen_addr": ":0", "discovery": true}, desc, deps)
	if err == nil {
		started = append(started, lanTr)
		if dt, ok := lanTr.(transport.DiscoverableTransport); ok {
			discables = append(discables, dt)
		}
	}

	// Attempt to start bluetooth advertising transport if available (will fail harmlessly if BluetoothAdapter missing).
	btTr, berr := transport.Build(advCtx, "bluetooth", map[string]any{"listen_addr": ":0", "discovery": true}, desc, deps)
	if berr == nil {
		started = append(started, btTr)
		if dt, ok := btTr.(transport.DiscoverableTransport); ok {
			discables = append(discables, dt)
		}
	}

	// Compose a discovery session that merges results from all started discoverable transports.
	var discSession service.DiscoverySession
	if len(discables) == 1 {
		discSession = &lanDiscoverySession{tr: discables[0]}
	} else if len(discables) > 1 {
		discSession = &multiDiscoverySession{trs: discables}
	}

	transfer := service.NewSharingTransferService(desc, deps, exportAdapter{exp}, imp, discSession, nil)
	var closer func()
	if len(started) > 0 {
		closer = func() {
			for _, tr := range started {
				_ = tr.Close()
			}
			cancel()
		}
	} else {
		cancel()
	}
	return &SharingBundle{SecretsService: secretsSvc, DeviceKeyManager: deviceKeyMgr, ExportService: exp, ImportService: imp, TransferService: transfer, DeviceDescriptor: desc, AdvertiseCloser: closer}, nil
}

// multiDiscoverySession merges devices from multiple DiscoverableTransport instances.
type multiDiscoverySession struct {
	trs []transport.DiscoverableTransport
}

func (m *multiDiscoverySession) Devices() []transport.DeviceDescriptor {
	out := make([]transport.DeviceDescriptor, 0)
	for _, t := range m.trs {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		_ = ctx
		// call Discover on each transport with a short timeout; ignore errors per-transport
		if d, ok := t.(transport.DiscoverableTransport); ok {
			devs, _ := d.Discover(context.Background(), 50)
			out = append(out, devs...)
		}
		cancel()
	}
	return out
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
