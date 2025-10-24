package storage

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
)

var (
	_ datasource.DataSource              = &volumeStore{}
	_ datasource.DataSourceWithConfigure = &volumeStore{}
)

type volumeStore struct {
	clientSet clientset.Interface
}

// NewVolumeStore creates a new volumestore data source.
func NewVolumeStore() datasource.DataSource { return &volumeStore{} }

// Metadata defines the data source type name.
func (r *volumeStore) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumestore"
}

// Schema defines the schema for this data source.
func (r *volumeStore) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique object name.",
				MarkdownDescription: "The unique object name.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
			},
			"region": schema.StringAttribute{
				Description:         "The name of the region the volumestore belongs to.",
				MarkdownDescription: "The name of the region the volumestore belongs to.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
			},
			"max_volume_size": schema.StringAttribute{
				Description:         "The maximum volume size supported by the volumestore.",
				MarkdownDescription: "The maximum volume size supported by the volumestore.",
				Computed:            true,
			},
		},
	}
}

// Configure prepares the struct.
func (r *volumeStore) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read retrieves the volumestore and sets the state.
func (r *volumeStore) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumeStoreModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// VolumeStore is a global resource; fetch by name using the storage client.
	obj, err := r.clientSet.StorageV1Beta1().VolumeStore().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting VolumeStore",
			fmt.Sprintf("Could not get VolumeStore %q: %v", config.Name.ValueString(), err),
		)
		return
	}

	state := newVolumeStoreModel(obj)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
