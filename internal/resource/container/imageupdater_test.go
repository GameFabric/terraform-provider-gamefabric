package container_test

import (
	"fmt"
	"strings"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestImageUpdater(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceImageUpdaterDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceImageUpdaterConfigBasic("prod"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_imageupdater.test", "branch", "prod"),
					resource.TestCheckResourceAttr("gamefabric_imageupdater.test", "image", "my-image"),
					resource.TestCheckResourceAttr("gamefabric_imageupdater.test", "target.type", "armada"),
					resource.TestCheckResourceAttr("gamefabric_imageupdater.test", "target.name", "my-armada"),
					resource.TestCheckResourceAttr("gamefabric_imageupdater.test", "target.environment", "dflt"),
				),
			},
			{
				Config: testResourceImageUpdaterConfigBasic("dev"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_imageupdater.test", "branch", "dev"),
					resource.TestCheckResourceAttr("gamefabric_imageupdater.test", "image", "my-image"),
					resource.TestCheckResourceAttr("gamefabric_imageupdater.test", "target.type", "armada"),
					resource.TestCheckResourceAttr("gamefabric_imageupdater.test", "target.name", "my-armada"),
					resource.TestCheckResourceAttr("gamefabric_imageupdater.test", "target.environment", "dflt"),
				),
			},
		},
	})
}

func testResourceImageUpdaterConfigBasic(branch string) string {
	return fmt.Sprintf(`resource "gamefabric_imageupdater" "test" {
  branch = "%s"
  image = "my-image"
  target = {
    type = "armada"
    name = "my-armada"
	environment = "dflt"
  }
}`, branch)
}

func testResourceImageUpdaterDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_imageupdater" {
				continue
			}

			env, name, _ := strings.Cut(rs.Primary.ID, "/")
			resp, err := cs.ContainerV1().ImageUpdaters(env).Get(t.Context(), name, metav1.GetOptions{})
			if err == nil {
				if resp.Name == name {
					return fmt.Errorf("image updater still exists: %s", rs.Primary.ID)
				}
			}
		}
		return nil
	}
}
