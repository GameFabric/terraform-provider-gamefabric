package provisioning

import (
	"context"
	"fmt"
	"maps"

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
	_ datasource.DataSource              = &tokenService{}
	_ datasource.DataSourceWithConfigure = &tokenService{}
)

type tokenService struct {
	clientSet clientset.Interface
}

// NewTokenService creates a new token service data source.
func NewTokenService() datasource.DataSource {
	return &tokenService{}
}

// Metadata defines the data source type name.
func (r *tokenService) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_steelshield_tokenservice"
}

// Schema defines the schema for this data source.
func (r *tokenService) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description:         "The unique object name within its scope.",
			MarkdownDescription: "The unique object name within its scope.",
			Required:            true,
			Validators: []validator.String{
				validators.NameValidator{},
			},
		},
	}
	maps.Copy(attrs, tokenServiceComputedAttributes())

	resp.Schema = schema.Schema{
		Description:         "Data source for a single Token Service.",
		MarkdownDescription: "Data source for a single Token Service.",
		Attributes:          attrs,
	}
}

// Configure prepares the struct.
func (r *tokenService) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *tokenService) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config tokenServiceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.ProvisioningV1Beta1().TokenServices().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Token Service Not Found",
				fmt.Sprintf("Token Service %q was not found.", config.Name.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Getting Token Service",
			fmt.Sprintf("Could not get Token Service %q: %v", config.Name.ValueString(), err),
		)
		return
	}

	state := newTokenServiceModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// tokenServiceComputedAttributes returns the schema attributes shared between the singular
// gamefabric_steelshield_tokenservice data source and the nested items of the plural
// gamefabric_steelshield_tokenservices data source. Every attribute is Computed; "name" is intentionally
// excluded since it is Required on the singular data source and Computed on the plural one.
func tokenServiceComputedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"development_mode": schema.BoolAttribute{
			Description:         "Whether the token service runs in development mode. When true, the token service returns errors to game clients; when false, it hides them.",
			MarkdownDescription: "Whether the token service runs in development mode. When `true`, the token service returns errors to game clients; when `false`, it hides them.",
			Computed:            true,
		},
		"labels": schema.MapAttribute{
			Description:         "A map of keys and values that can be used to organize and categorize objects.",
			MarkdownDescription: "A map of keys and values that can be used to organize and categorize objects.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"annotations": schema.MapAttribute{
			Description:         "Annotations is an unstructured map of keys and values stored on an object.",
			MarkdownDescription: "Annotations is an unstructured map of keys and values stored on an object.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"game_name": schema.StringAttribute{
			Description:         "The name of the game.",
			MarkdownDescription: "The name of the game.",
			Computed:            true,
		},
		"platforms": schema.MapNestedAttribute{
			Description:         "Maps platform names to their game client token verification configuration.",
			MarkdownDescription: "Maps platform names to their game client token verification configuration.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"game_client_token_keys": schema.ListNestedAttribute{
						Description:         "A list of keys used to verify game client tokens for this platform.",
						MarkdownDescription: "A list of keys used to verify game client tokens for this platform.",
						Computed:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"key": schema.StringAttribute{
									Description:         "The key material used to verify tokens.",
									MarkdownDescription: "The key material used to verify tokens.",
									Computed:            true,
									Sensitive:           true,
								},
								"signing_algorithm": schema.StringAttribute{
									Description:         "The signing algorithm used with this key.",
									MarkdownDescription: "The signing algorithm used with this key.",
									Computed:            true,
								},
							},
						},
					},
				},
			},
		},
		"eos": schema.ListNestedAttribute{
			Description:         "Epic Online Services (EOS) auth mode configurations, only present in EOS mode.",
			MarkdownDescription: "Epic Online Services (EOS) auth mode configurations, only present in EOS mode.",
			Computed:            true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						Description:         "The EOS client ID.",
						MarkdownDescription: "The EOS client ID.",
						Computed:            true,
					},
					"token_types": schema.ListAttribute{
						Description:         "The accepted EOS token types.",
						MarkdownDescription: "The accepted EOS token types.",
						Computed:            true,
						ElementType:         types.StringType,
					},
					"deployment_id": schema.StringAttribute{
						Description:         "The EOS deployment ID.",
						MarkdownDescription: "The EOS deployment ID.",
						Computed:            true,
					},
					"product_id": schema.StringAttribute{
						Description:         "The EOS product ID.",
						MarkdownDescription: "The EOS product ID.",
						Computed:            true,
					},
					"sandbox_id": schema.StringAttribute{
						Description:         "The EOS sandbox ID.",
						MarkdownDescription: "The EOS sandbox ID.",
						Computed:            true,
					},
				},
			},
		},
		"jwks": schema.SingleNestedAttribute{
			Description:         "JWKS endpoint auth mode configuration, only present in JWKS mode.",
			MarkdownDescription: "JWKS endpoint auth mode configuration, only present in JWKS mode.",
			Computed:            true,
			Attributes: map[string]schema.Attribute{
				"url": schema.StringAttribute{
					Description:         "The JWKS endpoint URL.",
					MarkdownDescription: "The JWKS endpoint URL.",
					Computed:            true,
				},
			},
		},
		"state": schema.StringAttribute{
			Description:         "The high-level lifecycle state of the token service. One of \"Pending\", \"Available\", or \"Error\".",
			MarkdownDescription: "The high-level lifecycle state of the token service. One of `Pending`, `Available`, or `Error`.",
			Computed:            true,
		},
		"hostname": schema.StringAttribute{
			Description:         "The fully-qualified domain name assigned to this token service instance.",
			MarkdownDescription: "The fully-qualified domain name assigned to this token service instance.",
			Computed:            true,
		},
		"platform_key": schema.StringAttribute{
			Description:         "The platform key used for all platforms.",
			MarkdownDescription: "The platform key used for all platforms.",
			Computed:            true,
			Sensitive:           true,
		},
		"reason": schema.StringAttribute{
			Description:         "The failure reason. Null when there is no failure.",
			MarkdownDescription: "The failure reason. Null when there is no failure.",
			Computed:            true,
		},
		"state_last_changed": schema.StringAttribute{
			Description:         "The timestamp of the last state change, in RFC3339 format.",
			MarkdownDescription: "The timestamp of the last state change, in RFC3339 format.",
			Computed:            true,
		},
	}
}
