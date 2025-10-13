package core

import (
	"context"
	"fmt"
	"regexp"
	"slices"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/conv"
	provcontext "github.com/gamefabric/terraform-provider-gamefabric/internal/provider/context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &locations{}
	_ datasource.DataSourceWithConfigure = &locations{}
)

type locations struct {
	clientSet clientset.Interface
}

// NewLocations returns a new instance of the locations data source.
func NewLocations() datasource.DataSource {
	return &locations{}
}

func (r *locations) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locations"
}

// Schema defines the schema for this data source.
func (r *locations) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description:         "The type used to filter locations.",
				MarkdownDescription: "The type used to filter locations.",
				Optional:            true,
			},
			"city": schema.StringAttribute{
				Description:         "The city used to filter locations.",
				MarkdownDescription: "The city used to filter locations.",
				Optional:            true,
			},
			"continent": schema.StringAttribute{
				Description:         "The continent used to filter locations.",
				MarkdownDescription: "The continent used to filter locations.",
				Optional:            true,
			},
			"name_regex": schema.StringAttribute{
				Description:         "The regex used to filter locations by their name.",
				MarkdownDescription: "The regex used to filter locations by their name.",
				Optional:            true,
			},
			"locations": schema.ListAttribute{
				Description:         "The locations matching the filters.",
				MarkdownDescription: "The locations matching the filters.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure prepares the struct.
func (r *locations) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (r *locations) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config locationsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lbls := map[string]string{}
	if conv.IsKnown(config.Type) {
		lbls["g8c.io/provider-type"] = config.Type.ValueString()
	}
	if conv.IsKnown(config.City) {
		lbls["g8c.io/city"] = config.City.ValueString()
	}
	if conv.IsKnown(config.Continent) {
		lbls["g8c.io/continent"] = config.Continent.ValueString()
	}

	var (
		nameRegexp   *regexp.Regexp
		hasNameRegex bool
	)
	if conv.IsKnown(config.NameRegex) {
		reg, err := regexp.Compile(config.NameRegex.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Compiling Name Regex",
				fmt.Sprintf("Could not compile name regex %q: %v", config.NameRegex.ValueString(), err),
			)
			return
		}
		nameRegexp = reg
		hasNameRegex = true
	}

	list, err := r.clientSet.CoreV1().Locations().List(ctx, metav1.ListOptions{LabelSelector: lbls})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Environments",
			fmt.Sprintf("Could not get Environments: %v", err),
		)
		return
	}

	locs := make([]string, 0, len(list.Items))
	for _, loc := range list.Items {
		if hasNameRegex && !nameRegexp.MatchString(loc.Name) {
			continue
		}
		locs = append(locs, loc.Name)
	}
	slices.Sort(locs)

	state := newLocationsModel(locs)
	state.Type = config.Type
	state.City = config.City
	state.Continent = config.Continent
	state.NameRegex = config.NameRegex
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
