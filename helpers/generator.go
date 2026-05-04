package helpers

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateSessionID returns a 256-bit random token encoded as hexadecimal.
func GenerateSessionID() (string, error) {
	b := make([]byte, 32) // 256-bit session ID

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil

}
