package authentication

import (
	"context"
	"fmt"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
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
			"label_filters": schema.MapAttribute{
				Description:         "Filter service accounts by labels.",
				MarkdownDescription: "Filter service accounts by labels.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"items": schema.ListNestedAttribute{
				Description:         "List of service accounts.",
				MarkdownDescription: "List of service accounts.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":     schema.StringAttribute{Computed: true},
						"labels":   schema.MapAttribute{Computed: true, ElementType: types.StringType},
						"username": schema.StringAttribute{Computed: true},
						"email":    schema.StringAttribute{Computed: true},
						"state":    schema.StringAttribute{Computed: true},
						"password": schema.StringAttribute{Computed: true, Sensitive: true},
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
	var config serviceAccountsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labelFilters := make(map[string]string)
	for k, v := range config.LabelFilters {
		if !v.IsNull() && !v.IsUnknown() {
			labelFilters[k] = v.ValueString()
		}
	}

	list, err := r.clientSet.AuthenticationV1Beta1().ServiceAccounts().List(ctx, metav1.ListOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Listing Service Accounts",
			fmt.Sprintf("Could not list ServiceAccounts: %v", err),
		)
		return
	}

	items := newServiceAccountsModel(list, labelFilters)
	state := serviceAccountsDataSourceModel{
		LabelFilters: config.LabelFilters,
		Items:        items,
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
