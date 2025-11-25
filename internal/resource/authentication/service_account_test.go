package authentication_test

import (
	"fmt"
	"testing"

	"github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestServiceAccountResource(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCheckServiceAccountDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_service_account" "test" {
  name = "svc-test"
  labels = {
    env = "test"
    team = "devops"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_service_account.test", "name", "svc-test"),
					resource.TestCheckResourceAttr("gamefabric_service_account.test", "labels.env", "test"),
					resource.TestCheckResourceAttr("gamefabric_service_account.test", "labels.team", "devops"),
					resource.TestCheckResourceAttr("gamefabric_service_account.test", "email", "svc-test@ec.nitrado.systems"),
					resource.TestCheckResourceAttr("gamefabric_service_account.test", "password", "some-password"),
				),
			},
		},
	})
}

func TestServiceAccountResource_SkipUpdateWithNoChange(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCheckServiceAccountDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_service_account" "test" { name = "svc-test" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_service_account.test", "name", "svc-test"),
					resource.TestCheckResourceAttr("gamefabric_service_account.test", "email", "svc-test@ec.nitrado.systems"),
					resource.TestCheckResourceAttr("gamefabric_service_account.test", "password", "some-password"),
				),
			},
			{
				ResourceName: "gamefabric_service_account.test",
				ImportState:  true,
			},
			{
				Config: `resource "gamefabric_service_account" "test" { name = "svc-test" }`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_service_account.test", "name", "svc-test"),
				),
			},
		},
	})
}

func testCheckServiceAccountDestroy(t *testing.T, cs clientset.Interface) func(*terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_service_account" {
				continue
			}
			_, err := cs.AuthenticationV1Beta1().ServiceAccounts().Get(
				t.Context(),
				rs.Primary.Attributes["name"],
				v1.GetOptions{},
			)
			if err == nil {
				return fmt.Errorf("Service account %s still exists", rs.Primary.Attributes["name"])
			}
		}
		return nil
	}
}
