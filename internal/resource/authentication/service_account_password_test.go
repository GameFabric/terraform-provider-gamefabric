package authentication_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	authv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServiceAccountPasswordResource(t *testing.T) {
	t.Parallel()

	// Pre-create a service account that will be used for the password reset
	serviceAccount := &authv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: "svc-test",
		},
		Spec: authv1.ServiceAccountSpec{
			Username: "svc-test",
			Email:    "svc-test@ec.nitrado.systems",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, serviceAccount)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_service_account_password" "test" {
  service_account = "svc-test"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("gamefabric_service_account_password.test", "id"),
					resource.TestCheckResourceAttr("gamefabric_service_account_password.test", "service_account", "svc-test"),
					resource.TestCheckResourceAttr("gamefabric_service_account_password.test", "password", "some-reset-password"),
				),
			},
			{
				Config: `resource "gamefabric_service_account_password" "test" {
  service_account = "svc-test"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("gamefabric_service_account_password.test", "id"),
					resource.TestCheckResourceAttr("gamefabric_service_account_password.test", "service_account", "svc-test"),
					resource.TestCheckResourceAttr("gamefabric_service_account_password.test", "password", "some-reset-password"),
				),
			},
		},
	})
}
