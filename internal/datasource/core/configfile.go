package core

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
	_ datasource.DataSource              = &configFile{}
	_ datasource.DataSourceWithConfigure = &configFile{}
)

type configFile struct {
	clientSet clientset.Interface
}

// NewConfigFile creates a new config file data source.
func NewConfigFile() datasource.DataSource {
	return &configFile{}
}

// Metadata defines the data source type name.
func (r *configFile) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configfile"
}

// Schema defines the schema for this data source.
func (r *configFile) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique config file name within its environment.",
				MarkdownDescription: "The unique config file name within its environment.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
			},
			"environment": schema.StringAttribute{
				Description:         "The name of the environment the resource belongs to.",
				MarkdownDescription: "The name of the environment the resource belongs to.",
				Required:            true,
				Validators: []validator.String{
					validators.EnvironmentValidator{},
				},
			},
			"data": schema.StringAttribute{
				Description:         "The config file data.",
				MarkdownDescription: "The config file data.",
				Computed:            true,
			},
		},
	}
}

// Configure prepares the struct.
func (r *configFile) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *configFile) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config configFileModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.CoreV1().ConfigFiles(config.Environment.ValueString()).Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Config File Not Found",
				fmt.Sprintf("ConfigFile %q was not found.", config.Name.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Getting Config File",
			fmt.Sprintf("Could not get ConfigFile %q: %v", config.Name.ValueString(), err),
		)
		return
	}

	state := newConfigModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
