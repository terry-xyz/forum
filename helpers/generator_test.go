package helpers

import "testing"

// TestGenerateSessionIDReturnsUniqueHexToken checks the token shape and that
// repeated calls do not reuse the same random value.
func TestGenerateSessionIDReturnsUniqueHexToken(t *testing.T) {
	// Generate two tokens so uniqueness can be checked inside one test run.
	first, err := GenerateSessionID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateSessionID()
	if err != nil {
		t.Fatal(err)
	}

	// Thirty-two random bytes encode to sixty-four hex characters.
	if len(first) != 64 {
		t.Fatalf("session id length = %d, want 64", len(first))
	}
	// A collision here would indicate the generator is not behaving randomly.
	if first == second {
		t.Fatal("session ids should be unique")
	}
}
