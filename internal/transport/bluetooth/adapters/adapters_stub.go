//go:build !linux
// +build !linux

package adapters

// Non-Linux placeholder: platform-specific adapters register themselves via
// RegisterAdapter in their OS-specific files. We keep this file so the
// adapters package exists on all platforms, but no global GetSystemAdapter
// symbol is declared here (the registry in common.go provides lookup).
func init() {
	// intentionally empty: non-Linux platforms rely on OS-specific adapters
}
