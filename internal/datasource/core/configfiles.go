package core

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &configFiles{}
	_ datasource.DataSourceWithConfigure = &configFiles{}
)

type configFiles struct {
	clientSet clientset.Interface
}

// NewConfigFiles creates a new config files data source.
func NewConfigFiles() datasource.DataSource {
	return &configFiles{}
}

// Metadata defines the data source type name.
func (r *configFiles) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configfiles"
}

// Schema defines the schema for this data source.
func (r *configFiles) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment": schema.StringAttribute{
				Description:         "The name of the environment the object belongs to.",
				MarkdownDescription: "The name of the environment the object belongs to.",
				Required:            true,
				Validators: []validator.String{
					validators.EnvironmentValidator{},
				},
			},
			"label_filter": schema.MapAttribute{
				Description:         "A map of keys and values that is used to filter regions.",
				MarkdownDescription: "A map of keys and values that is used to filter regions.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"config_files": schema.ListNestedAttribute{
				Description:         "ConfigFiles is a list of config files in the environment that match the labels filter.",
				MarkdownDescription: "ConfigFiles is a list of config files in the environment that match the labels filter.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "The unique object name.",
							MarkdownDescription: "The unique object name.",
							Computed:            true,
						},
						"environment": schema.StringAttribute{
							Description:         "The name of the environment the object belongs to.",
							MarkdownDescription: "The name of the environment the object belongs to.",
							Computed:            true,
						},
						"data": schema.StringAttribute{
							Description:         "The config file data.",
							MarkdownDescription: "The config file data.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *configFiles) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *configFiles) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config configFilesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.clientSet.CoreV1().ConfigFiles(config.Environment.ValueString()).List(ctx, metav1.ListOptions{
		LabelSelector: conv.ForEachMapItem(config.LabelFilter, func(item types.String) string { return item.ValueString() }),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Config Files",
			fmt.Sprintf("Could not get ConfigFiles: %v", err),
		)
		return
	}
	slices.SortFunc(list.Items, func(a, b corev1.ConfigFile) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newConfigFilesModel(list.Items)
	state.Environment = config.Environment
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
