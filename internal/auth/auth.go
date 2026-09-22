package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"spotumn/internal/backend"
	"spotumn/internal/config"

	"golang.org/x/oauth2"
)

const (
	SpotifyAuthURL  = "https://accounts.spotify.com/authorize"
	SpotifyTokenURL = "https://accounts.spotify.com/api/token"
)

var Scopes = []string{
	"user-read-private",
	"user-read-email",
	"user-read-playback-state",
	"user-modify-playback-state",
	"user-read-currently-playing",
	"playlist-read-private",
	"playlist-read-collaborative",
	"user-library-read",
	"user-follow-read",
	"user-read-playback-position",
	"user-top-read",
	"user-read-recently-played",
	"streaming",
}

type AuthService struct {
	cfg        *config.Config
	oauthCfg   *oauth2.Config
	tokenMu    sync.RWMutex
	token      *oauth2.Token
	tokenFile  string
	httpClient *http.Client
}

func NewAuthService(cfg *config.Config) *AuthService {
	redirectURI := cfg.RedirectURI
	if redirectURI == "" {
		redirectURI = fmt.Sprintf("http://127.0.0.1:%d/login", cfg.Port)
	}
	oauthCfg := &oauth2.Config{
		ClientID: config.SpotifyClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  SpotifyAuthURL,
			TokenURL: SpotifyTokenURL,
		},
		RedirectURL: redirectURI,
		Scopes:      Scopes,
	}

	return &AuthService{
		cfg:       cfg,
		oauthCfg:  oauthCfg,
		tokenFile: filepath.Join(config.GetDir(), "credentials.json"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GenerateRandomBytes creates cryptographically secure random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// generatePKCE generates code_verifier and code_challenge (RFC 7636)
func generatePKCE() (verifier, challenge string, err error) {
	bytes, err := GenerateRandomBytes(64)
	if err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(bytes)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return verifier, challenge, nil
}

// LoadSavedToken loads the token from credentials.json
func (a *AuthService) LoadSavedToken() (*oauth2.Token, error) {
	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()

	data, err := os.ReadFile(a.tokenFile)
	if err != nil {
		return nil, err
	}

	var tok oauth2.Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, err
	}

	a.token = &tok
	return &tok, nil
}

// SaveToken saves token with strict 0600 permissions
func (a *AuthService) SaveToken(tok *oauth2.Token) error {
	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()

	a.token = tok
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}

	// 0600 mode prevents other local users from reading credentials
	return os.WriteFile(a.tokenFile, data, 0600)
}

// GetTokenSource returns a refreshed token source that auto-updates on disk
func (a *AuthService) GetTokenSource(ctx context.Context) oauth2.TokenSource {
	a.tokenMu.RLock()
	current := a.token
	a.tokenMu.RUnlock()

	ctx = context.WithValue(ctx, oauth2.HTTPClient, a.httpClient)
	ts := a.oauthCfg.TokenSource(ctx, current)

	return oauth2.ReuseTokenSource(current, &savingTokenSource{
		src:  ts,
		save: a.SaveToken,
	})
}

type savingTokenSource struct {
	src  oauth2.TokenSource
	save func(*oauth2.Token) error
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	tok, err := s.src.Token()
	if err != nil {
		return nil, err
	}
	_ = s.save(tok)
	return tok, nil
}

