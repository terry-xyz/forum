package helpers

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateSessionID returns a 256-bit random token encoded as hexadecimal.
func GenerateSessionID() (string, error) {
	// Allocate 32 random bytes. Once hex-encoded, this becomes a 64-character
	// string that is still backed by 256 bits of entropy.
	b := make([]byte, 32)

	// crypto/rand reads from the operating system CSPRNG, which is appropriate
	// for values that may become authentication/session tokens.
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Hex keeps the token URL- and cookie-friendly without losing entropy.
	return hex.EncodeToString(b), nil

}
