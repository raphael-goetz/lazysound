package api

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	Scope        string    `json:"scope,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
}

func (t Token) Expired(skew time.Duration) bool {
	if t.AccessToken == "" {
		return true
	}
	if t.ExpiresAt.IsZero() {
		// If unknown, treat as expired to be safe.
		return true
	}
	return time.Now().After(t.ExpiresAt.Add(-skew))
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

func (r tokenResponse) toToken(now time.Time) Token {
	t := Token{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		ExpiresIn:    r.ExpiresIn,
		Scope:        r.Scope,
		TokenType:    r.TokenType,
	}
	if r.ExpiresIn > 0 {
		t.ExpiresAt = now.Add(time.Duration(r.ExpiresIn) * time.Second)
	}
	return t
}

type TokenStore struct {
	path string
}

func NewTokenStoreDefault() (*TokenStore, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	base := filepath.Join(dir, ".soundcli")
	if err := os.MkdirAll(base, 0o700); err != nil {
		return nil, err
	}
	return &TokenStore{path: filepath.Join(base, "token.json")}, nil
}

func (s *TokenStore) Save(t Token) error {
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *TokenStore) Load() (Token, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return Token{}, err
	}
	var t Token
	if err := json.Unmarshal(b, &t); err != nil {
		return Token{}, err
	}
	if t.AccessToken == "" {
		return Token{}, errors.New("empty token store")
	}
	return t, nil
}

