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
	"github.com/hashicorp/terraform-plugin-framework/path"
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
			"label_filter": schema.DynamicAttribute{
				Description:         "A map of keys and values that is used to filter locations. Only items with all specified labels (exact matches) will be returned. Can be combined with other filters. Each map key can have a value that is either a string or an array of strings.",
				MarkdownDescription: "A map of keys and values that is used to filter locations. Only items with all specified labels (exact matches) will be returned. Can be combined with other filters. Each map key can have a value that is either a string or an array of strings.",
				Optional:            true,
			},
			"names": schema.ListAttribute{
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

//nolint:cyclop // Splitting this would not make it much simpler.
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

	// Once supported we can replace this with the LabelSelector.
	matchLabels, err := parseLabelFilter(config.LabelFilter, path.Root("label_filter"))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Applying Label Filter",
			err.Error(),
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
		if !matchLabels(loc.GetLabels()) {
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
	state.LabelFilter = config.LabelFilter
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type matchLabelsFunc func(map[string]string) bool

var (
	rejectLabels = func(map[string]string) bool { return false }
	acceptLabels = func(map[string]string) bool { return true }
)

func parseLabelFilter(filter types.Dynamic, path path.Path) (matchLabelsFunc, error) {
	if filter.IsUnknown() {
		return rejectLabels, nil
	}

	if filter.IsNull() {
		// No filter defined.
		return acceptLabels, nil
	}

	wantLabels, err := toLabelFilter(filter, path)
	if err != nil {
		return rejectLabels, err
	}

	return func(labels map[string]string) bool {
		if len(wantLabels) == 0 {
			return true
		}
		if len(labels) == 0 {
			return false
		}

		for key, values := range wantLabels {
			if len(values) == 0 {
				continue // Ignore.
			}

			var match bool
			for _, val := range values {
				match = match || strings.EqualFold(val, labels[key])
			}
			if !match {
				return false
			}
		}
		return true
	}, nil
}

func toLabelFilter(filter types.Dynamic, path path.Path) (map[string][]string, error) {
	val := filter.UnderlyingValue()
	obj, ok := val.(types.Object)
	if !ok {
		return nil, fmt.Errorf("%s must be basetypes.Object (map), got: %T", path.String(), val)
	}

	res := make(map[string][]string, len(obj.Attributes()))
	for key, attr := range obj.Attributes() {
		if !conv.IsKnown(attr) {
			continue
		}

		switch typ := attr.(type) {
		case types.String:
			if typ.ValueString() == "" {
				continue
			}

			res[key] = append(res[key], typ.ValueString())
		case types.Tuple:
			for i, elem := range typ.Elements() {
				if !conv.IsKnown(elem) {
					continue
				}

				str, ok := elem.(types.String)
				if !ok {
					return nil, fmt.Errorf("%s must be basetypes.String, got: %T", path.AtMapKey(key).AtListIndex(i).String(), elem)
				}

				if str.ValueString() == "" {
					continue
				}
				res[key] = append(res[key], str.ValueString())
			}
		default:
			return nil, fmt.Errorf("%s must be basetypes.String or basetypes.Tuple (array), got: %T", path.AtMapKey(key).String(), attr)
		}
	}

	return res, nil
}
