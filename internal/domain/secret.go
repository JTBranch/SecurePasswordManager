package domain

type SecretType string

const (
    SecretTypeKeyValue SecretType = "key_value"
    SecretTypeJSON     SecretType = "json"
    SecretTypeOther    SecretType = "other"
)

type Secret struct {
    SecretName     string     `json:"secretName"`
    SecretValueEnc string     `json:"secretValueEnc"`
    Type           SecretType `json:"type"`
    Version        int        `json:"version"`
    UpdatedAt      string     `json:"updatedAt"`
    UpdatedBy      string     `json:"updatedBy,omitempty"`
}

type SecretsFile struct {
    AppVersion  string   `json:"appVersion"`
    AppUser     string   `json:"appUser"`
    LastUpdated string   `json:"lastUpdated"`
    Secrets     []Secret `json:"secrets"`
}

type SecretView struct {
    SecretName            string
    Type           string
    SecretValueEnc string
    // Revealed is UI-only, do NOT include in JSON tags or persistence
    Revealed       bool `json:"-"`
}