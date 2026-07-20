package provider_test

import (
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gamefabric/gf-core/pkg/apiclient/fake"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestProvider_Meta(t *testing.T) {
	resp := &tfprovider.MetadataResponse{}

	prov := provider.New("1.0.0")()
	prov.Metadata(t.Context(), tfprovider.MetadataRequest{}, resp)

	assert.Equal(t, resp.Version, "1.0.0")
	assert.Equal(t, resp.TypeName, "gamefabric")
}

func TestProvider_Schema(t *testing.T) {
	resp := &tfprovider.SchemaResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, resp)

	require.Len(t, resp.Diagnostics, 0)
	require.Len(t, resp.Schema.Attributes, 4)

	assert.Contains(t, slices.Collect(maps.Keys(resp.Schema.Attributes)), "host")
	assert.Contains(t, slices.Collect(maps.Keys(resp.Schema.Attributes)), "customer_id")
	assert.Contains(t, slices.Collect(maps.Keys(resp.Schema.Attributes)), "service_account")
	assert.Contains(t, slices.Collect(maps.Keys(resp.Schema.Attributes)), "password")
}

func TestProvider_ConfigureWhenClientSetIsSet(t *testing.T) {
	cs, err := fake.New()
	require.NoError(t, err)

	prov := provider.NewWithClientSet("1.0.0", cs)()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, &tfprovider.SchemaResponse{})

	resp := &tfprovider.ConfigureResponse{}
	assert.NotPanics(t, func() {
		prov.Configure(t.Context(), tfprovider.ConfigureRequest{}, resp)

		assert.NotNil(t, resp.DataSourceData.(*provcontext.Context).ClientSet)
		assert.NotNil(t, resp.ResourceData.(*provcontext.Context).ClientSet)
	})
}

func TestProvider_ConfigureValidates(t *testing.T) {
	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	// Avoid interference from environment variables.
	for _, env := range []string{"GAMEFABRIC_HOST", "GAMEFABRIC_CUSTOMER_ID", "GAMEFABRIC_SERVICE_ACCOUNT", "GAMEFABRIC_PASSWORD"} {
		err := os.Unsetenv(env)
		require.NoError(t, err)
	}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(t.Context(), tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, ""),
				"customer_id":     tftypes.NewValue(tftypes.String, ""),
				"service_account": tftypes.NewValue(tftypes.String, ""),
				"password":        tftypes.NewValue(tftypes.String, ""),
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 1)
}

func TestProvider_ConfigureValidatesHostAndCustomerID(t *testing.T) {
	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(t.Context(), tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, "other"),
				"customer_id":     tftypes.NewValue(tftypes.String, "something"),
				"service_account": tftypes.NewValue(tftypes.String, "test"),
				"password":        tftypes.NewValue(tftypes.String, "test"),
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 1)
}

func TestProvider_Configure(t *testing.T) {
	var called atomic.Int64
	srv := testOAuthServer(t, &called, "service_account", "secr3t")

	// Inject custom HTTP client for the oauth2 call.
	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, srv.Client())

	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(ctx, tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, strings.TrimPrefix(srv.URL, "https://")),
				"customer_id":     tftypes.NewValue(tftypes.String, nil),
				"service_account": tftypes.NewValue(tftypes.String, "service_account"),
				"password":        tftypes.NewValue(tftypes.String, "secr3t"),
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 0)
	assert.True(t, called.Load() > 0)

	assert.NotPanics(t, func() {
		assert.NotNil(t, resp.DataSourceData.(*provcontext.Context).ClientSet)
		assert.NotNil(t, resp.ResourceData.(*provcontext.Context).ClientSet)
	})
}

func TestProvider_ConfigureUsesCustomerID(t *testing.T) {
	var called atomic.Int64
	srv := testOAuthServer(t, &called, "service_account", "secr3t")

	// Inject custom HTTP client for the oauth2 call.
	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, srv.Client())

	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(ctx, tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, strings.TrimPrefix(srv.URL, "https://")),
				"customer_id":     tftypes.NewValue(tftypes.String, nil),
				"service_account": tftypes.NewValue(tftypes.String, "service_account"),
				"password":        tftypes.NewValue(tftypes.String, "secr3t"),
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 0)
	assert.True(t, called.Load() > 0)

	assert.NotPanics(t, func() {
		assert.NotNil(t, resp.DataSourceData.(*provcontext.Context).ClientSet)
		assert.NotNil(t, resp.ResourceData.(*provcontext.Context).ClientSet)
	})
}

