package crypto

// Zeroize overwrites the provided byte slice with zeros.
// It is a best-effort memory clearing; the compiler may not always
// guarantee wiping, but keeping a single pattern central aids auditing.
func Zeroize(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
