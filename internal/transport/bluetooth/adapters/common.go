package adapters

import (
	"context"
	"fmt"
	"go-password-manager/internal/logger"
	"runtime"
)

// Adapter factory registry allows OS-specific adapter files to register
// a factory function without exporting a common symbol that would collide
// across build tags.
var (
	adapterRegistry   = map[string]func(string) (interface{}, error){}
	connectorRegistry = map[string]func(ctx context.Context, deviceID string) (interface{}, error){}
	scannerRegistry   = map[string]func(serviceUUID string, limit int) []string{}
)

// RegisterAdapter registers a factory for the given platform key.
func RegisterAdapter(key string, factory func(string) (interface{}, error)) {
	adapterRegistry[key] = factory
}

// GetSystemAdapter looks up a registered factory and invokes it. If no
// factory is registered for the given key, it returns (nil, nil).
func GetSystemAdapter(name string) (interface{}, error) {
	if name == "" {
		name = runtime.GOOS
	}
	if f, ok := adapterRegistry[name]; ok {
		return f(name)
	}
	return nil, nil
}

// ConnectToDevice provides a thin exported helper that calls the
// platform-specific connect function for the given platform name. This
// avoids import cycles by keeping the platform dialing logic inside the
// adapters package while allowing the parent bluetooth package to invoke
// it when constructing an Adapter wrapper.
func ConnectToDevice(name string, ctx context.Context, deviceID string) (interface{}, error) {
	if name == "" {
		name = runtime.GOOS
	}
	logger.Debug(fmt.Sprintf("bluetooth.adapters: ConnectToDevice platform=%s device=%s", name, deviceID))
	if c, ok := connectorRegistry[name]; ok {
		return c(ctx, deviceID)
	}
	logger.Debug("bluetooth.adapters: no connector registered for platform " + name)
	return nil, nil
}

// RegisterConnector allows platform-specific files to register a connect
// helper that dials a remote device. This keeps connector implementations
// in their OS-specific files while exposing a stable call-site for the
// parent bluetooth package.
func RegisterConnector(key string, connector func(ctx context.Context, deviceID string) (interface{}, error)) {
	connectorRegistry[key] = connector
}

// RegisterScanner registers a platform-specific scan helper.
func RegisterScanner(key string, scanner func(serviceUUID string, limit int) []string) {
	scannerRegistry[key] = scanner
}

// Scan invokes the registered platform scanner for the given platform key.
// If no scanner is registered, returns an empty slice.
func Scan(name string, serviceUUID string, limit int) []string {
	if name == "" {
		name = runtime.GOOS
	}
	logger.Debug("bluetooth.adapters: Scan platform=" + name + " serviceUUID=" + serviceUUID)
	if s, ok := scannerRegistry[name]; ok {
		return s(serviceUUID, limit)
	}
	logger.Debug("bluetooth.adapters: no scanner registered for platform " + name)
	return nil
}
