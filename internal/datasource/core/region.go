package core

import (
	"context"
	_ "embed"
	"fmt"

	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
			"types": schema.MapNestedAttribute{
				Description:         "Types defines the types on infrastructure available in the region.",
				MarkdownDescription: "Types defines the types on infrastructure available in the region.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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

	var (
		obj *corev1.Region
		err error
	)
	switch {
	case conv.IsKnown(config.Name):
		obj, err = r.clientSet.CoreV1().Regions(config.Environment.ValueString()).Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				resp.Diagnostics.AddError(
					"Region Not Found",
					fmt.Sprintf("Region %q was not found.", config.Name.ValueString()),
				)
				return
			}
			resp.Diagnostics.AddError(
				"Error Getting Region",
				fmt.Sprintf("Could not get Region %q: %v", config.Name.ValueString(), err),
			)
			return
		}
	case conv.IsKnown(config.DisplayName):
		list, err := r.clientSet.CoreV1().Regions(config.Environment.ValueString()).List(ctx, metav1.ListOptions{})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Getting Region",
				fmt.Sprintf("Could not get Region with display name %q: %v", config.DisplayName.ValueString(), err),
			)
			return
		}
		for _, item := range list.Items {
			if item.Spec.DisplayName != config.DisplayName.ValueString() {
				continue
			}
			if obj != nil {
				resp.Diagnostics.AddError(
					"Multiple Regions Found",
					fmt.Sprintf("Multiple regions found with display name %q", config.DisplayName.ValueString()),
				)
				return
			}
			obj = &item
		}
		if obj == nil {
			resp.Diagnostics.AddError(
				"Region Not Found",
				fmt.Sprintf("Region with display name %q was not found.", config.DisplayName.ValueString()),
			)
			return
		}
	default:
		resp.Diagnostics.AddError(
			"Insufficient Information",
			"Either name or display_name must be specified.",
		)
		return
	}

	state := newRegionModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
