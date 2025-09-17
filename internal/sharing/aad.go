package sharing

// buildKeyBoxAAD constructs AAD: bundleID || 0x00 || signingPub
func buildKeyBoxAAD(bundleID string, signingPub []byte) []byte {
	aad := make([]byte, 0, len(bundleID)+1+len(signingPub))
	aad = append(aad, []byte(bundleID)...)
	aad = append(aad, 0x00)
	aad = append(aad, signingPub...)
	return aad
}

// buildSecretsAAD constructs AAD: bundleID || 0x01 || signingPub
func buildSecretsAAD(bundleID string, signingPub []byte) []byte {
	aad := make([]byte, 0, len(bundleID)+1+len(signingPub))
	aad = append(aad, []byte(bundleID)...)
	aad = append(aad, 0x01)
	aad = append(aad, signingPub...)
	return aad
}
