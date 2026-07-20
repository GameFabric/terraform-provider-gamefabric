package cache

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// DiskCache is a disk backed cache for OAuth2 tokens. It stores tokens in a directory on disk so that
// subsequent runs can reuse a valid access/refresh token without prompting the user again.
type DiskCache struct {
	basePath string
}

// NewDiskCache creates a new DiskCache instance. It uses the GAMEFABRIC_CACHE_DIR environment variable as the base path for storing tokens.
// If the environment variable is not set, it defaults to the user's configuration directory.
func NewDiskCache() (*DiskCache, error) {
	basePath := os.Getenv("GAMEFABRIC_CACHE_DIR")
	if basePath == "" {
		cfgDir, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		basePath = cfgDir
	}
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}
	basePath = filepath.Join(filepath.Clean(absPath), "gamefabric")

	return &DiskCache{basePath: basePath}, nil
}

type token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Expiry       time.Time `json:"expiry,omitzero"`
}

// Load retrieves the OAuth2 token for the given host from the disk cache. If the token does not exist,
// it returns an error indicating that there is no cached token.
func (c *DiskCache) Load(host string) (*oauth2.Token, error) {
	path := c.tokenFile(host)
	data, err := os.ReadFile(path) //nolint:gosec // This is not a security issue; we are just reading a file from disk.
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("no cached token")
		}
		return nil, err
	}

	var d token
	if err = json.Unmarshal(data, &d); err != nil {
		return nil, err
	}

	return &oauth2.Token{
		AccessToken:  d.AccessToken,
		TokenType:    d.TokenType,
		RefreshToken: d.RefreshToken,
		Expiry:       d.Expiry,
	}, nil
}

// Save writes the given OAuth2 token for the specified host to the disk cache. It creates the necessary directories if they do not exist.
func (c *DiskCache) Save(host string, tok *oauth2.Token) error {
	path := c.tokenFile(host)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.Marshal(token{ //nolint:gosec // This is not a security issue; we are just serializing a struct to JSON.
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

func (c *DiskCache) tokenFile(host string) string {
	safeHost := strings.NewReplacer(":", "_", "/", "_", "\\", "_").Replace(host)
	return filepath.Clean(filepath.Join(c.basePath, safeHost+".json"))
}
