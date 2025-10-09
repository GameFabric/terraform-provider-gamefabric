package core

import (
	"context"
	"fmt"

	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &region{}
	_ datasource.DataSourceWithConfigure = &region{}
)

type region struct {
	clientSet clientset.Interface
}

// NewRegion creates a new region data source.
func NewRegion() datasource.DataSource {
	return &region{}
}

// Metadata defines the data source type name.
func (r *region) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_region"
}

// Schema defines the schema for this data source.
func (r *region) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique object name within its scope.",
				MarkdownDescription: "The unique object name within its scope.",
				Optional:            true,
				Validators: []validator.String{
					validators.NameValidator{},
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("display_name")),
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
			"display_name": schema.StringAttribute{
				Description:         "DisplayName is the user-friendly name of a region.",
				MarkdownDescription: "DisplayName is the user-friendly name of a region.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("name")),
				},
			},
			"labels": schema.MapAttribute{
				Description:         "A map of keys and values that can be used to organize and categorize objects.",
				MarkdownDescription: "A map of keys and values that can be used to organize and categorize objects.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"description": schema.StringAttribute{
				Description:         "Description is the optional description of the region.",
				MarkdownDescription: "Description is the optional description of the region.",
				Computed:            true,
			},
			"types": schema.MapNestedAttribute{
				Description:         "Types defines the types on infrastructure available in the region.",
				MarkdownDescription: "Types defines the types on infrastructure available in the region.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"locations": schema.ListAttribute{
							Description:         "Locations defines the locations for a type.",
							MarkdownDescription: "Locations defines the locations for a type.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"env": schema.ListNestedAttribute{
							Description:         "Env is a list of environment variables to set on all containers in this region.",
							MarkdownDescription: "Env is a list of environment variables to set on all containers in this region.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: EnvVarAttributes(),
							},
						},
						"scheduling": schema.StringAttribute{
							Description:         "Scheduling strategy. Defaults to &#34;Packed&#34;",
							MarkdownDescription: "Scheduling strategy. Defaults to &#34;Packed&#34;",
							Computed:            true,
						},
						"cpu": schema.StringAttribute{
							Description:         "CPU is the CPU limit for the region type.",
							MarkdownDescription: "CPU is the CPU limit for the region type.",
							Computed:            true,
						},
						"memory": schema.StringAttribute{
							Description:         "Memory is the memory limit for the region type.",
							MarkdownDescription: "Memory is the memory limit for the region type.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *region) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *region) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config regionModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.CoreV1().Regions(config.Environment.ValueString()).Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.State.RemoveResource(ctx)
		default:
			resp.Diagnostics.AddError(
				"Error Reading Region",
				fmt.Sprintf("Could not read Region %q: %v", config.Name.ValueString(), err),
			)
		}
		return
	}

	state := newRegionModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
