package audit

import (
	"context"
	"fmt"
	"slices"
	"strings"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	auditv1alpha1 "github.com/gamefabric/gf-core/pkg/api/audit/v1alpha1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var (
	_ datasource.DataSource              = &exportStores{}
	_ datasource.DataSourceWithConfigure = &exportStores{}
)

type exportStores struct {
	clientSet clientset.Interface
}

// NewExportStores creates a new exportstores data source.
func NewExportStores() datasource.DataSource { return &exportStores{} }

func (r *exportStores) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_exportstores"
}

func (r *exportStores) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"exportstores": schema.ListNestedAttribute{
				Description:         "A list of export stores.",
				MarkdownDescription: "A list of export stores.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "The unique object name.",
							MarkdownDescription: "The unique object name.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (r *exportStores) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *exportStores) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	list, err := r.clientSet.AuditV1Alpha1().ExportStores().List(ctx, metav1.ListOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting ExportStores",
			fmt.Sprintf("Could not get ExportStores: %v", err),
		)
		return
	}

	slices.SortFunc(list.Items, func(a, b auditv1alpha1.ExportStore) int {
		return strings.Compare(a.Name, b.Name)
	})

	state := newExportStoresModel(list.Items)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
