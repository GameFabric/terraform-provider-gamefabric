package container

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &branches{}
	_ datasource.DataSourceWithConfigure = &branches{}
)

type branches struct {
	clientSet clientset.Interface
}

// NewBranches creates a new branches data source.
func NewBranches() datasource.DataSource {
	return &branches{}
}

// Metadata defines the data source type name.
func (r *branches) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branches"
}

// Schema defines the schema for this data source.
func (r *branches) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique object name within its scope (filters by exact match).",
				MarkdownDescription: "The unique object name within its scope (filters by exact match).",
				Optional:            true,
			},
			"display_name": schema.StringAttribute{
				Description:         "DisplayName is friendly name of the branch (filters by exact match).",
				MarkdownDescription: "DisplayName is friendly name of the branch (filters by exact match).",
				Optional:            true,
			},
			"label_filter": schema.MapAttribute{
				Description:         "A map of keys and values that is used to filter branches (exact matches of all provided labels).",
				MarkdownDescription: "A map of keys and values that is used to filter branches (exact matches of all provided labels).",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"branches": schema.ListNestedAttribute{
				Description:         "Branches is a list of branches that match the filters.",
				MarkdownDescription: "Branches is a list of branches that match the filters.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *branches) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *branches) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config branchesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build label selector from label_filter
	labelSelector := conv.ForEachMapItem(config.LabelFilter, func(item types.String) string { return item.ValueString() })

	// Get all branches with label filter
	list, err := r.clientSet.ContainerV1().Branches().List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Branches",
			fmt.Sprintf("Could not get Branches: %v", err),
		)
		return
	}

	// Apply additional filters
	filteredBranches := make([]containerv1.Branch, 0)
	displayNameCounts := make(map[string]int)
	for _, branch := range list.Items {
		// Filter by name if specified
		if conv.IsKnown(config.Name) && branch.Name != config.Name.ValueString() {
			continue
		}

		// Filter by display_name if specified
		if conv.IsKnown(config.DisplayName) && branch.Spec.DisplayName != config.DisplayName.ValueString() {
			continue
		}

		filteredBranches = append(filteredBranches, branch)

		// Track display name occurrences for validation
		if conv.IsKnown(config.DisplayName) {
			displayNameCounts[branch.Spec.DisplayName]++
		}
	}

	// Validate that display_name filter doesn't match multiple branches
	if conv.IsKnown(config.DisplayName) {
		if count := displayNameCounts[config.DisplayName.ValueString()]; count > 1 {
			resp.Diagnostics.AddError(
				"Multiple Branches Found",
				fmt.Sprintf("Multiple branches (%d) found with display name %q. Display names are not unique identifiers. Use name or label_filter instead, or ensure display names are unique.", count, config.DisplayName.ValueString()),
			)
			return
		}
	}

	// Sort branches by name for consistent output
	slices.SortFunc(filteredBranches, func(a, b containerv1.Branch) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newBranchesModel(filteredBranches)
	state.Name = config.Name
	state.DisplayName = config.DisplayName
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
