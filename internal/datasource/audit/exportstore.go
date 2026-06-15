package audit

import (
	"context"
	"fmt"

	apierrors "github.com/gamefabric/gf-apicore/api/errors"
	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ datasource.DataSource              = &exportStore{}
	_ datasource.DataSourceWithConfigure = &exportStore{}
)

type exportStore struct {
	clientSet clientset.Interface
}

// NewExportStore creates a new exportstore data source.
func NewExportStore() datasource.DataSource { return &exportStore{} }

// Metadata defines the data source type name.
func (r *exportStore) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_exportstore"
}

// Schema defines the schema for this data source.
func (r *exportStore) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"suspended": schema.BoolAttribute{
				Description:         "Whether audit log export is suspended.",
				MarkdownDescription: "Whether audit log export is suspended.",
				Computed:            true,
			},
			"s3": schema.SingleNestedAttribute{
				Description:         "S3-compatible storage configuration.",
				MarkdownDescription: "S3-compatible storage configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"endpoint": schema.StringAttribute{
						Description:         "The S3-compatible endpoint URL.",
						MarkdownDescription: "The S3-compatible endpoint URL.",
						Computed:            true,
					},
					"region": schema.StringAttribute{
						Description:         "The storage region.",
						MarkdownDescription: "The storage region.",
						Computed:            true,
					},
					"bucket": schema.StringAttribute{
						Description:         "The bucket name.",
						MarkdownDescription: "The bucket name.",
						Computed:            true,
					},
					"prefix": schema.StringAttribute{
						Description:         "An optional key prefix within the bucket.",
						MarkdownDescription: "An optional key prefix within the bucket.",
						Computed:            true,
					},
					"auth": schema.SingleNestedAttribute{
						Description:         "Authentication configuration for the S3-compatible store.",
						MarkdownDescription: "Authentication configuration for the S3-compatible store.",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"access_key_id": schema.StringAttribute{
								Description:         "The access key ID.",
								MarkdownDescription: "The access key ID.",
								Computed:            true,
							},
							"secret_access_key": schema.StringAttribute{
								Description:         "The secret access key. Returned masked by the API.",
								MarkdownDescription: "The secret access key. Returned masked by the API.",
								Computed:            true,
								Sensitive:           true,
							},
						},
					},
				},
			},
			"state": schema.StringAttribute{
				Description:         "The current operational state of the export store.",
				MarkdownDescription: "The current operational state of the export store.",
				Computed:            true,
			},
			"message": schema.StringAttribute{
				Description:         "A human-readable message with additional detail.",
				MarkdownDescription: "A human-readable message with additional detail.",
				Computed:            true,
			},
		},
	}
}

// Configure prepares the struct.
func (r *exportStore) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read retrieves the exportstore and sets the state.
func (r *exportStore) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config exportStoreModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	obj, err := r.clientSet.AuditV1Alpha1().ExportStores().Get(ctx, config.Name.ValueString(), metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"ExportStore Not Found",
				fmt.Sprintf("ExportStore %q was not found.", config.Name.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error Getting ExportStore",
			fmt.Sprintf("Could not get ExportStore %q: %v", config.Name.ValueString(), err),
		)
		return
	}

	state := newExportStoreModel(obj)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
