package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"spotumn/internal/config"
	"golang.org/x/oauth2"
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

func TestRedirectURIDefault(t *testing.T) {
	cfg := &config.Config{
		Port: 8989,
	}
	svc := NewAuthService(cfg)
	if svc.oauthCfg.RedirectURL != "http://127.0.0.1:8989/login" {
		t.Errorf("expected redirect URL 'http://127.0.0.1:8989/login', got '%s'", svc.oauthCfg.RedirectURL)
	}
	if svc.oauthCfg.ClientID != config.SpotifyClientID {
		t.Errorf("expected built-in ClientID '%s', got '%s'", config.SpotifyClientID, svc.oauthCfg.ClientID)
	}

	cfgCustom := &config.Config{
		Port:        8989,
		RedirectURI: "http://127.0.0.1:8989/custom",
	}
	svcCustom := NewAuthService(cfgCustom)
	if svcCustom.oauthCfg.RedirectURL != "http://127.0.0.1:8989/custom" {
		t.Errorf("expected redirect URL 'http://127.0.0.1:8989/custom', got '%s'", svcCustom.oauthCfg.RedirectURL)
	}
}

func TestAccountManager(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	mgr := NewAccountManager()

	tok1 := &oauth2.Token{AccessToken: "token1"}
	tok2 := &oauth2.Token{AccessToken: "token2"}
	tok3 := &oauth2.Token{AccessToken: "token3"}
	tok4 := &oauth2.Token{AccessToken: "token4"}
	tok5 := &oauth2.Token{AccessToken: "token5"}

	// Add accounts 1 to 4
	if _, err := mgr.AddAccount("user1", "User 1", tok1); err != nil {
		t.Fatalf("failed to add account 1: %v", err)
	}
	if _, err := mgr.AddAccount("user2", "User 2", tok2); err != nil {
		t.Fatalf("failed to add account 2: %v", err)
	}
	if _, err := mgr.AddAccount("user3", "User 3", tok3); err != nil {
		t.Fatalf("failed to add account 3: %v", err)
	}
	if _, err := mgr.AddAccount("user4", "User 4", tok4); err != nil {
		t.Fatalf("failed to add account 4: %v", err)
	}

	accounts := mgr.GetAccounts()
	if len(accounts) != 4 {
		t.Fatalf("expected 4 accounts, got %d", len(accounts))
	}

	// 5th account should be rejected
	if _, err := mgr.AddAccount("user5", "User 5", tok5); err == nil {
		t.Fatal("expected error when adding 5th account")
	}

	// Switch account to index 1 (User 2)
	switched, err := mgr.SwitchAccount(1)
	if err != nil {
		t.Fatalf("failed to switch account: %v", err)
	}
	if switched.ID != "user2" {
		t.Errorf("expected switched account ID 'user2', got %s", switched.ID)
	}
	if mgr.GetActiveIndex() != 1 {
		t.Errorf("expected active index 1, got %d", mgr.GetActiveIndex())
	}
}

func TestRefreshTokenPreservation(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	cfg := &config.Config{Port: 8989}
	svc := NewAuthService(cfg)

	originalTok := &oauth2.Token{
		AccessToken:  "orig_access",
		RefreshToken: "keep_this_refresh",
	}
	if err := svc.SaveToken(originalTok); err != nil {
		t.Fatalf("failed to save initial token: %v", err)
	}

	// Spotify refresh responses omit refresh_token:
	refreshedWithoutRefresh := &oauth2.Token{
		AccessToken: "new_access",
	}
	if err := svc.SaveToken(refreshedWithoutRefresh); err != nil {
		t.Fatalf("failed to save refreshed token: %v", err)
	}

	loaded, err := svc.LoadSavedToken()
	if err != nil {
		t.Fatalf("failed to load token: %v", err)
	}
	if loaded.AccessToken != "new_access" {
		t.Errorf("expected access token 'new_access', got '%s'", loaded.AccessToken)
	}
	if loaded.RefreshToken != "keep_this_refresh" {
		t.Errorf("expected refresh token 'keep_this_refresh' to be preserved, got '%s'", loaded.RefreshToken)
	}
}


