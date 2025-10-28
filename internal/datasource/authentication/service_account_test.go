package authentication_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	authv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServiceAccount(t *testing.T) {
	t.Parallel()

	sa := &authv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-sa",
			Labels: map[string]string{"env": "test", "team": "devops"},
		},
		Spec: authv1.ServiceAccountSpec{
			Username: "svc-user",
			Email:    "svc@example.com",
		},
		Status: authv1.ServiceAccountStatus{
			State:    authv1.ServiceAccountStateApplied,
			Password: "s3cr3t",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, sa)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_service_account" "test-sa" {
  name = "test-sa"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_service_account.test-sa", "name", "test-sa"),
					resource.TestCheckResourceAttr("data.gamefabric_service_account.test-sa", "username", "svc-user"),
					resource.TestCheckResourceAttr("data.gamefabric_service_account.test-sa", "email", "svc@example.com"),
					resource.TestCheckResourceAttr("data.gamefabric_service_account.test-sa", "state", "Applied"),
					resource.TestCheckResourceAttr("data.gamefabric_service_account.test-sa", "labels.env", "test"),
					resource.TestCheckResourceAttr("data.gamefabric_service_account.test-sa", "labels.team", "devops"),
				),
			},
		},
	})
}
