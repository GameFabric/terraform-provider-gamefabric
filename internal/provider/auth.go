package provider

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/cache"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

//nolint:lll
const successHTML = `<!DOCTYPE html><html style="height:100%"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Authentication successful</title></head><body style="margin:0;height:100%;display:flex;align-items:center;justify-content:center;background:#09090b;font-family:system-ui,sans-serif"><main style="max-width:420px;padding:42px;border:1px solid #3f3f46;border-radius:21px;background:rgba(27,27,27,.8);text-align:center;color:#a1a1aa"><div style="width:56px;height:56px;margin:0 auto 21px;display:flex;align-items:center;justify-content:center;border:1px solid #4ade80;border-radius:50%;background:rgba(74,222,128,.04);color:#4ade80;font-size:28px">✓</div><h1 style="margin:0 0 14px;font-size:24px;color:#fff">Authentication successful</h1><p style="margin:0;font-size:14px;line-height:1.5">You are now authenticated with GameFabric. You can close this window and return to Terraform.</p></main></body></html>`

const browserFlowTimeout = 5 * time.Minute

var oauthScopes = []string{"openid", "email", "profile", "offline_access"}

// BrowserFunc is a function that opens a URL in the user's default browser. It can be overridden for testing.
var BrowserFunc = browser.OpenURL

func authWithPasswordGrant(ctx context.Context, apiURL *url.URL, user, pass string) (oauth2.TokenSource, error) {
	oauth := oauth2.Config{
		ClientID: "api",
		Scopes:   oauthScopes,
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInHeader,
			TokenURL:  apiURL.JoinPath("/auth/token").String(),
		},
	}
	tok, err := oauth.PasswordCredentialsToken(ctx, user, pass)
	if err != nil {
		return nil, fmt.Errorf("could not request oauth token from %q: %w", apiURL.Host, err)
	}

	return oauth.TokenSource(ctx, tok), nil
}

//nolint:cyclop // Splitting this function does not improve readability.
func authWithBrowserFlow(ctx context.Context, apiURL *url.URL) (oauth2.TokenSource, error) {
	tokCache, err := cache.NewDiskCache()
	if err != nil {
		return nil, fmt.Errorf("could not create token cache: %w", err)
	}

	oauthCfg := oauth2.Config{
		ClientID: "terraform",
		Scopes:   oauthScopes,
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInParams,
			AuthURL:   apiURL.JoinPath("/auth/auth").String(),
			TokenURL:  apiURL.JoinPath("/auth/token").String(),
		},
	}

	if cached, err := tokCache.Load(apiURL.Host); err == nil {
		// This worked. Check the cached token and refresh it if necessary. If the refresh fails, we'll fall back to the browser flow.
		ts := oauthCfg.TokenSource(ctx, cached)
		if tok, err := ts.Token(); err == nil {
			// Cache the potentially-refreshed token.
			if err = tokCache.Save(apiURL.Host, tok); err != nil {
				tflog.Warn(ctx, "Could not update token cache", map[string]any{
					"error": err,
				})
			}
			return ts, nil
		}
	}

	l, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("could not start local callback server: %w", err)
	}
	defer func() { _ = l.Close() }()

	verifier := oauth2.GenerateVerifier()
	state := generateState()
	oauthCfg.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/", l.Addr().(*net.TCPAddr).Port)
	authURL := oauthCfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))

	tflog.Info(ctx, "Opening browser for GameFabric authentication", map[string]any{
		"auth_url": authURL,
	})
	if err = BrowserFunc(authURL); err != nil {
		tflog.Error(ctx, "Could not open browser automatically. Please open the URL above manually.", map[string]any{
			"auth_url": authURL,
			"error":    err,
		})
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	srv := &http.Server{Handler: newTokenHandler(state, codeCh, errCh)} //nolint:gosec // This is not a security issue; we are just starting a local HTTP server to receive the OAuth callback.
	go func() {
		if err := srv.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, time.Second)
		defer shutdownCancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	var code string
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("browser authentication cancelled: %w", ctx.Err())
	case err = <-errCh:
		return nil, fmt.Errorf("browser authentication failed: %w", err)
	case <-time.After(browserFlowTimeout):
		if ctx.Err() != nil {
			return nil, fmt.Errorf("browser authentication cancelled: %w", ctx.Err())
		}
		return nil, fmt.Errorf("browser authentication timed out after %v", browserFlowTimeout)
	case code = <-codeCh:
		// Fall through to exchange the code for a token.
	}

	tok, err := oauthCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("could not exchange authorization code: %w", err)
	}

	if err = tokCache.Save(apiURL.Host, tok); err != nil {
		tflog.Warn(ctx, "Could not cache token", map[string]any{
			"error": err,
		})
	}

	return oauthCfg.TokenSource(ctx, tok), nil
}

func newTokenHandler(expectedState string, codeCh chan<- string, errCh chan<- error) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		q := req.URL.Query()
		err := q.Get("error")
		state := q.Get("state")
		code := q.Get("code")

		switch {
		case state != expectedState:
			rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprint(rw, "invalid state parameter")
			errCh <- errors.New("state mismatch — possible CSRF")
			return
		case err != "":
			rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprint(rw, "Authentication failed. Check Terraform output for details.")
			errCh <- fmt.Errorf("auth error %q: %s", err, q.Get("error_description"))
			return
		case code == "":
			rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			rw.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprint(rw, "missing code parameter")
			errCh <- errors.New("missing code parameter")
			return
		}

		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(rw, successHTML)
		codeCh <- q.Get("code")
	})
}

func generateState() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// This should never happen, but if it does, panic so we don't return a weak state.
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
