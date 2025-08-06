package core

import (
	"context"
	"fmt"

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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &environment{}
	_ datasource.DataSourceWithConfigure = &environment{}
)

type environment struct {
	clientSet clientset.Interface
}

func NewEnvironment() datasource.DataSource {
	return &environment{}
}

func (r *environment) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

// Schema defines the schema for this data source.
func (r *environment) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique object name within its scope.",
				MarkdownDescription: "The unique object name within its scope.",
				Optional:            true,
				Validators: []validator.String{
					validators.EnvironmentValidator{},
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("display_name")),
				},
			},
			"display_name": schema.StringAttribute{
				Description:         "DisplayName is friendly name of the environment.",
				MarkdownDescription: "DisplayName is friendly name of the environment.",
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
				Description:         "Description is the description of the environment.",
				MarkdownDescription: "Description is the description of the environment.",
				Computed:            true,
			},
		},
	}
}

// Configure prepares the struct.
func (r *environment) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	return
}

func (r *environment) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config environmentModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		obj *corev1.Environment
		err error
	)
	switch {
	case conv.IsKnown(config.Name):
		obj, err = r.clientSet.CoreV1().Environments().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Getting Environment",
				fmt.Sprintf("Could not get Environment %q: %v", config.Name.ValueString(), err),
			)
			return
		}
	case conv.IsKnown(config.DisplayName):
		list, err := r.clientSet.CoreV1().Environments().List(ctx, metav1.ListOptions{})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Getting Environment",
				fmt.Sprintf("Could not get Environment %q: %v", config.Name.ValueString(), err),
			)
			return
		}
		for _, item := range list.Items {
			if item.Spec.DisplayName != config.DisplayName.ValueString() {
				continue
			}
			if obj != nil {
				resp.Diagnostics.AddError(
					"Multiple Environments Found",
					fmt.Sprintf("Multiple environments found with display name %q", config.DisplayName.ValueString()),
				)
				return
			}
			obj = &item
		}
	}

	state := newEnvironmentModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
