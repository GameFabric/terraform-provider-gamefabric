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
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	dsauthentication "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/authentication"
	dscontainer "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/container"
	dscore "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/core"
	dsnotification "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/notification"
	dsprotection "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/protection"
	dsprovisioning "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/provisioning"
	dsrbac "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/rbac"
	dsstorage "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/storage"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/deviceauth"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/armada"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/audit"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/authentication"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/billing"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/formation"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/notification"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/protection"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/rbac"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/storage"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

var _ provider.Provider = &Provider{}

// BrowserOpenerKey is a context key for injecting a custom browser opener.
// Used in tests to simulate browser interaction without actually opening a browser.
type BrowserOpenerKey struct{}

const (
	envHost           = "GAMEFABRIC_HOST"
	envCustomerID     = "GAMEFABRIC_CUSTOMER_ID"
	envServiceAccount = "GAMEFABRIC_SERVICE_ACCOUNT"
	envPassword       = "GAMEFABRIC_PASSWORD"

	// browserFlowClientID is the OAuth2 client ID for the browser-based
	// authorization code flow. This must be registered as a public client in Dex.
	browserFlowClientID = "terraform"

	browserFlowTimeout = 5 * time.Minute
)

// successHTML is served to the browser after a successful authorization callback.
const successHTML = `<!DOCTYPE html><html><body>
<h1>Authentication successful</h1>
<p>You are now authenticated with GameFabric. You can close this window and return to Terraform.</p>
</body></html>`

// providerModel is the provider configuration model.
type providerModel struct {
	Host           types.String `tfsdk:"host"`
	CustomerID     types.String `tfsdk:"customer_id"`
	ServiceAccount types.String `tfsdk:"service_account"`
	Password       types.String `tfsdk:"password"`
}

// Provider is the GameFabric provider implementation.
type Provider struct {
	version   string
	clientSet clientset.Interface
}

// New returns a new provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{
			version: version,
		}
	}
}

// NewWithClientSet returns a new provider with the given client set.
func NewWithClientSet(version string, clientSet clientset.Interface) func() provider.Provider {
	return func() provider.Provider {
		return &Provider{
			version:   version,
			clientSet: clientSet,
		}
	}
}

