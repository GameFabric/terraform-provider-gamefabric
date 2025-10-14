package core

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

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
			"country": schema.StringAttribute{
				Description:         "The country used to filter locations.",
				MarkdownDescription: "The country used to filter locations.",
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

	typ, hasTyp := config.Type.ValueString(), conv.IsKnown(config.Type)
	city, hasCity := config.City.ValueString(), conv.IsKnown(config.City)
	country, hasCountry := config.Country.ValueString(), conv.IsKnown(config.Country)
	continent, hasContinent := config.Continent.ValueString(), conv.IsKnown(config.Continent)

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

	list, err := r.clientSet.CoreV1().Locations().List(ctx, metav1.ListOptions{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Locations",
			fmt.Sprintf("Could not get Locations: %v", err),
		)
		return
	}

	locs := make([]string, 0, len(list.Items))
	for _, loc := range list.Items {
		annos := loc.GetAnnotations()
		if annos == nil {
			annos = map[string]string{}
		}

		if hasTyp && !strings.EqualFold(annos["g8c.io/provider-type"], typ) {
			continue
		}
		if hasCity && !strings.EqualFold(annos["g8c.io/city"], city) {
			continue
		}
		if hasCountry && !strings.EqualFold(annos["g8c.io/country"], country) {
			continue
		}
		if hasContinent && !strings.EqualFold(annos["g8c.io/continent"], continent) {
			continue
		}
		if hasNameRegex && !nameRegexp.MatchString(loc.Name) {
			continue
		}
		locs = append(locs, loc.Name)
	}
	slices.Sort(locs)

	state := newLocationsModel(locs)
	state.Type = config.Type
	state.City = config.City
	state.Country = config.Country
	state.Continent = config.Continent
	state.NameRegex = config.NameRegex
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