func TestProvider_ConfigureUsesEnvs(t *testing.T) {
	var called atomic.Int64
	srv := testOAuthServer(t, &called, "service_account", "secr3t")

	// Inject custom HTTP client for the oauth2 call.
	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, srv.Client())

	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	// Set environment variables.
	testEnv(t, "GAMEFABRIC_HOST", strings.TrimPrefix(srv.URL, "https://"))
	testEnv(t, "GAMEFABRIC_SERVICE_ACCOUNT", "service_account")
	testEnv(t, "GAMEFABRIC_PASSWORD", "secr3t")

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(ctx, tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, ""),
				"customer_id":     tftypes.NewValue(tftypes.String, ""),
				"service_account": tftypes.NewValue(tftypes.String, ""),
				"password":        tftypes.NewValue(tftypes.String, ""),
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 0)
	assert.True(t, called.Load() > 0)

	assert.NotPanics(t, func() {
		assert.NotNil(t, resp.DataSourceData.(*provcontext.Context).ClientSet)
		assert.NotNil(t, resp.ResourceData.(*provcontext.Context).ClientSet)
	})
}

func TestProvider_ConfigureValidatesPartialCredentials(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		serviceAccount string
		password       string
		wantErrSummary string
	}{
		{
			name:           "only service_account",
			serviceAccount: "sa",
			password:       "",
			wantErrSummary: "Missing Password",
		},
		{
			name:           "only password",
			serviceAccount: "",
			password:       "pass",
			wantErrSummary: "Missing Service Account",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			schemaResp := &tfprovider.SchemaResponse{}
			resp := &tfprovider.ConfigureResponse{}

			prov := provider.New("1.0.0")()
			prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
			prov.Configure(t.Context(), tfprovider.ConfigureRequest{
				Config: tfsdk.Config{
					Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
						"host":            tftypes.NewValue(tftypes.String, "example.gamefabric.dev"),
						"customer_id":     tftypes.NewValue(tftypes.String, nil),
						"service_account": tftypes.NewValue(tftypes.String, tc.serviceAccount),
						"password":        tftypes.NewValue(tftypes.String, tc.password),
					}),
					Schema: schemaResp.Schema,
				},
				ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
			}, resp)

			require.Len(t, resp.Diagnostics, 1)
			assert.Equal(t, tc.wantErrSummary, resp.Diagnostics[0].Summary())
		})
	}
}

func TestProvider_BrowserFlow(t *testing.T) {
	var tokenCalls atomic.Int64
	srv := testBrowserAuthServer(t, &tokenCalls)

	// Use a temp dir for the token cache so tests are hermetic.
	testEnv(t, "GAMEFABRIC_CACHE_DIR", t.TempDir())

	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, srv.Client())
	provider.BrowserFunc = testBrowserFunc(srv)

	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(ctx, tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, strings.TrimPrefix(srv.URL, "https://")),
				"customer_id":     tftypes.NewValue(tftypes.String, nil),
				"service_account": tftypes.NewValue(tftypes.String, nil), // omitted → browser flow
				"password":        tftypes.NewValue(tftypes.String, nil), // omitted → browser flow
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 0, "unexpected diagnostics: %v", resp.Diagnostics)

	// The mock token endpoint must have been called for the code exchange.
	assert.True(t, tokenCalls.Load() > 0, "token endpoint was never called")

	assert.NotPanics(t, func() {
		assert.NotNil(t, resp.DataSourceData.(*provcontext.Context).ClientSet)
		assert.NotNil(t, resp.ResourceData.(*provcontext.Context).ClientSet)
	})
}

func TestProvider_BrowserFlowViaCacheOnSecondRun(t *testing.T) {
	var tokenCalls atomic.Int64
	srv := testBrowserAuthServer(t, &tokenCalls)

	cacheDir := t.TempDir()
	testEnv(t, "GAMEFABRIC_CACHE_DIR", cacheDir)

	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, srv.Client())
	provider.BrowserFunc = testBrowserFunc(srv)
	host := strings.TrimPrefix(srv.URL, "https://")

	schemaResp := &tfprovider.SchemaResponse{}

	configureProv := func() *tfprovider.ConfigureResponse {
		resp := &tfprovider.ConfigureResponse{}
		prov := provider.New("1.0.0")()
		prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
		prov.Configure(ctx, tfprovider.ConfigureRequest{
			Config: tfsdk.Config{
				Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
					"host":            tftypes.NewValue(tftypes.String, host),
					"customer_id":     tftypes.NewValue(tftypes.String, nil),
					"service_account": tftypes.NewValue(tftypes.String, nil),
					"password":        tftypes.NewValue(tftypes.String, nil),
				}),
				Schema: schemaResp.Schema,
			},
			ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
		}, resp)
		return resp
	}

	// First run: goes through full browser flow.
	resp1 := configureProv()
	require.Len(t, resp1.Diagnostics, 0)
	callsAfterFirst := tokenCalls.Load()
	assert.True(t, callsAfterFirst > 0, "token endpoint should have been called on first run")

	// Second run: must reuse cached token, so token endpoint call count stays the same.
	resp2 := configureProv()
	require.Len(t, resp2.Diagnostics, 0)
	assert.Equal(t, callsAfterFirst, tokenCalls.Load(), "token endpoint should NOT be called again on cache hit")
}

