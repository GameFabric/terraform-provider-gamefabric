package core_test

import (
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	kresource "k8s.io/apimachinery/pkg/api/resource"
)

func TestRegion(t *testing.T) {
	region := &corev1.Region{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-region",
			Environment: "dflt",
		},
		Spec: corev1.RegionSpec{
			DisplayName: "My Region",
			Types: []corev1.RegionType{
				{Name: "small"},
			},
		},
		Status: corev1.RegionStatus{
			Types: []corev1.RegionTypeStatus{
				{
					Name: "small",
					Limit: corev1.Resources{
						CPU:    kresource.MustParse("2"),
						Memory: kresource.MustParse("4Gi"),
					},
				},
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, region)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_region" "test1" {
  name = "test-region"
  environment = "dflt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_region.test1", "name", "test-region"),
					resource.TestCheckResourceAttr("data.gamefabric_region.test1", "display_name", "My Region"),
					resource.TestCheckResourceAttr("data.gamefabric_region.test1", "types.%", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_region.test1", "types.small.cpu", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_region.test1", "types.small.memory", "4Gi"),
				),
			},
			{
				Config: `data "gamefabric_region" "test2" {
  display_name = "My Region"
  environment = "dflt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_region.test2", "name", "test-region"),
					resource.TestCheckResourceAttr("data.gamefabric_region.test2", "display_name", "My Region"),
					resource.TestCheckResourceAttr("data.gamefabric_region.test2", "types.%", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_region.test2", "types.small.cpu", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_region.test2", "types.small.memory", "4Gi"),
				),
			},
		},
	})
}

func TestRegion_HandlesMultipleMatches(t *testing.T) {
	region := &corev1.Region{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-region",
			Environment: "dflt",
		},
		Spec: corev1.RegionSpec{
			DisplayName: "My Region",
		},
	}
	otherRegion := &corev1.Region{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "other-region",
			Environment: "dflt",
		},
		Spec: corev1.RegionSpec{
			DisplayName: "My Region",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, region, otherRegion)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_region" "test" {
  display_name = "My Region"
  environment = "dflt"
}
`,
				ExpectError: regexp.MustCompile(`Multiple Regions Found`),
			},
		},
	})
}

func TestRegion_HandlesMultipleSelectors(t *testing.T) {
	region := &corev1.Region{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-region",
			Environment: "dflt",
		},
		Spec: corev1.RegionSpec{
			DisplayName: "My Region",
		},
	}
	otherRegion := &corev1.Region{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "other-region",
			Environment: "dflt",
		},
		Spec: corev1.RegionSpec{
			DisplayName: "My Region",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, region, otherRegion)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_region" "test" {
  name = "dflt"
  environment = "dflt" 
  display_name = "My Env"
}
`,
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}