// Metadata defines the provider type name and version.
func (p *Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "gamefabric"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "GameFabric Terraform Provider",
		MarkdownDescription: "GameFabric provider is used to interact, provision and manage your GameFabric resources.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description:         "The GameFabric API host for example: 'example.gamefabric.dev'.",
				MarkdownDescription: "The GameFabric API host for example: `example.gamefabric.dev`.",
				Optional:            true,
			},
			"customer_id": schema.StringAttribute{
				Description:         "The customer ID (first segment of your installation URL). If your installation URL is 'customerID.gamefabric.dev', set this to 'customerID'.",
				MarkdownDescription: "The customer ID (first segment of your installation URL). If your installation URL is `customerID.gamefabric.dev`, set this to `customerID`.",
				Optional:            true,
			},
			"service_account": schema.StringAttribute{
				Description: "The service account username. When both service_account and password are omitted, " +
					"the provider automatically opens your browser for interactive SSO login (OAuth2 Authorization Code Flow).",
				MarkdownDescription: "The service account username. When both `service_account` and `password` are omitted, " +
					"the provider automatically opens your browser for interactive SSO login (OAuth2 Authorization Code Flow).",
				Optional: true,
			},
			"password": schema.StringAttribute{
				Description: "The service account password. When both service_account and password are omitted, " +
					"the provider automatically opens your browser for interactive SSO login (OAuth2 Authorization Code Flow).",
				MarkdownDescription: "The service account password. When both `service_account` and `password` are omitted, " +
					"the provider automatically opens your browser for interactive SSO login (OAuth2 Authorization Code Flow).",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares the provider for data sources and resources.
func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	if p.clientSet != nil {
		// Use the pre-configured client set (for testing).
		provCtx := provcontext.NewContext(p.clientSet)
		resp.DataSourceData = provCtx
		resp.ResourceData = provCtx
		return
	}

	cfg := &providerModel{}
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	applyEnvVars(cfg)

	resp.Diagnostics.Append(validate(cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	p.clientSet, err = buildClientSetFromConfig(ctx, cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create GameFabric client",
			err.Error(),
		)
		return
	}

	provCtx := provcontext.NewContext(p.clientSet)
	resp.DataSourceData = provCtx
	resp.ResourceData = provCtx
}

// buildClientSetFromConfig auto-detects the authentication method from the
// provider configuration and creates a client set accordingly.
func buildClientSetFromConfig(ctx context.Context, cfg *providerModel) (clientset.Interface, error) {
	// If neither credential is set, use interactive browser-based authorization.
	if cfg.ServiceAccount.ValueString() == "" && cfg.Password.ValueString() == "" {
		return newClientSetBrowserFlow(ctx, cfg.Host.ValueString())
	}

	// Otherwise both are set (validated earlier), so use password grant.
	return newClientSet(ctx, cfg.Host.ValueString(), cfg.ServiceAccount.ValueString(), cfg.Password.ValueString())
}

// DataSources defines the data sources implemented in the provider.
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		dsauthentication.NewServiceAccount,
		dsauthentication.NewServiceAccounts,
		dscontainer.NewBranch,
		dscontainer.NewBranches,
		dscontainer.NewImage,
		dscore.NewConfigFile,
		dscore.NewConfigFiles,
		dscore.NewEnvironment,
		dscore.NewEnvironments,
		dscore.NewLocation,
		dscore.NewLocations,
		dscore.NewRegion,
		dscore.NewRegions,
		dscore.NewSecret,
		dscore.NewSecrets,
		dsnotification.NewReceiverDataSource,
		dsprotection.NewGatewayPolicies,
		dsprotection.NewGatewayPolicy,
		dsprotection.NewProtocol,
		dsprotection.NewProtocols,
		dsprovisioning.NewAllocator,
		dsprovisioning.NewPingDiscovery,
		dsrbac.NewGroup,
		dsrbac.NewGroups,
		dsstorage.NewVolume,
		dsstorage.NewVolumeStore,
		dsstorage.NewVolumeStores,
		dsstorage.NewVolumes,
	}
}

// Resources defines the resources implemented in the provider.
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		armada.NewArmada,
		armada.NewArmadaSet,
		audit.NewExportStore,
		authentication.NewProvider,
		authentication.NewServiceAccountResource,
		authentication.NewServiceAccountPasswordResource,
		billing.NewCloudBudgetResource,
		container.NewBranch,
		container.NewImageUpdater,
		core.NewConfigFile,
		core.NewEnvironment,
		core.NewRegion,
		core.NewSecret,
		formation.NewFormation,
		formation.NewVessel,
		notification.NewReceiverResource,
		protection.NewGatewayPolicy,
		rbac.NewGroup,
		rbac.NewRole,
		rbac.NewRoleBinding,
		storage.NewVolume,
		storage.NewVolumeStoreRetentionPolicy,
	}
}

func newClientSet(ctx context.Context, host, user, pass string) (clientset.Interface, error) {
	apiURL, err := url.Parse("https://" + host)
	if err != nil {
		return nil, fmt.Errorf("could not parse API URL %q: %w", host, err)
	}

	oauth := oauth2.Config{
		ClientID: "api",
		Scopes:   []string{"openid", "email", "profile", "offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInHeader,
			TokenURL:  apiURL.JoinPath("/auth/token").String(),
		},
	}
	tok, err := oauth.PasswordCredentialsToken(ctx, user, pass)
	if err != nil {
		return nil, fmt.Errorf("could not request oauth token from %q: %w", host, err)
	}

	restCfg := rest.Config{
		BaseURL:           apiURL.String(),
		Timeout:           10 * time.Second,
		BearerTokenSource: oauth.TokenSource(ctx, tok),
	}

	cs, err := clientset.New(restCfg)
	if err != nil {
		return nil, fmt.Errorf("could not create the gamefabric client: %w", err)
	}

	return cs, nil
}

