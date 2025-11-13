package provider_test

import (
	"context"
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

	require.Len(t, resp.Diagnostics, 3)
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
