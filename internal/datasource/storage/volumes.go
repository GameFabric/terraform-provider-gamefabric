package storage

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1beta1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
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
	_ datasource.DataSource              = &volumes{}
	_ datasource.DataSourceWithConfigure = &volumes{}
)

type volumes struct {
	clientSet clientset.Interface
}

// NewVolumes creates a new volumes data source.
func NewVolumes() datasource.DataSource {
	return &volumes{}
}

// Metadata defines the data source type name.
func (r *volumes) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumes"
}

// Schema defines the schema for this data source.
func (r *volumes) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
				Description:         "A map of keys and values that is used to filter volumes.",
				MarkdownDescription: "A map of keys and values that is used to filter volumes.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"volumes": schema.ListNestedAttribute{
				Description:         "Volumes is a list of volumes in the environment that match the labels filter.",
				MarkdownDescription: "Volumes is a list of volumes in the environment that match the labels filter.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "The unique object name within its scope.",
							MarkdownDescription: "The unique object name within its scope.",
							Computed:            true,
						},
						"environment": schema.StringAttribute{
							Description:         "The name of the environment the object belongs to.",
							MarkdownDescription: "The name of the environment the object belongs to.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *volumes) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *volumes) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.clientSet.StorageV1Beta1().Volumes(config.Environment.ValueString()).List(ctx, metav1.ListOptions{
		LabelSelector: conv.ForEachMapItem(config.LabelFilter, func(item types.String) string { return item.ValueString() }),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Volumes",
			fmt.Sprintf("Could not get Volumes: %v", err),
		)
		return
	}
	slices.SortFunc(list.Items, func(a, b storagev1beta1.Volume) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newVolumesModel(list.Items)
	state.Environment = config.Environment
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
