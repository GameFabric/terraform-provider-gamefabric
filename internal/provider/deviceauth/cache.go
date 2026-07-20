// Package deviceauth provides helpers for caching OAuth2 tokens obtained via
// the Device Authorization Flow (RFC 8628). Tokens are stored on disk so that
// subsequent Terraform runs can reuse a valid access/refresh token without
// prompting the user again.
package deviceauth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

type tokenOnDisk struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Expiry       time.Time `json:"expiry,omitzero"`
}

func Load(host string) (*oauth2.Token, error) {
	path, err := tokenFile(host)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("no cached token")
		}
		return nil, err
	}

	var d tokenOnDisk
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  d.AccessToken,
		TokenType:    d.TokenType,
		RefreshToken: d.RefreshToken,
		Expiry:       d.Expiry,
	}, nil
}

// Save writes the token to disk.
func Save(host string, tok *oauth2.Token) error {
	path, err := tokenFile(host)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.Marshal(tokenOnDisk{ //nolint:gosec // intentionally marshaling credential fields to the local cache file
		AccessToken:  tok.AccessToken,
		TokenType:    tok.TokenType,
		RefreshToken: tok.RefreshToken,
		Expiry:       tok.Expiry,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func tokenFile(host string) (string, error) {
	base := os.Getenv("GAMEFABRIC_CACHE_DIR")
	if base == "" {
		cfgDir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		base = cfgDir
	}

	// Sanitise the host so it is safe to use as a directory name.
	safe := strings.NewReplacer(":", "_", "/", "_", "\\", "_").Replace(host)
	return filepath.Join(base, "gamefabric", safe, "credentials.json"), nil
}