// Authorize performs the PKCE authorization flow using a local loopback server
func (a *AuthService) Authorize(ctx context.Context) (*oauth2.Token, error) {
	// Try loading saved token first
	if tok, err := a.LoadSavedToken(); err == nil && tok != nil {
		if tok.Valid() {
			return tok, nil
		}
		// If token is expired but has a refresh token, silently refresh without opening browser!
		if tok.RefreshToken != "" {
			ctxWithHTTP := context.WithValue(ctx, oauth2.HTTPClient, a.httpClient)
			ts := a.oauthCfg.TokenSource(ctxWithHTTP, tok)
			if refreshed, err := ts.Token(); err == nil && refreshed.Valid() {
				_ = a.SaveToken(refreshed)
				return refreshed, nil
			}
		}
	}

	verifier, challenge, err := generatePKCE()
	if err != nil {
		return nil, fmt.Errorf("failed generating PKCE: %w", err)
	}

	stateBytes, err := GenerateRandomBytes(16)
	if err != nil {
		return nil, fmt.Errorf("failed generating state: %w", err)
	}
	state := hex.EncodeToString(stateBytes)

	// Auth URL with PKCE challenge
	authURL := a.oauthCfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", challenge),
	)

	// Channel to receive the auth code or error
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Bind strictly to loopback IP (127.0.0.1) to avoid exposing callback on LAN
	addr := fmt.Sprintf("127.0.0.1:%d", a.cfg.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to bind local callback on %s: %w", addr, err)
	}

	mux := http.NewServeMux()
	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	callbackHandler := func(w http.ResponseWriter, r *http.Request) {
		// Set secure response headers
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline';")

		q := r.URL.Query()
		if queryState := q.Get("state"); queryState != state {
			http.Error(w, "Invalid state parameter (CSRF detected)", http.StatusBadRequest)
			errChan <- errors.New("state parameter mismatch: potential CSRF attack")
			return
		}

		if errStr := q.Get("error"); errStr != "" {
			fmt.Fprintf(w, "<h3>Authentication error: %s</h3>", html.EscapeString(errStr))
			errChan <- fmt.Errorf("spotify auth error: %s", errStr)
			return
		}

		code := q.Get("code")
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			errChan <- errors.New("no authorization code in callback")
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>spotumn Authorization</title>
<style>
body { font-family: sans-serif; background: #11111b; color: #cdd6f4; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
.card { background: #181825; border: 1px solid #b4befe; border-radius: 12px; padding: 32px 48px; text-align: center; }
h1 { color: #a6e3a1; font-size: 24px; margin-bottom: 8px; }
p { color: #a6adc8; font-size: 14px; }
</style>
</head>
<body>
<div class="card">
  <h1>Authentication Successful</h1>
  <p>You can close this tab and return to <strong>spotumn</strong>.</p>
</div>
</body>
</html>`))

		codeChan <- code
	}

	mux.HandleFunc("/login", callbackHandler)
	mux.HandleFunc("/callback", callbackHandler)
	mux.HandleFunc("/", callbackHandler)

	go func() {
		_ = server.Serve(listener)
	}()

	// Open browser or show URL
	backend.OpenURL(authURL)

	// Wait for callback or context cancel
	select {
	case <-ctx.Done():
		_ = server.Shutdown(context.Background())
		return nil, ctx.Err()
	case err := <-errChan:
		_ = server.Shutdown(context.Background())
		return nil, err
	case code := <-codeChan:
		_ = server.Shutdown(context.Background())

		// Exchange code for token with PKCE verifier
		tok, err := a.exchangePKCE(ctx, code, verifier)
		if err != nil {
			return nil, fmt.Errorf("token exchange failed: %w", err)
		}

		_ = a.SaveToken(tok)
		return tok, nil
	}
}

// exchangePKCE exchanges the auth code with the code_verifier
func (a *AuthService) exchangePKCE(ctx context.Context, code, verifier string) (*oauth2.Token, error) {
	v := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {a.oauthCfg.RedirectURL},
		"client_id":     {config.SpotifyClientID},
		"code_verifier": {verifier},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", SpotifyTokenURL, strings.NewReader(v.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("spotify token error (HTTP %d): %v", resp.StatusCode, errResp)
	}

	var tok oauth2.Token
	var raw struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	tok.AccessToken = raw.AccessToken
	tok.TokenType = raw.TokenType
	tok.RefreshToken = raw.RefreshToken
	tok.Expiry = time.Now().Add(time.Duration(raw.ExpiresIn) * time.Second)

	return &tok, nil
}
