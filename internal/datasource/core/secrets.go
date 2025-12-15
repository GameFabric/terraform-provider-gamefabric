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
	_ datasource.DataSource              = &secrets{}
	_ datasource.DataSourceWithConfigure = &secrets{}
)

type secrets struct {
	clientSet clientset.Interface
}

// NewSecrets creates a new secrets data source.
func NewSecrets() datasource.DataSource {
	return &secrets{}
}

// Metadata defines the data source type name.
func (r *secrets) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secrets"
}

// Schema defines the schema for this data source.
func (r *secrets) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment": schema.StringAttribute{
				Description:         "The name of the environment the secrets belong to.",
				MarkdownDescription: "The name of the environment the secrets belong to.",
				Required:            true,
				Validators: []validator.String{
					validators.EnvironmentValidator{},
				},
			},
			"label_filter": schema.MapAttribute{
				Description:         "A map of keys and values that is used to filter secrets. Only items with all specified labels (exact matches) will be returned.",
				MarkdownDescription: "A map of keys and values that is used to filter secrets. Only items with all specified labels (exact matches) will be returned.",
				Optional:            true,
				ElementType:         types.StringType,
				Validators: []validator.Map{
					validators.LabelsValidator{},
				},
			},
			"secrets": schema.ListNestedAttribute{
				Description:         "Secrets is a list of secrets in the environment that match the labels filter.",
				MarkdownDescription: "Secrets is a list of secrets in the environment that match the labels filter.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"environment": schema.StringAttribute{
							Description:         "The name of the environment the secret belongs to.",
							MarkdownDescription: "The name of the environment the secret belongs to.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							Description:         "The unique secret name within its environment.",
							MarkdownDescription: "The unique secret name within its environment.",
							Computed:            true,
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
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *secrets) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *secrets) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config secretsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.clientSet.CoreV1().Secrets(config.Environment.ValueString()).List(ctx, metav1.ListOptions{
		LabelSelector: conv.ForEachMapItem(config.LabelFilter, func(item types.String) string { return item.ValueString() }),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Secrets",
			fmt.Sprintf("Could not get Secrets: %v", err),
		)
		return
	}
	slices.SortFunc(list.Items, func(a, b corev1.Secret) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newSecretsModel(list.Items)
	state.Environment = config.Environment
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
