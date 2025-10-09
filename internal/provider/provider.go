package provider

import (
	"context"

	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	dscore "github.com/gamefabric/terraform-provider-gamefabric/internal/datasource/core"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/resource/core"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &Provider{}

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
		Description: "GameFabric Terraform Provider",
		Attributes:  map[string]schema.Attribute{},
	}
}

// Configure prepares the provider for data sources and resources.
func (p *Provider) Configure(_ context.Context, _ provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// TODO: Implement provider configuration.

	provCtx := provcontext.NewContext(p.clientSet)
	resp.DataSourceData = provCtx
	resp.ResourceData = provCtx
}

// DataSources defines the data sources implemented in the provider.
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		dscore.NewEnvironment,
		dscore.NewRegion,
		dscore.NewRegions,
	}
}

// Resources defines the resources implemented in the provider.
func (p *Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		core.NewEnvironment,
		core.NewRegion,
	}
}
