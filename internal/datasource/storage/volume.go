package storage

import (
	"context"
	"fmt"

	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ datasource.DataSource              = &volume{}
	_ datasource.DataSourceWithConfigure = &volume{}
)

type volume struct {
	clientSet clientset.Interface
}

// NewVolume creates a new volume data source.
func NewVolume() datasource.DataSource {
	return &volume{}
}

// Metadata defines the data source type name.
func (r *volume) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

// Schema defines the schema for this data source.
func (r *volume) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique object name.",
				MarkdownDescription: "The unique object name.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
			},
			"environment": schema.StringAttribute{
				Description:         "The name of the environment the object belongs to.",
				MarkdownDescription: "The name of the environment the object belongs to.",
				Required:            true,
				Validators: []validator.String{
					validators.EnvironmentValidator{},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *volume) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *volume) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumeModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.StorageV1Beta1().Volumes(config.Environment.ValueString()).Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Volume Not Found",
				fmt.Sprintf("Volume %q was not found.", config.Name.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Getting Volume",
			fmt.Sprintf("Could not get Volume %q: %v", config.Name.ValueString(), err),
		)
		return
	}

	state := newVolumeModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
