package helpers

import "testing"

func TestGenerateSessionIDReturnsUniqueHexToken(t *testing.T) {
	first, err := GenerateSessionID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateSessionID()
	if err != nil {
		t.Fatal(err)
	}

	if len(first) != 64 {
		t.Fatalf("session id length = %d, want 64", len(first))
	}
	if first == second {
		t.Fatal("session ids should be unique")
	}
}
