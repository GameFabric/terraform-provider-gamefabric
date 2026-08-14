package cache_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestDiskCache_SaveAndLoadToken(t *testing.T) {
	t.Setenv("GAMEFABRIC_CACHE_DIR", t.TempDir())

	disk, err := cache.NewDiskCache()
	require.NoError(t, err)

	expiry := time.Now().Add(time.Hour).Truncate(time.Second)
	tok := &oauth2.Token{
		AccessToken:  "access-abc",
		TokenType:    "bearer",
		RefreshToken: "refresh-xyz",
		Expiry:       expiry,
	}
	err = disk.Save("example.gamefabric.dev", tok)
	require.NoError(t, err)

	got, err := disk.Load("example.gamefabric.dev")
	require.NoError(t, err)

	assert.Equal(t, tok.AccessToken, got.AccessToken)
	assert.Equal(t, tok.TokenType, got.TokenType)
	assert.Equal(t, tok.RefreshToken, got.RefreshToken)
	assert.True(t, tok.Expiry.Equal(got.Expiry), "expiry mismatch: want %v got %v", tok.Expiry, got.Expiry)
}

func TestDiskCache_LoadHandlesMissingFile(t *testing.T) {
	t.Setenv("GAMEFABRIC_CACHE_DIR", t.TempDir())

	disk, err := cache.NewDiskCache()
	require.NoError(t, err)

	_, err = disk.Load("no-such-host.gamefabric.dev")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no cached token")
}

func TestDiskCache_LoadHandlesCorruptFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GAMEFABRIC_CACHE_DIR", dir)

	// Write garbage where the credentials file would live.
	credPath := filepath.Join(dir, "gamefabric", "corrupt.gamefabric.dev.json")
	disk, err := cache.NewDiskCache()
	require.NoError(t, err)

	err = disk.Save("corrupt.gamefabric.dev", &oauth2.Token{AccessToken: "x"})
	require.NoError(t, err)

	// Overwrite the file with invalid JSON.
	require.NoError(t, os.WriteFile(credPath, []byte("not-json"), 0o600))

	_, err = disk.Load("corrupt.gamefabric.dev")
	require.Error(t, err)
}

func TestDiskCache_SaveDoesHostSanitization(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GAMEFABRIC_CACHE_DIR", dir)

	disk, err := cache.NewDiskCache()
	require.NoError(t, err)

	// A host with characters that are illegal in directory names on some OSes.
	host := "host:8080/path"
	err = disk.Save(host, &oauth2.Token{AccessToken: "tok"})
	require.NoError(t, err)

	// Must load back without error — path was sanitized correctly.
	got, err := disk.Load(host)
	require.NoError(t, err)
	assert.Equal(t, "tok", got.AccessToken)

	// Verify no literal ':' or '/' appear in the cache subdirectory name.
	matches, err := filepath.Glob(filepath.Join(dir, "gamefabric", "*.json"))
	require.NoError(t, err)
	require.Len(t, matches, 1)

	name := filepath.Base(matches[0])
	assert.NotContains(t, name, ":")
	assert.NotContains(t, name, "/")
}