func TestProvider_BrowserFlowHandlesStateMismatch(t *testing.T) {
	srv := testBrowserAuthServerWithCallbackOverride(t, func(w http.ResponseWriter, r *http.Request) {
		// Redirect to the callback with the correct code but the WRONG state.
		q := r.URL.Query()
		callbackURL := q.Get("redirect_uri") + "?code=test-auth-code&state=WRONG_STATE"
		http.Redirect(w, r, callbackURL, http.StatusFound)
	})

	testEnv(t, "GAMEFABRIC_CACHE_DIR", t.TempDir())

	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, srv.Client())
	provider.BrowserFunc = testBrowserFunc(srv)

	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(ctx, tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, strings.TrimPrefix(srv.URL, "https://")),
				"customer_id":     tftypes.NewValue(tftypes.String, nil),
				"service_account": tftypes.NewValue(tftypes.String, nil),
				"password":        tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 1)
	assert.Contains(t, resp.Diagnostics[0].Detail(), "state mismatch")
}

func TestProvider_BrowserFlowHandlesAuthError(t *testing.T) {
	srv := testBrowserAuthServerWithCallbackOverride(t, func(w http.ResponseWriter, r *http.Request) {
		// Redirect to the callback with an error instead of a code.
		q := r.URL.Query()
		callbackURL := q.Get("redirect_uri") + "?error=access_denied&error_description=User+denied+access&state=" + q.Get("state")
		http.Redirect(w, r, callbackURL, http.StatusFound)
	})

	testEnv(t, "GAMEFABRIC_CACHE_DIR", t.TempDir())

	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, srv.Client())
	provider.BrowserFunc = testBrowserFunc(srv)

	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(ctx, tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, strings.TrimPrefix(srv.URL, "https://")),
				"customer_id":     tftypes.NewValue(tftypes.String, nil),
				"service_account": tftypes.NewValue(tftypes.String, nil),
				"password":        tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 1)
	assert.Contains(t, resp.Diagnostics[0].Detail(), "access_denied")
}

func TestProvider_BrowserFlowHandlesTimeout(t *testing.T) {
	var tokenCalls atomic.Int64
	srv := testBrowserAuthServer(t, &tokenCalls)

	testEnv(t, "GAMEFABRIC_CACHE_DIR", t.TempDir())

	// The browser opener does nothing — the callback URL is never visited.
	provider.BrowserFunc = func(string) error { return nil }

	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, srv.Client())
	// Short deadline so the test does not wait 5 minutes.
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(ctx, tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, strings.TrimPrefix(srv.URL, "https://")),
				"customer_id":     tftypes.NewValue(tftypes.String, nil),
				"service_account": tftypes.NewValue(tftypes.String, nil),
				"password":        tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 1)
	// The test uses a short parent-context deadline, so the flow is cancelled
	// rather than timing out via browserFlowTimeout. Either message is valid.
	detail := strings.ToLower(resp.Diagnostics[0].Detail())
	assert.True(t, strings.Contains(detail, "timed out") || strings.Contains(detail, "cancelled"), "unexpected detail: %s", detail)
}

func TestProvider_EnforceAuthMethodBrowser(t *testing.T) {
	var tokenCalls atomic.Int64
	srv := testBrowserAuthServer(t, &tokenCalls)

	testEnv(t, "GAMEFABRIC_CACHE_DIR", t.TempDir())
	testEnv(t, "GAMEFABRIC_ENFORCE_AUTH_METHOD", "browser")

	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, srv.Client())
	provider.BrowserFunc = testBrowserFunc(srv)

	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(ctx, tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, strings.TrimPrefix(srv.URL, "https://")),
				"customer_id":     tftypes.NewValue(tftypes.String, nil),
				"service_account": tftypes.NewValue(tftypes.String, "sa"),     // set, but should be ignored
				"password":        tftypes.NewValue(tftypes.String, "secret"), // set, but should be ignored
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 0, "unexpected diagnostics: %v", resp.Diagnostics)
	assert.True(t, tokenCalls.Load() > 0, "browser flow token endpoint should have been called")
}

