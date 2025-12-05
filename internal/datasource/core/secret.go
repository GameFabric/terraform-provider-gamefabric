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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &secret{}
	_ datasource.DataSourceWithConfigure = &secret{}
)

type secret struct {
	clientSet clientset.Interface
}

// NewSecret creates a new secret data source.
func NewSecret() datasource.DataSource {
	return &secret{}
}

// Metadata defines the data source type name.
func (r *secret) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

// Schema defines the schema for this data source.
func (r *secret) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment": schema.StringAttribute{
				Description:         "The name of the environment the secret belongs to.",
				MarkdownDescription: "The name of the environment the secret belongs to.",
				Required:            true,
				Validators: []validator.String{
					validators.EnvironmentValidator{},
				},
			},
			"name": schema.StringAttribute{
				Description:         "The unique secret name within its environment.",
				MarkdownDescription: "The unique secret name within its environment.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
			},
			"description": schema.StringAttribute{
				Description:         "Description of the secret.",
				MarkdownDescription: "Description of the secret.",
				Computed:            true,
			},
			"data": schema.ListAttribute{
				Description:         "Data contains the secret keys.",
				MarkdownDescription: "Data contains the secret keys.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"labels": schema.MapAttribute{
				Description:         "Labels are key/value pairs that can be used to organize and categorize objects.",
				MarkdownDescription: "Labels are key/value pairs that can be used to organize and categorize objects.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

// Configure prepares the struct.
func (r *secret) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *secret) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config secretModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.CoreV1().Secrets(config.Environment.ValueString()).Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			resp.Diagnostics.AddError(
				"Secret Not Found",
				fmt.Sprintf("Secret %q was not found.", config.Name.ValueString()),
			)
			return
		default:
			resp.Diagnostics.AddError(
				"Error Getting Secret",
				fmt.Sprintf("Could not get Secret %q: %v", config.Name.ValueString(), err),
			)
			return
		}
	}

	state := newSecretModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
