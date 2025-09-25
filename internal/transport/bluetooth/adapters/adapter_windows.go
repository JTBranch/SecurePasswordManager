//go:build windows
// +build windows

package adapters

// Intentionally do not register a Windows adapter. Bluetooth sharing is
// supported on Linux and macOS only. Per-OS adapter implementations register
// themselves via RegisterAdapter in their build-tagged files; we leave the
// Windows file present as a no-op to make the intent explicit.
func init() {
	// intentionally empty: Windows Bluetooth adapter not implemented.
}
