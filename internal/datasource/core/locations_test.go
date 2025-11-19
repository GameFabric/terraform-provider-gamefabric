package core_test

import (
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestLocations(t *testing.T) {
	t.Parallel()

	loc1 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc-1",
			Annotations: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "amsterdam",
				"g8c.io/country":       "netherlands",
				"g8c.io/continent":     "europe",
			},
		},
	}
	loc2 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc2",
			Annotations: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "amsterdam",
				"g8c.io/country":       "netherlands",
				"g8c.io/continent":     "europe",
			},
		},
	}
	loc3 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc-3",
			Annotations: map[string]string{
				"g8c.io/provider-type": "something",
				"g8c.io/city":          "amsterdam",
				"g8c.io/country":       "netherlands",
				"g8c.io/continent":     "europe",
			},
		},
	}
	loc4 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc-4",
			Annotations: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "something",
				"g8c.io/country":       "netherlands",
				"g8c.io/continent":     "europe",
			},
		},
	}
	loc5 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc-5",
			Annotations: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "amsterdam",
				"g8c.io/country":       "something",
				"g8c.io/continent":     "europe",
			},
		},
	}
	loc6 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "loc-6",
			Annotations: map[string]string{
				"g8c.io/provider-type": "baremetal",
				"g8c.io/city":          "amsterdam",
				"g8c.io/country":       "netherlands",
				"g8c.io/continent":     "something",
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, loc1, loc2, loc3, loc4, loc5, loc6)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_locations" "test1" {
  type = "Baremetal"
  city = "Amsterdam"
  country = "Netherlands"
  continent = "Europe"
  name_regex = "loc-.*"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "type", "Baremetal"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "city", "Amsterdam"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "country", "Netherlands"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "continent", "Europe"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "name_regex", "loc-.*"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.0", "loc-1"),
				),
			},
		},
	})
}

func TestLocations_LabelFilterByCountries(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t, testAmsterdamAzure, testAmsterdamBm, testFrankfurtAws)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_locations" "test1" {
  label_filter = {
    "g8c.io/country" = ["DE", "nl"] // Include case insensitivity in this test.
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.%", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.g8c.io/country.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.g8c.io/country.0", "DE"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.g8c.io/country.1", "nl"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.#", "3"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.0", "amsterdam-azure"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.1", "amsterdam-baremetal"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.2", "frankfurt-aws"),
				),
			},
		},
	})
}

func TestLocations_LabelFilterByCountry(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t, testAmsterdamAzure, testAmsterdamBm, testFrankfurtAws)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_locations" "test1" {
  label_filter = {
    "g8c.io/country" = "NL" // Include case insensitivity in this test.
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.%", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.g8c.io/country", "NL"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.0", "amsterdam-azure"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.1", "amsterdam-baremetal"),
				),
			},
		},
	})
}

func TestLocations_NoLabelFilter(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t, testAmsterdamAzure, testAmsterdamBm, testFrankfurtAws)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_locations" "test1" {
  label_filter = {}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.%", "0"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.#", "3"),
				),
			},
		},
	})
}

func TestLocations_NoLabelFilterNoLabels(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t, &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	})

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_locations" "test1" {
  label_filter = {}
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.%", "0"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.#", "1"),
				),
			},
		},
	})
}

func TestLocations_LabelFilterButNoLabels(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t, &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	})

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_locations" "test1" {
  label_filter = {
    "g8c.io/country" = "de"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.%", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.#", "0"),
				),
			},
		},
	})
}

// TestLocations_LabelFilterDocumentBehavior documents current behavior for certain inputs.
//
// It may change in the future, but it should be an informed decision.
func TestLocations_LabelFilterDocumentBehavior(t *testing.T) {
	t.Parallel()

	loc1 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-1",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
	}
	loc2 := &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-2",
			Labels: map[string]string{
				"bar": "foo",
			},
		},
	}

	t.Run("handles label filter key with empty values", func(t *testing.T) {
		pf, _ := providertest.ProtoV6ProviderFactories(t, loc1, loc2)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: pf,
			Steps: []resource.TestStep{
				{
					Config: `data "gamefabric_locations" "test1" {
  label_filter = {
    "foo" = [] // This is ignored.
  }
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.%", "1"),
						resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.#", "2"),
					),
				},
			},
		})
	})

	// This is to document behavior only and may change in the future.
	t.Run("handles label filter key with empty string", func(t *testing.T) {
		pf, _ := providertest.ProtoV6ProviderFactories(t, loc1, loc2)

		resource.Test(t, resource.TestCase{
			IsUnitTest:               true,
			ProtoV6ProviderFactories: pf,
			Steps: []resource.TestStep{
				{
					Config: `data "gamefabric_locations" "test1" {
  label_filter = {
    "foo" = "" // This is ignored.
  }
}
`,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.%", "1"),
						resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.#", "2"),
					),
				},
			},
		})
	})
}

func TestLocations_LabelFilterWithNoMatch(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t, testAmsterdamAzure, testAmsterdamBm, testFrankfurtAws)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_locations" "test1" {
  label_filter = {
    "g8c.io/country" = "nope"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.%", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "label_filter.g8c.io/country", "nope"),
					resource.TestCheckResourceAttr("data.gamefabric_locations.test1", "names.#", "0"),
				),
			},
		},
	})
}

func TestLocations_LabelFilterInvalid(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_locations" "test1" {
  label_filter = 123
}
`,
				ExpectError: regexp.MustCompile("Error Applying Label Filter"),
			},
		},
	})
}

func TestLocations_LabelFilterInvalidElement(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_locations" "test1" {
  label_filter = {
    "g8c.io/country" = 123
  }
}
`,
				ExpectError: regexp.MustCompile("Error Applying Label Filter"),
			},
		},
	})
}

var (
	testAmsterdamAzure = &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amsterdam-azure",
			Labels: map[string]string{
				"g8c.io/city":          "amsterdam",
				"g8c.io/continent":     "eu",
				"g8c.io/country":       "nl",
				"g8c.io/provider-type": "cloud",
			},
			Annotations: map[string]string{
				"g8c.io/city":            "Amsterdam",
				"g8c.io/continent":       "Europe",
				"g8c.io/country":         "Netherlands",
				"g8c.io/provider-type":   "Cloud",
				"g8c.io/system-location": "true",
			},
		},
	}
	testAmsterdamBm = &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "amsterdam-baremetal",
			Labels: map[string]string{
				"g8c.io/city":          "amsterdam",
				"g8c.io/continent":     "eu",
				"g8c.io/country":       "nl",
				"g8c.io/provider-type": "baremetal",
			},
			Annotations: map[string]string{
				"g8c.io/city":            "Amsterdam",
				"g8c.io/continent":       "Europe",
				"g8c.io/country":         "Netherlands",
				"g8c.io/provider-type":   "Baremetal",
				"g8c.io/system-location": "true",
			},
		},
	}
	testFrankfurtAws = &corev1.Location{
		ObjectMeta: metav1.ObjectMeta{
			Name: "frankfurt-aws",
			Labels: map[string]string{
				"g8c.io/city":          "frankfurt",
				"g8c.io/continent":     "eu",
				"g8c.io/country":       "de",
				"g8c.io/provider-type": "cloud",
			},
			Annotations: map[string]string{
				"g8c.io/city":            "Frankfurt",
				"g8c.io/continent":       "Europe",
				"g8c.io/country":         "Germany",
				"g8c.io/provider-type":   "Cloud",
				"g8c.io/system-location": "true",
			},
		},
	}
)
