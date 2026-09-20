package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestGeneratePKCE(t *testing.T) {
	verifier, challenge, err := generatePKCE()
	if err != nil {
		t.Fatalf("unexpected error generating PKCE: %v", err)
	}

	if len(verifier) < 43 || len(verifier) > 128 {
		t.Fatalf("code verifier length %d invalid for RFC 7636", len(verifier))
	}

	// Verify challenge is S256(verifier)
	h := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(h[:])

	if challenge != expectedChallenge {
		t.Fatalf("challenge mismatch: expected %s, got %s", expectedChallenge, challenge)
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	b1, err := GenerateRandomBytes(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b2, err := GenerateRandomBytes(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(b1) != 16 || len(b2) != 16 {
		t.Fatal("expected length 16")
	}

	// Two random outputs should not collide
	allEqual := true
	for i := range b1 {
		if b1[i] != b2[i] {
			allEqual = false
			break
		}
	}
	if allEqual {
		t.Fatal("two random byte slices should not be identical")
	}
}
