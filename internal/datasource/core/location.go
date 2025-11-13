package core

import (
	"context"
	"fmt"

	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &location{}
	_ datasource.DataSourceWithConfigure = &location{}
)

type location struct {
	clientSet clientset.Interface
}

// NewLocation returns a new instance of the location data source.
func NewLocation() datasource.DataSource {
	return &location{}
}

func (r *location) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

// Schema defines the schema for this data source.
func (r *location) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique location object name.",
				MarkdownDescription: "The unique location object name.",
				Required:            true,
			},
		},
	}
}

// Configure prepares the struct.
func (r *location) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *location) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config locationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.clientSet.CoreV1().Locations().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Location Not Found",
				fmt.Sprintf("Location %q was not found.", config.Name.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Getting Location",
			fmt.Sprintf("Could not get Location %q: %v", config.Name.ValueString(), err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
