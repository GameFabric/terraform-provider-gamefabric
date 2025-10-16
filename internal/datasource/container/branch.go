package container

import (
	"context"
	_ "embed"
	"fmt"

	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
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
	_ datasource.DataSource              = &branch{}
	_ datasource.DataSourceWithConfigure = &branch{}
)

//go:embed branch.md
var branchSchemaDescription string

type branch struct {
	clientSet clientset.Interface
}

// NewBranch creates a new branch data source.
func NewBranch() datasource.DataSource {
	return &branch{}
}

// Metadata defines the data source type name.
func (r *branch) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch"
}

// Schema defines the schema for this data source.
func (r *branch) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: branchSchemaDescription,
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
			"display_name": schema.StringAttribute{
				Description:         "DisplayName is friendly name of the branch.",
				MarkdownDescription: "DisplayName is friendly name of the branch.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("name")),
				},
			},
			"retention_policy_rules": schema.ListNestedAttribute{
				Description:         "RetentionPolicyRules are the rules that define how images are retained.",
				MarkdownDescription: "RetentionPolicyRules are the rules that define how images are retained.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Name is the name of the image retention policy.",
							MarkdownDescription: "Name is the name of the image retention policy.",
							Computed:            true,
						},
						"image_regex": schema.StringAttribute{
							Description:         "ImageRegex is the optional regex selector for images that this policy applies to.",
							MarkdownDescription: "ImageRegex is the optional regex selector for images that this policy applies to.",
							Computed:            true,
						},
						"tag_regex": schema.StringAttribute{
							Description:         "TagRegex is the optional regex selector for tags that this policy applies to.",
							MarkdownDescription: "TagRegex is the optional regex selector for tags that this policy applies to.",
							Computed:            true,
						},
						"keep_count": schema.Int64Attribute{
							Description:         "KeepCount is the minimum number of tags to keep per image.",
							MarkdownDescription: "KeepCount is the minimum number of tags to keep per image.",
							Computed:            true,
						},
						"keep_days": schema.Int64Attribute{
							Description:         "KeepDays is the minimum number of days an image tag must be kept for.",
							MarkdownDescription: "KeepDays is the minimum number of days an image tag must be kept for.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *branch) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *branch) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config branchModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		obj *containerv1.Branch
		err error
	)
	switch {
	case conv.IsKnown(config.Name):
		obj, err = r.clientSet.ContainerV1().Branches().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				resp.Diagnostics.AddError(
					"Branch Not Found",
					fmt.Sprintf("No branch found with name %q", config.Name.ValueString()),
				)
				return
			}
			resp.Diagnostics.AddError(
				"Error Getting Branch",
				fmt.Sprintf("Could not get Branch %q: %v", config.Name.ValueString(), err),
			)
			return
		}
	case conv.IsKnown(config.DisplayName):
		list, err := r.clientSet.ContainerV1().Branches().List(ctx, metav1.ListOptions{})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Getting Branch",
				fmt.Sprintf("Could not get Branch with display name %q: %v", config.DisplayName.ValueString(), err),
			)
			return
		}
		for _, item := range list.Items {
			if item.Spec.DisplayName != config.DisplayName.ValueString() {
				continue
			}
			if obj != nil {
				resp.Diagnostics.AddError(
					"Multiple Branches Found",
					fmt.Sprintf("Multiple branches found with display name %q", config.DisplayName.ValueString()),
				)
				return
			}
			obj = &item
		}
		if obj == nil {
			resp.Diagnostics.AddError(
				"Branch Not Found",
				"No branch found matching the specified criteria.",
			)
			return
		}
	default:
		resp.Diagnostics.AddError(
			"Insufficient Search Criteria",
			"Either name or display_name must be specified.",
		)
		return
	}

	state := newBranchModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
