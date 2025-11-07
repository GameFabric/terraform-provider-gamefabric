package provider

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/gamefabric/gf-apiclient/rest"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	dscontainer "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/container"
	dscore "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/core"
	dsprotection "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/protection"
	dsrbac "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/rbac"
	dsstorage "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/storage"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/armada"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/container"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/formation"
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
				Description:         "The GameFabric API host URL.",
				MarkdownDescription: "The GameFabric API host URL.",
				Optional:            true,
			},
			"customer_id": schema.StringAttribute{
				Description:         "The customer ID (first segment of your installation URL). If your installation URL is 'customerID.gamefabric.dev', set this to 'customerID'.",
				MarkdownDescription: "The customer ID (first segment of your installation URL). If your installation URL is `customerID.gamefabric.dev`, set this to `customerID`.",
				Optional:            true,
			},
			"service_account": schema.StringAttribute{
				Description:         "The service account username.",
				MarkdownDescription: "The service account username.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				Description:         "The service account password.",
				MarkdownDescription: "The service account password.",
				Optional:            true,
				Sensitive:           true,
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

	resp.Diagnostics.Append(validate(cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	p.clientSet, err = newClientSet(ctx, cfg.Host.ValueString(), cfg.ServiceAccount.ValueString(), cfg.Password.ValueString())
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

// DataSources defines the data sources implemented in the provider.
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		dscore.NewConfigFile,
		dscore.NewConfigFiles,
		dscore.NewEnvironment,
		dscore.NewEnvironments,
		dscore.NewLocation,
		dscore.NewLocations,
		dscore.NewRegion,
		dscore.NewRegions,
		dscontainer.NewBranch,
		dscontainer.NewBranches,
		dscontainer.NewImage,
		dsprotection.NewGatewayPolicy,
		dsprotection.NewGatewayPolicies,
		dsprotection.NewProtocol,
		dsprotection.NewProtocols,
		dsrbac.NewGroup,
		dsrbac.NewGroups,
		dsstorage.NewVolume,
		dsstorage.NewVolumes,
		dsstorage.NewVolumeStore,
		dsstorage.NewVolumeStores,
	}
}

// Resources defines the resources implemented in the provider.
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		armada.NewArmada,
		armada.NewArmadaSet,
		core.NewConfigFile,
		core.NewEnvironment,
		core.NewRegion,
		container.NewBranch,
		container.NewImageUpdater,
		formation.NewFormation,
		formation.NewVessel,
		protection.NewGatewayPolicy,
		rbac.NewGroup,
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
	if cfg.ServiceAccount.ValueString() == "" {
		diags.Append(diag.NewErrorDiagnostic(
			"Missing Service Account",
			"The provider cannot create the GameFabric client as there is no service account configured. "+
				"Please set the service_account value in the provider configuration or use the "+envServiceAccount+" environment variable.",
		))
	}
	if cfg.Password.ValueString() == "" {
		diags.Append(diag.NewErrorDiagnostic(
			"Missing Password",
			"The provider cannot create the GameFabric client as there is no password configured. "+
				"Please set the password value in the provider configuration or use the "+envPassword+" environment variable.",
		))
	}
	return diags
}
