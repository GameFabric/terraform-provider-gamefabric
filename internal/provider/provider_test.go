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

	// Only host is missing; absent credentials auto-select the device flow.
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

// TestProvider_ConfigureValidatesPartialCredentials ensures that providing only
// one of service_account / password produces a clear error instead of silently
// falling through to the device flow.
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

// TestProvider_DeviceFlow verifies the full Device Authorization Flow path.
// Credentials are intentionally omitted — the provider auto-detects this and
// falls back to RFC 8628 device authorization. The mock server returns
// authorization_pending twice before issuing a token.
func TestProvider_DeviceFlow(t *testing.T) {
	var tokenCalls atomic.Int64
	srv := testDeviceAuthServer(t, &tokenCalls)

	// Use a temp dir for the token cache so tests are hermetic.
	testEnv(t, "GAMEFABRIC_CACHE_DIR", t.TempDir())

	// Inject the test server's TLS client so golang.org/x/oauth2 reaches it.
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
				"service_account": tftypes.NewValue(tftypes.String, nil), // omitted → device flow
				"password":        tftypes.NewValue(tftypes.String, nil), // omitted → device flow
			}),
			Schema: schemaResp.Schema,
		},
		ClientCapabilities: tfprovider.ConfigureProviderClientCapabilities{},
	}, resp)

	require.Len(t, resp.Diagnostics, 0, "unexpected diagnostics: %v", resp.Diagnostics)

	// The mock token endpoint must have been polled at least once.
	assert.True(t, tokenCalls.Load() > 0, "token endpoint was never called")

	assert.NotPanics(t, func() {
		assert.NotNil(t, resp.DataSourceData.(*provcontext.Context).ClientSet)
		assert.NotNil(t, resp.ResourceData.(*provcontext.Context).ClientSet)
	})
}

// TestProvider_DeviceFlowViaCacheOnSecondRun verifies that a second Configure
// call reuses the cached token and does NOT go through the device flow again.
func TestProvider_DeviceFlowViaCacheOnSecondRun(t *testing.T) {
	var tokenCalls atomic.Int64
	srv := testDeviceAuthServer(t, &tokenCalls)

	cacheDir := t.TempDir()
	testEnv(t, "GAMEFABRIC_CACHE_DIR", cacheDir)

	ctx := context.WithValue(t.Context(), oauth2.HTTPClient, srv.Client())
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

	// First run: goes through full device flow.
	resp1 := configureProv()
	require.Len(t, resp1.Diagnostics, 0)
	callsAfterFirst := tokenCalls.Load()
	assert.True(t, callsAfterFirst > 0, "token endpoint should have been called on first run")

	// Second run: must reuse cached token, so token endpoint call count stays the same.
	resp2 := configureProv()
	require.Len(t, resp2.Diagnostics, 0)
	assert.Equal(t, callsAfterFirst, tokenCalls.Load(), "token endpoint should NOT be called again on cache hit")
}

// testDeviceAuthServer returns a TLS test server that simulates Dex's device
// authorization flow:
//   - POST /auth/device/code → device code response
//   - POST /auth/token → returns authorization_pending twice, then a token
func testDeviceAuthServer(t *testing.T, tokenCalls *atomic.Int64) *httptest.Server {
	t.Helper()

	var pendingCalls atomic.Int64

	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/auth/device/code":
			// RFC 8628 §3.2 — device authorization response.
			// Use interval=1 so the test polls every second instead of the
			// default 5-second interval, keeping test runtime short.
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"device_code":      "test-device-code-abc123",
				"user_code":        "WXYZ-1234",
				"verification_uri": r.Host + "/auth/device",
				"interval":         1,
				"expires_in":       300,
			})

		case "/auth/token":
			tokenCalls.Add(1)
			if pendingCalls.Add(1) <= 2 {
				// Simulate authorization_pending for the first two polls.
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "authorization_pending",
				})
				return
			}
			// Authorization complete — return a token.
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "short-lived-access-token",
				"token_type":    "bearer",
				"refresh_token": "refresh-token-xyz",
				"expires_in":    3600,
			})

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

func testEnv(t *testing.T, key, value string) {
	t.Helper()

	err := os.Setenv(key, value)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = os.Unsetenv(key)
		require.NoError(t, err)
	})
}
