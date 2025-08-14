package versioning

import (
	"errors"
	"sync"
)

// Version represents a single version of a secret.
type Version struct {
	Value     string
	Timestamp int64
}

// SecretVersioning manages versions of a secret.
type SecretVersioning struct {
	versions map[string][]Version
	mu       sync.RWMutex
}

// NewSecretVersioning creates a new SecretVersioning instance.
func NewSecretVersioning() *SecretVersioning {
	return &SecretVersioning{
		versions: make(map[string][]Version),
	}
}

// AddVersion adds a new version for a given secret.
func (sv *SecretVersioning) AddVersion(secretKey string, value string, timestamp int64) {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	sv.versions[secretKey] = append(sv.versions[secretKey], Version{Value: value, Timestamp: timestamp})
}

// GetLatestVersion retrieves the latest version of a secret.
func (sv *SecretVersioning) GetLatestVersion(secretKey string) (Version, error) {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	versions, exists := sv.versions[secretKey]
	if !exists || len(versions) == 0 {
		return Version{}, errors.New("no versions found for the given secret key")
	}
	return versions[len(versions)-1], nil
}

// GetVersion retrieves a specific version of a secret by index.
func (sv *SecretVersioning) GetVersion(secretKey string, index int) (Version, error) {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	versions, exists := sv.versions[secretKey]
	if !exists || index < 0 || index >= len(versions) {
		return Version{}, errors.New("invalid version index")
	}
	return versions[index], nil
}

// ListVersions lists all versions of a secret.
func (sv *SecretVersioning) ListVersions(secretKey string) ([]Version, error) {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	versions, exists := sv.versions[secretKey]
	if !exists {
		return nil, errors.New("no versions found for the given secret key")
	}
	return versions, nil
}
