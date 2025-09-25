package bluetooth

import (
	"context"
	"go-password-manager/internal/transport/bluetooth/adapters"
)

// sysAdapter wraps a platform key and delegates ConnectToDevice calls
// into the adapters package. The returned connection value must implement
// the bluetooth.Conn interface (checked dynamically at call time).
type sysAdapter struct{ platform string }

func (s *sysAdapter) ConnectToDevice(ctx context.Context, deviceID string) (Conn, error) {
	v, err := adapters.ConnectToDevice(s.platform, ctx, deviceID)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	if c, ok := v.(Conn); ok {
		return c, nil
	}
	return nil, nil
}

// GetSystemAdapter delegates to the adapters subpackage and returns a value
// implementing the bluetooth.Adapter interface, or an error if not available.
func GetSystemAdapter(name string) (Adapter, error) {
	v, err := adapters.GetSystemAdapter(name)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	// The adapters package manages platform initialization and advertises
	// that an adapter is available by returning a non-nil value from its
	// factory. We return a wrapper that delegates ConnectToDevice calls
	// into the adapters package via the exported ConnectToDevice helper.
	return &sysAdapter{platform: name}, nil
}
