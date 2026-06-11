package notification

import (
	"context"
	"fmt"

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
	_ datasource.DataSource              = &receiver{}
	_ datasource.DataSourceWithConfigure = &receiver{}
)

type receiver struct {
	clientSet clientset.Interface
}

// NewReceiverDataSource creates a new receiver data source.
func NewReceiverDataSource() datasource.DataSource {
	return &receiver{}
}

// Metadata defines the data source type name.
func (r *receiver) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_receiver"
}

// Schema defines the schema for this data source.
func (r *receiver) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Data source for reading a Notification Receiver in GameFabric.",
		MarkdownDescription: "Data source for reading a Notification Receiver in GameFabric.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique object name of the receiver.",
				MarkdownDescription: "The unique object name of the receiver.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
			},
			"labels": schema.MapAttribute{
				Description:         "A map of keys and values that can be used to organize and categorize objects.",
				MarkdownDescription: "A map of keys and values that can be used to organize and categorize objects.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"annotations": schema.MapAttribute{
				Description:         "A map of keys and values that can be used to organize and categorize objects.",
				MarkdownDescription: "A map of keys and values that can be used to organize and categorize objects.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"email_to": schema.ListAttribute{
				Description:         "The list of email addresses to send notifications to.",
				MarkdownDescription: "The list of email addresses to send notifications to.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure prepares the struct.
func (r *receiver) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *receiver) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config receiverModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.NotificationV1Alpha1().Receivers().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Receiver",
			fmt.Sprintf("Could not get Receiver %q: %v", config.Name.ValueString(), err),
		)
		return
	}

	state := newReceiverModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
