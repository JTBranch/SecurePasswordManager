package domain

// SecretView represents a UI-specific view of a secret
type SecretView struct {
	SecretName     string
	Type           string
	SecretValueEnc string
	// Revealed is UI-only, do NOT include in JSON tags or persistence
	Revealed bool `json:"-"`
}
