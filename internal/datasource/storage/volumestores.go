package storage

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &volumeStores{}
	_ datasource.DataSourceWithConfigure = &volumeStores{}
)

type volumeStores struct {
	clientSet clientset.Interface
}

// NewVolumeStores creates a new volumestores data source.
func NewVolumeStores() datasource.DataSource { return &volumeStores{} }

// Metadata defines the data source type name.
func (r *volumeStores) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumestores"
}

// Schema defines the schema for this data source.
func (r *volumeStores) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"volumestores": schema.ListNestedAttribute{
				Description:         "A list of volumestores.",
				MarkdownDescription: "A list of volumestores.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "The unique object name.",
							MarkdownDescription: "The unique object name.",
							Computed:            true,
						},
						"region": schema.StringAttribute{
							Description:         "The name of the region the volumestore belongs to.",
							MarkdownDescription: "The name of the region the volumestore belongs to.",
							Computed:            true,
						},
						"max_volume_size": schema.StringAttribute{
							Description:         "The maximum volume size supported by the volumestore.",
							MarkdownDescription: "The maximum volume size supported by the volumestore.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *volumeStores) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read retrieves all volumestores and sets the state.
func (r *volumeStores) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	// no config expected for this data source

	list, err := r.clientSet.StorageV1Beta1().VolumeStore().List(ctx, metav1.ListOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting VolumeStores",
			fmt.Sprintf("Could not get VolumeStores: %v", err),
		)
		return
	}

	slices.SortFunc(list.Items, func(a, b storagev1.VolumeStore) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newVolumeStoresModel(list.Items)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
