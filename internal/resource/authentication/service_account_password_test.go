package authentication_test

import (
	"testing"

	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServiceAccountPasswordResource(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: `resource "gamefabric_service_account_password" "test" {
  service_account = "svc-test"
  labels = {
    env = "test"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("gamefabric_service_account_password.test", "id"),
					resource.TestCheckResourceAttr("gamefabric_service_account_password.test", "service_account", "svc-test"),
					resource.TestCheckResourceAttrSet("gamefabric_service_account_password.test", "password_wo"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "gamefabric_service_account_password.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
