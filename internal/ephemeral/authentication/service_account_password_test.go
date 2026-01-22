package authentication_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	authv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServiceAccountPasswordEphemeral(t *testing.T) {
	t.Parallel()

	// Pre-create a service account and environment that will be used for the password reset
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
				Config: `
ephemeral "gamefabric_service_account_password" "test" {
  service_account = "svc-test"
}

# Use the ephemeral password in a write-only secret to verify it works
resource "gamefabric_secret" "test" {
  name        = "test-secret"
  environment = "test"
  data_wo = {
    password = ephemeral.gamefabric_service_account_password.test.password
  }
  data_wo_version = 1
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the secret was created, which proves the ephemeral resource provided the password
					resource.TestCheckResourceAttr("gamefabric_secret.test", "name", "test-secret"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "environment", "test"),
				),
			},
		},
	})
}
