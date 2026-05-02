package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func (c *ApiClient) AuthCodePKCE(ctx context.Context) (Token, error) {
	// PKCE
	verifier, err := randomURLSafeString(64)
	if err != nil {
		return Token{}, err
	}
	challenge := pkceS256Challenge(verifier)

	// CSRF state
	state, err := randomURLSafeString(32)
	if err != nil {
		return Token{}, err
	}

	// Start local callback server
	cbURL, err := url.Parse(c.cfg.RedirectURI)
	if err != nil {
		return Token{}, fmt.Errorf("invalid redirect_uri: %w", err)
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	srv := &http.Server{Handler: http.NewServeMux()}
	mux := srv.Handler.(*http.ServeMux)

	mux.HandleFunc(cbURL.Path, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- errors.New("state mismatch")
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			errCh <- errors.New("missing code in callback")
			return
		}

		io.WriteString(w, "SoundCloud authentication complete. You can close this tab and return to the CLI.\n")
		codeCh <- code
	})

	ln, err := listenOnRedirect(cbURL)
	if err != nil {
		return Token{}, err
	}

	go func() {
		_ = srv.Serve(ln) // closed later
	}()

	// Build authorize URL
	u, err := url.Parse(authURL)
	if err != nil {
		return Token{}, err
	}
	q := u.Query()
	q.Set("client_id", c.cfg.ClientID)
	q.Set("redirect_uri", c.cfg.RedirectURI)
	q.Set("response_type", "code")
	q.Set("code_challenge", challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)
	q.Set("display", "popup")
	if strings.TrimSpace(c.cfg.Scope) != "" {
		q.Set("scope", c.cfg.Scope)
	}
	u.RawQuery = q.Encode()

	// Open browser
	if err := openBrowser(u.String()); err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't open browser automatically.\nOpen this URL manually:\n%s\n\n", u.String())
	} else {
		fmt.Printf("Opening browser for authorization...\nIf it doesn't open, paste this URL:\n%s\n\n", u.String())
	}

	// Wait for callback or error or context cancel
	var code string
	select {
	case <-ctx.Done():
		_ = srv.Close()
		return Token{}, ctx.Err()
	case err := <-errCh:
		_ = srv.Close()
		return Token{}, err
	case code = <-codeCh:
		_ = srv.Close()
	}

	// Exchange code for token
	tok, err := c.exchangeAuthCode(ctx, code, verifier)
	if err != nil {
		return Token{}, err
	}

	if err := c.store.Save(tok); err != nil {
		return Token{}, err
	}
	return tok, nil
}

func (s *ApiClient) exchangeAuthCode(ctx context.Context, code, verifier string) (Token, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", s.cfg.ClientID)
	form.Set("client_secret", s.cfg.ClientSecret)
	form.Set("redirect_uri", s.cfg.RedirectURI)
	form.Set("code_verifier", verifier)
	form.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := s.http.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Token{}, fmt.Errorf("token exchange failed: %s: %s", res.Status, string(body))
	}

	var raw tokenResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return Token{}, fmt.Errorf("invalid token JSON: %w", err)
	}
	return raw.toToken(time.Now()), nil
}

func (s *ApiClient) refreshToken(ctx context.Context, refresh string) (Token, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", s.cfg.ClientID)
	form.Set("client_secret", s.cfg.ClientSecret)
	form.Set("refresh_token", refresh)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := s.http.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return Token{}, fmt.Errorf("refresh failed: %s: %s", res.Status, string(body))
	}

	var raw tokenResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return Token{}, fmt.Errorf("invalid refresh JSON: %w", err)
	}
	return raw.toToken(time.Now()), nil
}

// Helpers

func pkceS256Challenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func randomURLSafeString(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func openBrowser(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "-a", "Brave Browser", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default: // linux, etc.
		cmd = exec.Command("xdg-open", u)
	}
	return cmd.Start()
}

func listenOnRedirect(cbURL *url.URL) (net.Listener, error) {
	host := cbURL.Host
	if host == "" {
		return nil, errors.New("redirect_uri must include host:port, e.g. http://127.0.0.1:7878/callback")
	}
	// Ensure there is a port.
	if !strings.Contains(host, ":") {
		return nil, errors.New("redirect_uri must include an explicit port, e.g. http://127.0.0.1:7878/callback")
	}
	return net.Listen("tcp", host)
}

func (c *ApiClient) EnsureValidToken(ctx context.Context) (Token, error) {
	tok, err := c.store.Load()
	if err != nil {
		return Token{}, fmt.Errorf("not authenticated (run: lazysoundctl auth): %w", err)
	}

	// Refresh a minute early to avoid edge cases.
	if !tok.Expired(60 * time.Second) {
		return tok, nil
	}

	if tok.RefreshToken == "" {
		return Token{}, fmt.Errorf("token expired and no refresh_token available (run: lazysoundctl auth)")
	}

	newTok, err := c.refreshToken(ctx, tok.RefreshToken)
	if err != nil {
		return Token{}, err
	}
	if err := c.store.Save(newTok); err != nil {
		return Token{}, err
	}
	return newTok, nil
}
