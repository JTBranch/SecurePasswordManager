package domain

type SecretType string

const (
	SecretTypeKeyValue SecretType = "key_value"
	SecretTypeJSON     SecretType = "json"
	SecretTypeOther    SecretType = "other"
)

type SecretVersion struct {
	SecretValueEnc string `json:"secretValueEnc"`
	Version        int    `json:"version"`
	UpdatedAt      string `json:"updatedAt"`
	UpdatedBy      string `json:"updatedBy,omitempty"`
}

type Secret struct {
	SecretName     string          `json:"secretName"`
	Type           SecretType      `json:"type"`
	CurrentVersion int             `json:"currentVersion"`
	Versions       []SecretVersion `json:"versions"`
}

// Helper method to get the current (latest) version
func (s *Secret) GetCurrentVersion() *SecretVersion {
	for _, version := range s.Versions {
		if version.Version == s.CurrentVersion {
			return &version
		}
	}
	return nil
}

// Helper method to get all versions sorted by version number (latest first)
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

type SecretsFile struct {
	AppVersion  string   `json:"appVersion"`
	AppUser     string   `json:"appUser"`
	LastUpdated string   `json:"lastUpdated"`
	Secrets     []Secret `json:"secrets"`
}

type SecretView struct {
	SecretName     string
	Type           string
	SecretValueEnc string
	// Revealed is UI-only, do NOT include in JSON tags or persistence
	Revealed bool `json:"-"`
}
