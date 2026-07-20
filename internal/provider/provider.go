package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
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
	"golang.org/x/oauth2"
)

var _ provider.Provider = &Provider{}

const (
	envHost           = "GAMEFABRIC_HOST"
	envCustomerID     = "GAMEFABRIC_CUSTOMER_ID"
	envServiceAccount = "GAMEFABRIC_SERVICE_ACCOUNT"
	envPassword       = "GAMEFABRIC_PASSWORD"

	// deviceFlowClientID is the OAuth2 client ID for the device authorization
	// flow. This must be registered as a public client in Dex.
	deviceFlowClientID = "terraform"

	deviceFlowTimeout = 1 * time.Minute
)

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
					"the provider automatically uses the OAuth2 Device Authorization Flow (RFC 8628) for interactive login.",
				MarkdownDescription: "The service account username. When both `service_account` and `password` are omitted, " +
					"the provider automatically uses the OAuth2 Device Authorization Flow (RFC 8628) for interactive login.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				Description: "The service account password. When both service_account and password are omitted, " +
					"the provider automatically uses the OAuth2 Device Authorization Flow (RFC 8628) for interactive login.",
				MarkdownDescription: "The service account password. When both `service_account` and `password` are omitted, " +
					"the provider automatically uses the OAuth2 Device Authorization Flow (RFC 8628) for interactive login.",
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
	// If neither credential is set, use interactive Device Authorization Flow.
	if cfg.ServiceAccount.ValueString() == "" && cfg.Password.ValueString() == "" {
		return newClientSetDeviceFlow(ctx, cfg.Host.ValueString())
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

// newClientSetDeviceFlow authenticates via the OAuth2 Device Authorization
// Flow (RFC 8628). It first checks the local token cache; if a valid (or
// refreshable) token is found it is reused without prompting the user.
// Otherwise it initiates a new device flow by printing a verification URI and
// user code to stderr, then polls until the user completes the browser login.
func newClientSetDeviceFlow(ctx context.Context, host string) (clientset.Interface, error) {
	apiURL, err := url.Parse("https://" + host)
	if err != nil {
		return nil, fmt.Errorf("could not parse API URL %q: %w", host, err)
	}

	cfg := oauth2.Config{
		ClientID: deviceFlowClientID,
		Scopes:   []string{"openid", "email", "profile", "offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthStyle:     oauth2.AuthStyleInParams,
			TokenURL:      apiURL.JoinPath("/auth/token").String(),
			DeviceAuthURL: apiURL.JoinPath("/auth/device/code").String(),
		},
	}

	// Attempt to reuse a cached token (including silent refresh via refresh_token).
	if cached, err := deviceauth.Load(host); err == nil {
		ts := cfg.TokenSource(ctx, cached)
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
		// Cached token is unusable — fall through to a new device flow.
	}

	// Initiate a new device authorization request.
	da, err := cfg.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not initiate device authorization with %q: %w", host, err)
	}

	printDeviceAuthPrompt(da.VerificationURI, da.UserCode)

	// Poll until the user completes the browser login or the code expires.
	// Also cancel immediately on SIGINT/SIGTERM so Ctrl-C doesn't leave
	// Terraform hanging for the full deviceFlowTimeout duration.
	pollCtx, cancel := context.WithTimeout(ctx, deviceFlowTimeout)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			cancel()
		case <-pollCtx.Done():
		}
	}()
	defer signal.Stop(sigCh)

	tok, err := cfg.DeviceAccessToken(pollCtx, da)
	if err != nil {
		return nil, fmt.Errorf("device authorization did not complete: %w", err)
	}

	// Persist the token so the next run can skip the browser step.
	_ = deviceauth.Save(host, tok)

	restCfg := rest.Config{
		BaseURL:           apiURL.String(),
		Timeout:           10 * time.Second,
		BearerTokenSource: cfg.TokenSource(ctx, tok),
	}

	cs, err := clientset.New(restCfg)
	if err != nil {
		return nil, fmt.Errorf("could not create the gamefabric client: %w", err)
	}

	return cs, nil
}

// printDeviceAuthPrompt writes the device authorization instructions to the
// user's terminal. It opens /dev/tty directly so the message is always visible
// regardless of how Terraform captures the provider's stdout/stderr. Falls back
// to os.Stderr (visible in TF_LOG=DEBUG) if no terminal is available (CI).
func printDeviceAuthPrompt(verificationURI, userCode string) {
	msg := fmt.Sprintf(
		"\nTo authenticate with GameFabric, open the following URL in your browser:\n\n  %s\n\nAnd enter the code: %s\n\nWaiting for authorization...\n\n",
		verificationURI, userCode,
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
	// - Both absent → Device Authorization Flow (auto-detected, no config needed).
	// - Only one set → ambiguous intent, surface a clear error.
	hasServiceAccount := cfg.ServiceAccount.ValueString() != ""
	hasPassword := cfg.Password.ValueString() != ""
	if hasServiceAccount && !hasPassword {
		diags.Append(diag.NewErrorDiagnostic(
			"Missing Password",
			"service_account is set but password is not. Provide both to use password-based authentication, "+
				"or omit both to authenticate interactively via the Device Authorization Flow.",
		))
	}
	if hasPassword && !hasServiceAccount {
		diags.Append(diag.NewErrorDiagnostic(
			"Missing Service Account",
			"password is set but service_account is not. Provide both to use password-based authentication, "+
				"or omit both to authenticate interactively via the Device Authorization Flow.",
		))
	}
	return diags
}