func TestProvider_EnforceAuthMethodServiceAccount(t *testing.T) {
	var called atomic.Int64
	srv := testOAuthServer(t, &called, "sa", "secret")

	testEnv(t, "GAMEFABRIC_ENFORCE_AUTH_METHOD", "serviceaccount")

	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, srv.Client())

	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(ctx, tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, strings.TrimPrefix(srv.URL, "https://")),
				"customer_id":     tftypes.NewValue(tftypes.String, nil),
				"service_account": tftypes.NewValue(tftypes.String, "sa"),
				"password":        tftypes.NewValue(tftypes.String, "secret"),
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 0, "unexpected diagnostics: %v", resp.Diagnostics)
	assert.True(t, called.Load() > 0, "password grant endpoint should have been called")
}

func TestProvider_EnforceAuthMethodInvalid(t *testing.T) {
	testEnv(t, "GAMEFABRIC_ENFORCE_AUTH_METHOD", "badvalue")

	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(t.Context(), tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, "example.gamefabric.dev"),
				"customer_id":     tftypes.NewValue(tftypes.String, nil),
				"service_account": tftypes.NewValue(tftypes.String, nil),
				"password":        tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 1)
	assert.Equal(t, "Invalid Auth Method", resp.Diagnostics[0].Summary())
}

func TestProvider_EnforceAuthMethodServiceAccountHandlesMissingCredentials(t *testing.T) {
	testEnv(t, "GAMEFABRIC_ENFORCE_AUTH_METHOD", "serviceaccount")

	schemaResp := &tfprovider.SchemaResponse{}
	resp := &tfprovider.ConfigureResponse{}

	prov := provider.New("1.0.0")()
	prov.Schema(t.Context(), tfprovider.SchemaRequest{}, schemaResp)
	prov.Configure(t.Context(), tfprovider.ConfigureRequest{
		Config: tfsdk.Config{
			Raw: tftypes.NewValue(tftypes.Object{}, map[string]tftypes.Value{
				"host":            tftypes.NewValue(tftypes.String, "example.gamefabric.dev"),
				"customer_id":     tftypes.NewValue(tftypes.String, nil),
				"service_account": tftypes.NewValue(tftypes.String, nil),
				"password":        tftypes.NewValue(tftypes.String, nil),
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 1)
	assert.Equal(t, "Missing Credentials", resp.Diagnostics[0].Summary())
}

func testBrowserAuthServer(t *testing.T, tokenCalls *atomic.Int64) *httptest.Server {
	t.Helper()

	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/auth":
			// Immediately redirect to the redirect_uri with an authorization code,
			// simulating a user who logged in instantly.
			q := r.URL.Query()
			callbackURL := q.Get("redirect_uri") + "?code=test-auth-code&state=" + q.Get("state")
			http.Redirect(w, r, callbackURL, http.StatusFound)

		case "/auth/token":
			tokenCalls.Add(1)
			_ = r.ParseForm()
			if r.FormValue("grant_type") != "authorization_code" || r.FormValue("code") != "test-auth-code" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "test-access-token",
				"token_type":    "bearer",
				"refresh_token": "test-refresh-token",
				"expires_in":    3600,
			})

		default:
			http.NotFound(w, r)
		}
	}))
}

func testBrowserAuthServerWithCallbackOverride(t *testing.T, authHandler http.HandlerFunc) *httptest.Server {
	t.Helper()

	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/auth":
			authHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
}

func testOAuthServer(t *testing.T, called *atomic.Int64, user, pass string) *httptest.Server {
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/auth/token", r.URL.Path)

		err := r.ParseForm()

		require.NoError(t, err)
		assert.Equal(t, user, r.FormValue("username"))
		assert.Equal(t, pass, r.FormValue("password"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"token"}`))

		called.Add(1)
	}))
}

func testBrowserFunc(srv *httptest.Server) func(string) error {
	return func(authURL string) error {
		go func() {
			resp, err := srv.Client().Get(authURL) //nolint:noctx // test helper, context not needed
			if err == nil {
				_ = resp.Body.Close()
			}
		}()
		return nil
	}
}

func testEnv(t *testing.T, key, value string) {
	t.Helper()

	err := os.Setenv(key, value)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = os.Unsetenv(key)
		require.NoError(t, err)
	})
}
