package container

import (
	"context"
	"fmt"
	"slices"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/validators"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ datasource.DataSource              = &image{}
	_ datasource.DataSourceWithConfigure = &image{}
)

type image struct {
	clientSet clientset.Interface
}

// NewImage creates a new image data source.
func NewImage() datasource.DataSource {
	return &image{}
}

// Metadata defines the data source type name.
func (r *image) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

// Schema defines the schema for this data source.
func (r *image) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"branch": schema.StringAttribute{
				Description:         "The branch in which to find the image.",
				MarkdownDescription: "The branch in which to find the image.",
				Required:            true,
				Validators: []validator.String{
					validators.NameValidator{},
				},
			},
			"image": schema.StringAttribute{
				Description:         "The name of the container image.",
				MarkdownDescription: "The name of the container image.",
				Required:            true,
			},
			"tag": schema.StringAttribute{
				Description:         "The tag or version of the container image.",
				MarkdownDescription: "The tag or version of the container image.",
				Required:            true,
			},
			"image_ref": schema.SingleNestedAttribute{
				Description:         "Provides information about the resolved image reference, such as its unique name within the branch and the branch in which it is found.",
				MarkdownDescription: "Provides information about the resolved image reference, such as its unique name within the branch and the branch in which it is found.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Description:         "The unique image name within its branch.",
						MarkdownDescription: "The unique image name within its branch.",
						Computed:            true,
					},
					"branch": schema.StringAttribute{
						Description:         "The branch in which to find the image.",
						MarkdownDescription: "The branch in which to find the image.",
						Computed:            true,
					},
				},
			},
		},
	}
}

// Configure prepares the struct.
func (r *image) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *image) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config imageModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fieldSelector := map[string]string{
		"spec.image": config.Image.ValueString(),
	}
	if tag := config.Tag.ValueString(); tag != "latest" {
		fieldSelector["spec.tag"] = tag
	}

	objs, err := r.clientSet.ContainerV1().Images(config.Branch.ValueString()).List(ctx, metav1.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Image",
			fmt.Sprintf("Could not get Image %q: %v", config.Image.ValueString(), err),
		)
		return
	}
	if len(objs.Items) == 0 {
		switch {
		case len(fieldSelector) == 1:
			resp.Diagnostics.AddError(
				"Image Not Found",
				fmt.Sprintf("Image %q not found", config.Image.ValueString()),
			)
		default:
			resp.Diagnostics.AddError(
				"Image Not Found",
				fmt.Sprintf("Image %q with tag %q not found", config.Image.ValueString(), config.Tag.ValueString()),
			)
		}
		return
	}

	latestImg := slices.MaxFunc(objs.Items, func(a, b containerv1.Image) int {
		return a.CreatedTimestamp.Compare(b.CreatedTimestamp)
	})

	state := newImageModel(latestImg)
	state.Branch = config.Branch
	state.Image = config.Image
	state.Tag = config.Tag
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
