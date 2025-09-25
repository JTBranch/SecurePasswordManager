//go:build windows
// +build windows

package bluetooth

import (
	"context"
	"errors"

	"go-password-manager/internal/transport"
)

// On Windows we intentionally disable the bluetooth transport. Register a
// stub factory so code that enumerates transports still sees "bluetooth",
// but attempting to construct it returns a clear error.
type windowsFactory struct{}

func (windowsFactory) New(ctx context.Context, cfg map[string]any, local transport.DeviceDescriptor, deps transport.Dependencies) (transport.BundleTransport, error) {
	return nil, errors.New("bluetooth transport is not supported on Windows in this build")
}

func init() {
	transport.Register("bluetooth", windowsFactory{})
}
