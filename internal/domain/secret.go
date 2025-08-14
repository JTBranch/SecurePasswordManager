package domain

// SecretType represents the type of secret being stored
type SecretType string

// Secret type constants
const (
	// SecretTypeKeyValue represents a simple key-value pair secret
	SecretTypeKeyValue SecretType = "key_value"
	// SecretTypeJSON represents a JSON-formatted secret
	SecretTypeJSON SecretType = "json"
	// SecretTypeOther represents any other type of secret
	SecretTypeOther SecretType = "other"
)

// SecretVersion represents a specific version of a secret with its encrypted value
type SecretVersion struct {
	SecretValueEnc string `json:"secretValueEnc"`
	Version        int    `json:"version"`
	UpdatedAt      string `json:"updatedAt"`
	UpdatedBy      string `json:"updatedBy,omitempty"`
}

// Secret represents a secret with its metadata and version history
type Secret struct {
	SecretName     string          `json:"secretName"`
	Type           SecretType      `json:"type"`
	CurrentVersion int             `json:"currentVersion"`
	Versions       []SecretVersion `json:"versions"`
}

// GetCurrentVersion returns the current (latest) version of the secret
func (s *Secret) GetCurrentVersion() *SecretVersion {
	for _, version := range s.Versions {
		if version.Version == s.CurrentVersion {
			return &version
		}
	}
	return nil
}

// GetVersionsSorted returns all versions sorted by version number (latest first)
func (s *Secret) GetVersionsSorted() []SecretVersion {
	versions := make([]SecretVersion, len(s.Versions))
	copy(versions, s.Versions)

	// Sort by version number (latest first)
	for i := 0; i < len(versions)-1; i++ {
		for j := i + 1; j < len(versions); j++ {
			if versions[i].Version < versions[j].Version {
				versions[i], versions[j] = versions[j], versions[i]
			}
		}
	}

	return versions
}

// SecretsFile represents the file structure for storing secrets
type SecretsFile struct {
	AppVersion  string   `json:"appVersion"`
	AppUser     string   `json:"appUser"`
	LastUpdated string   `json:"lastUpdated"`
	Secrets     []Secret `json:"secrets"`
}

// SecretView represents a UI view model for displaying secrets
type SecretView struct {
	SecretName     string
	Type           string
	SecretValueEnc string
	// Revealed is UI-only, do NOT include in JSON tags or persistence
	Revealed bool `json:"-"`
}
