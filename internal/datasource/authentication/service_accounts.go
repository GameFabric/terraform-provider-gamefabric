package authentication

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &serviceAccounts{}
var _ datasource.DataSourceWithConfigure = &serviceAccounts{}

type serviceAccounts struct {
	clientSet clientset.Interface
}

func NewServiceAccounts() datasource.DataSource { return &serviceAccounts{} }

func (r *serviceAccounts) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_accounts"
}

func (r *serviceAccounts) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"label_filter": schema.MapAttribute{
				Description:         "Filter service accounts by labels.",
				MarkdownDescription: "Filter service accounts by labels.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"service_accounts": schema.ListNestedAttribute{
				Description:         "List of service accounts.",
				MarkdownDescription: "List of service accounts.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":   schema.StringAttribute{Computed: true},
						"email":  schema.StringAttribute{Computed: true},
						"labels": schema.MapAttribute{Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

func (r *serviceAccounts) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *serviceAccounts) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config serviceAccountsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	options := metav1.ListOptions{
		LabelSelector: conv.ForEachMapItem(config.LabelFilter, func(item types.String) string {
			return item.ValueString()
		}),
	}
	list, err := r.clientSet.AuthenticationV1Beta1().ServiceAccounts().List(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Service Accounts",
			fmt.Sprintf("Could not get ServiceAccounts: %v", err),
		)
		return
	}

	slices.SortFunc(list.Items, func(a, b v1beta1.ServiceAccount) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newServiceAccountsModel(list.Items)
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