// newClientSetBrowserFlow authenticates via the OAuth2 Authorization Code Flow
// with PKCE and a local loopback redirect (RFC 8252). It first checks the local
// token cache; if a valid (or refreshable) token is found it is reused without
// opening a browser. Otherwise it starts a local HTTP server, opens the browser
// to the authorization URL, and waits for the redirect callback to deliver the
// authorization code.
func newClientSetBrowserFlow(ctx context.Context, host string) (clientset.Interface, error) {
	apiURL, err := url.Parse("https://" + host)
	if err != nil {
		return nil, fmt.Errorf("could not parse API URL %q: %w", host, err)
	}

	oauthCfg := oauth2.Config{
		ClientID: browserFlowClientID,
		Scopes:   []string{"openid", "email", "profile", "offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInParams,
			AuthURL:   apiURL.JoinPath("/auth/auth").String(),
			TokenURL:  apiURL.JoinPath("/auth/token").String(),
		},
	}

	// Attempt to reuse a cached token (including silent refresh via refresh_token).
	if cached, err := deviceauth.Load(host); err == nil {
		ts := oauthCfg.TokenSource(ctx, cached)
		if tok, err := ts.Token(); err == nil {
			// Cache the potentially-refreshed token.
			_ = deviceauth.Save(host, tok)
			restCfg := rest.Config{
				BaseURL:           apiURL.String(),
				Timeout:           10 * time.Second,
				BearerTokenSource: ts,
			}
			cs, err := clientset.New(restCfg)
			if err != nil {
				return nil, fmt.Errorf("could not create the gamefabric client: %w", err)
			}
			return cs, nil
		}
		// Cached token is unusable — fall through to a new browser flow.
	}

	// Start a local callback server. Port 0 lets the OS pick a free port.
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("could not start local callback server: %w", err)
	}
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/", listener.Addr().(*net.TCPAddr).Port)
	oauthCfg.RedirectURL = redirectURI

	// Generate PKCE verifier and random state.
	verifier := oauth2.GenerateVerifier()
	state, err := generateState()
	if err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("could not generate PKCE state: %w", err)
	}

	// Build authorization URL, print it, and open the browser.
	authURL := oauthCfg.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier))
	printBrowserAuthPrompt(authURL)
	openInBrowser(ctx, authURL)

	// Serve the callback. Buffered channels so the handler never blocks even
	// if we have already timed out or been interrupted.
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	cbSrv := &http.Server{Handler: callbackHandler(state, codeCh, errCh)} //nolint:gosec // local server, no TLS needed on loopback
	go func() { _ = cbSrv.Serve(listener) }()
	defer func() { _ = cbSrv.Close() }()

	// Wait for the authorization code, a timeout, or a signal. Cancel on
	// SIGINT/SIGTERM so Ctrl-C doesn't leave Terraform hanging.
	waitCtx, cancel := context.WithTimeout(ctx, browserFlowTimeout)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	var code string
	select {
	case code = <-codeCh:
	case err = <-errCh:
		return nil, fmt.Errorf("browser authentication failed: %w", err)
	case <-sigCh:
		cancel()
		return nil, errors.New("browser authentication cancelled by user")
	case <-waitCtx.Done():
		return nil, fmt.Errorf("browser authentication timed out after %v", browserFlowTimeout)
	}

	// Exchange the authorization code for tokens.
	tok, err := oauthCfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return nil, fmt.Errorf("could not exchange authorization code: %w", err)
	}

	// Persist the token so the next run can skip the browser step.
	_ = deviceauth.Save(host, tok)

	restCfg := rest.Config{
		BaseURL:           apiURL.String(),
		Timeout:           10 * time.Second,
		BearerTokenSource: oauthCfg.TokenSource(ctx, tok),
	}

	cs, err := clientset.New(restCfg)
	if err != nil {
		return nil, fmt.Errorf("could not create the gamefabric client: %w", err)
	}

	return cs, nil
}

// callbackHandler returns an HTTP handler that captures the authorization code
// from Dex's redirect. It writes a success page to the browser and sends the
// code (or an error) to the provided channels. sync.Once ensures that only the
// first request is processed even if the browser sends multiple requests.
func callbackHandler(expectedState string, codeCh chan<- string, errCh chan<- error) http.Handler {
	var once sync.Once
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() {
			q := r.URL.Query()
			if q.Get("state") != expectedState {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprint(w, "invalid state parameter")
				errCh <- errors.New("state mismatch — possible CSRF")
				return
			}
			if e := q.Get("error"); e != "" {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprint(w, "Authentication failed. Check Terraform output for details.")
				errCh <- fmt.Errorf("auth error %q: %s", e, q.Get("error_description"))
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprint(w, successHTML)
			codeCh <- q.Get("code")
		})
	})
}

