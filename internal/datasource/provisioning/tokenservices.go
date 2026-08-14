package provisioning

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	provisioningv1beta1 "github.com/gamefabric/gf-core/pkg/api/provisioning/v1beta1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &tokenServices{}
	_ datasource.DataSourceWithConfigure = &tokenServices{}
)

type tokenServices struct {
	clientSet clientset.Interface
}

// NewTokenServices returns a new instance of the token services data source.
func NewTokenServices() datasource.DataSource {
	return &tokenServices{}
}

// Metadata defines the data source type name.
func (r *tokenServices) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_steelshield_tokenservices"
}

// Schema defines the schema for this data source.
func (r *tokenServices) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description:         "The unique object name within its scope.",
			MarkdownDescription: "The unique object name within its scope.",
			Computed:            true,
		},
	}
	maps.Copy(itemAttrs, tokenServiceComputedAttributes())

	resp.Schema = schema.Schema{
		Description:         "Data source for a list of Token Services.",
		MarkdownDescription: "Data source for a list of Token Services.",
		Attributes: map[string]schema.Attribute{
			"label_filter": schema.MapAttribute{
				Description:         "A map of keys and values that is used to filter token services. Only items with all specified labels (exact matches) will be returned.",
				MarkdownDescription: "A map of keys and values that is used to filter token services. Only items with all specified labels (exact matches) will be returned.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"token_services": schema.ListNestedAttribute{
				Description:         "The token services that match the label filter.",
				MarkdownDescription: "The token services that match the label filter.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: itemAttrs,
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *tokenServices) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	procCtx, ok := req.ProviderData.(*provcontext.Context)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *provider.Context, got %T", req.ProviderData),
		)
		return
	}

	r.clientSet = procCtx.ClientSet
}

func (r *tokenServices) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tokenServicesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.clientSet.ProvisioningV1Beta1().TokenServices().List(ctx, metav1.ListOptions{
		LabelSelector: conv.ForEachMapItem(config.LabelFilter, func(item types.String) string { return item.ValueString() }),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Token Services",
			fmt.Sprintf("Could not get Token Services: %v", err),
		)
		return
	}
	slices.SortFunc(list.Items, func(a, b provisioningv1beta1.TokenService) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newTokenServicesModel(list.Items)
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
