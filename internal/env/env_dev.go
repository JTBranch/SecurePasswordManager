//go:build dev
// +build dev

package env

// Override default for development builds
func init() {
	// This will be called before Load() and can modify defaults
	defaultDebugLogging = true
}