// generateState returns a cryptographically random base64url-encoded string
// used as the OAuth2 state parameter to prevent CSRF.
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// openInBrowser opens the given URL in the user's default browser. Tests can
// override this by injecting a BrowserOpenerKey value into the context.
func openInBrowser(ctx context.Context, authURL string) {
	if opener, ok := ctx.Value(BrowserOpenerKey{}).(func(string)); ok {
		opener(authURL)
		return
	}
	_ = browser.OpenURL(authURL)
}

// printBrowserAuthPrompt writes the authorization URL to the user's terminal.
// It opens /dev/tty directly so the message is always visible regardless of
// how Terraform captures the provider's stdout/stderr. Falls back to os.Stderr
// (visible in TF_LOG=DEBUG) if no terminal is available (CI/headless).
func printBrowserAuthPrompt(authURL string) {
	msg := fmt.Sprintf(
		"\nOpening your browser for GameFabric authentication:\n\n  %s\n\nWaiting for authorization...\n\n",
		authURL,
	)
	tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	if err == nil {
		_, _ = fmt.Fprint(tty, msg)
		_ = tty.Close()
		return
	}
	_, _ = fmt.Fprint(os.Stderr, msg)
}

// applyEnvVars fills any unset model fields from environment variables and
// derives the host from the customer ID when only the latter is provided.
func applyEnvVars(cfg *providerModel) {
	if env := os.Getenv(envHost); cfg.Host.ValueString() == "" && env != "" {
		cfg.Host = types.StringValue(env)
	}
	if env := os.Getenv(envCustomerID); cfg.CustomerID.ValueString() == "" && env != "" {
		cfg.CustomerID = types.StringValue(env)
	}
	if env := os.Getenv(envServiceAccount); cfg.ServiceAccount.ValueString() == "" && env != "" {
		cfg.ServiceAccount = types.StringValue(env)
	}
	if env := os.Getenv(envPassword); cfg.Password.ValueString() == "" && env != "" {
		cfg.Password = types.StringValue(env)
	}
	if cfg.CustomerID.ValueString() != "" && cfg.Host.ValueString() == "" {
		cfg.Host = types.StringValue(cfg.CustomerID.ValueString() + ".gamefabric.dev")
	}
}

func validate(cfg *providerModel) []diag.Diagnostic {
	diags := make(diag.Diagnostics, 0, 4)
	if cfg.Host.ValueString() == "" {
		diags.Append(diag.NewErrorDiagnostic(
			"Missing Host",
			"The provider cannot create the GameFabric client as there is no host configured. "+
				"Please set the host value in the provider configuration or use the "+envHost+" environment variable. "+
				"If you have a customer ID, the host will be derived from it.",
		))
	}
	if cfg.CustomerID.ValueString() != "" && cfg.Host.ValueString() != cfg.CustomerID.ValueString()+".gamefabric.dev" {
		diags.Append(diag.NewErrorDiagnostic(
			"Conflicting Host and Customer ID",
			"The provider cannot create the GameFabric client as both the host and customer ID are configured, "+
				"but the host is not derived from the customer ID. ",
		))
	}
	// Credentials must be fully present or fully absent.
	// - Both set → password grant.
	// - Both absent → browser-based authorization code flow (auto-detected).
	// - Only one set → ambiguous intent, surface a clear error.
	hasServiceAccount := cfg.ServiceAccount.ValueString() != ""
	hasPassword := cfg.Password.ValueString() != ""
	if hasServiceAccount && !hasPassword {
		diags.Append(diag.NewErrorDiagnostic(
			"Missing Password",
			"service_account is set but password is not. Provide both to use password-based authentication, "+
				"or omit both to authenticate interactively via the browser.",
		))
	}
	if hasPassword && !hasServiceAccount {
		diags.Append(diag.NewErrorDiagnostic(
			"Missing Service Account",
			"password is set but service_account is not. Provide both to use password-based authentication, "+
				"or omit both to authenticate interactively via the browser.",
		))
	}
	return diags
}
