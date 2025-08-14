package domain

type SecretView struct {
	SecretName     string
	Type           string
	SecretValueEnc string
	// Revealed is UI-only, do NOT include in JSON tags or persistence
	Revealed bool `json:"-"`
}
