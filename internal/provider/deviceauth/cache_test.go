package deviceauth_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/deviceauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestCache_SaveAndLoad(t *testing.T) {
	t.Setenv("GAMEFABRIC_CACHE_DIR", t.TempDir())

	expiry := time.Now().Add(time.Hour).Truncate(time.Second)
	tok := &oauth2.Token{
		AccessToken:  "access-abc",
		TokenType:    "bearer",
		RefreshToken: "refresh-xyz",
		Expiry:       expiry,
	}

	require.NoError(t, deviceauth.Save("example.gamefabric.dev", tok))

	got, err := deviceauth.Load("example.gamefabric.dev")
	require.NoError(t, err)

	assert.Equal(t, tok.AccessToken, got.AccessToken)
	assert.Equal(t, tok.TokenType, got.TokenType)
	assert.Equal(t, tok.RefreshToken, got.RefreshToken)
	assert.True(t, tok.Expiry.Equal(got.Expiry), "expiry mismatch: want %v got %v", tok.Expiry, got.Expiry)
}

func TestCache_Load_MissingFile(t *testing.T) {
	t.Setenv("GAMEFABRIC_CACHE_DIR", t.TempDir())

	_, err := deviceauth.Load("no-such-host.gamefabric.dev")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no cached token")
}

func TestCache_Load_CorruptFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GAMEFABRIC_CACHE_DIR", dir)

	// Write garbage where the credentials file would live.
	credPath := filepath.Join(dir, "gamefabric", "corrupt.gamefabric.dev", "credentials.json")
	require.NoError(t, deviceauth.Save("corrupt.gamefabric.dev", &oauth2.Token{AccessToken: "x"}))

	// Overwrite the file with invalid JSON.
	require.NoError(t, os.WriteFile(credPath, []byte("not-json"), 0o600))

	_, err := deviceauth.Load("corrupt.gamefabric.dev")
	require.Error(t, err)
}

func TestCache_Save_HostSanitization(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GAMEFABRIC_CACHE_DIR", dir)

	// A host with characters that are illegal in directory names on some OSes.
	host := "host:8080/path"
	require.NoError(t, deviceauth.Save(host, &oauth2.Token{AccessToken: "tok"}))

	// Must load back without error — path was sanitized correctly.
	got, err := deviceauth.Load(host)
	require.NoError(t, err)
	assert.Equal(t, "tok", got.AccessToken)

	// Verify no literal ':' or '/' appear in the cache subdirectory name.
	matches, err := filepath.Glob(filepath.Join(dir, "gamefabric", "*", "credentials.json"))
	require.NoError(t, err)
	require.Len(t, matches, 1)

	dir2 := filepath.Dir(matches[0])
	name := filepath.Base(dir2)
	assert.NotContains(t, name, ":")
	assert.NotContains(t, name, "/")
}
