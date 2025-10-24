package storage

import (
	"context"
	"fmt"

	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	storagev1 "github.com/gamefabric/gf-core/pkg/api/storage/v1beta1"
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
	_ datasource.DataSource              = &volumeStore{}
	_ datasource.DataSourceWithConfigure = &volumeStore{}
)

type volumeStore struct {
	clientSet clientset.Interface
}

// NewVolumeStore creates a new volumestore data source.
func NewVolumeStore() datasource.DataSource { return &volumeStore{} }

func (r *volumeStore) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volumestore"
}

func (r *volumeStore) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique object name.",
				MarkdownDescription: "The unique object name.",
				Optional:            true,
				Validators: []validator.String{
					validators.NameValidator{},
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("region")),
				},
			},
			"region": schema.StringAttribute{
				Description:         "The name of the region the volumestore belongs to.",
				MarkdownDescription: "The name of the region the volumestore belongs to.",
				Optional:            true,
				Validators: []validator.String{
					validators.NameValidator{},
					stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("name")),
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

func (r *volumeStore) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config volumeStoreModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		obj *storagev1.VolumeStore
		err error
	)

	switch {
	case conv.IsKnown(config.Name):
		obj, err = r.clientSet.StorageV1Beta1().VolumeStore().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				resp.Diagnostics.AddError(
					"VolumeStore Not Found",
					fmt.Sprintf("VolumeStore %q was not found.", config.Name.ValueString()),
				)
				return
			}
			resp.Diagnostics.AddError(
				"Error Getting VolumeStore",
				fmt.Sprintf("Could not get VolumeStore %q: %v", config.Name.ValueString(), err),
			)
			return
		}
	case conv.IsKnown(config.Region):
		list, lerr := r.clientSet.StorageV1Beta1().VolumeStore().List(ctx, metav1.ListOptions{})
		if lerr != nil {
			resp.Diagnostics.AddError(
				"Error Getting VolumeStores",
				fmt.Sprintf("Could not list VolumeStores: %v", lerr),
			)
			return
		}
		for _, item := range list.Items {
			if item.Spec.Region != config.Region.ValueString() {
				continue
			}
			if obj != nil {
				resp.Diagnostics.AddError(
					"Multiple VolumeStores Found",
					fmt.Sprintf("Multiple VolumeStores found in region %q", config.Region.ValueString()),
				)
				return
			}
			obj = &item
		}
		if obj == nil {
			resp.Diagnostics.AddError(
				"VolumeStore Not Found",
				fmt.Sprintf("VolumeStore in region %q was not found.", config.Region.ValueString()),
			)
			return
		}
	default:
		resp.Diagnostics.AddError(
			"Insufficient Information",
			"Either name or region must be specified.",
		)
		return
	}

	state := newVolumeStoreModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
