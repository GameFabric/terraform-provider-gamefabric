package rbac_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	rbacv1 "github.com/gamefabric/gf-core/pkg/api/rbac/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestGroups(t *testing.T) {
	t.Parallel()

	grp := &rbacv1.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-group",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, grp)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_groups" "test1" {
  label_filter = {
	foo = "bar"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_groups.test1", "label_filter.foo", "bar"),
					resource.TestCheckResourceAttr("data.gamefabric_groups.test1", "names.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_groups.test1", "names.0", "test-group"),
				),
			},
		},
	})
}

func TestGroups_AllowsGettingAll(t *testing.T) {
	t.Parallel()

	grp1 := &rbacv1.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-group1",
			Labels: map[string]string{
				"foo": "bar",
			},
		},
	}
	grp2 := &rbacv1.Group{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-group2",
			Labels: map[string]string{
				"foo": "bat",
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, grp1, grp2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_groups" "test1" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_groups.test1", "names.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_groups.test1", "names.0", "test-group1"),
					resource.TestCheckResourceAttr("data.gamefabric_groups.test1", "names.1", "test-group2"),
				),
			},
		},
	})
}
