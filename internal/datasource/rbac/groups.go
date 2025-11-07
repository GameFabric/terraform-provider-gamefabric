package rbac

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	rbacv1 "github.com/gamefabric/gf-core/pkg/api/rbac/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &groups{}
	_ datasource.DataSourceWithConfigure = &groups{}
)

type groups struct {
	clientSet clientset.Interface
}

// NewGroups returns a new instance of the groups data source.
func NewGroups() datasource.DataSource {
	return &groups{}
}

func (r *groups) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

// Schema defines the schema for this data source.
func (r *groups) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"label_filter": schema.MapAttribute{
				Description:         "A map of keys and values that is used to filter groups.",
				MarkdownDescription: "A map of keys and values that is used to filter groups.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"names": schema.ListAttribute{
				Description:         "A list of group names matching the labels filter.",
				MarkdownDescription: "A list of group names matching the labels filter.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure prepares the struct.
func (r *groups) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *groups) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config groupsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.clientSet.RBACV1().Groups().List(ctx, metav1.ListOptions{
		LabelSelector: conv.ForEachMapItem(config.LabelFilter, func(item types.String) string { return item.ValueString() }),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Groups",
			fmt.Sprintf("Could not get Groups: %v", err),
		)
		return
	}
	slices.SortFunc(list.Items, func(a, b rbacv1.Group) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newGroupsModel(list.Items)
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
