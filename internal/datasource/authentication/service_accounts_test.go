package authentication_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	authv1 "github.com/gamefabric/gf-core/pkg/api/authentication/v1beta1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestServiceAccounts(t *testing.T) {
	t.Parallel()

	sa1 := authv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "user1",
			Labels: map[string]string{"env": "test", "team": "devops"},
		},
		Spec: authv1.ServiceAccountSpec{
			Username: "user1",
			Email:    "user1@example.com",
		},
		Status: authv1.ServiceAccountStatus{
			State:    authv1.ServiceAccountStateApplied,
			Password: "pw1",
		},
	}
	sa2 := authv1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "user2",
			Labels: map[string]string{"env": "prod", "team": "devops"},
		},
		Spec: authv1.ServiceAccountSpec{
			Username: "user2",
			Email:    "user2@example.com",
		},
		Status: authv1.ServiceAccountStatus{
			State:    authv1.ServiceAccountStatePending,
			Password: "pw2",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, &sa1, &sa2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_service_accounts" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_service_accounts.all", "items.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_service_accounts.all", "items.0.name", "user1"),
					resource.TestCheckResourceAttr("data.gamefabric_service_accounts.all", "items.1.name", "user2"),
					resource.TestCheckResourceAttr("data.gamefabric_service_accounts.all", "items.0.email", "user1@example.com"),
					resource.TestCheckResourceAttr("data.gamefabric_service_accounts.all", "items.1.email", "user2@example.com"),
				),
			},
			{
				Config: `data "gamefabric_service_accounts" "filtered" {
  label_filters = {
    env = "test"
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_service_accounts.filtered", "items.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_service_accounts.filtered", "items.0.name", "sa-1"),
					resource.TestCheckResourceAttr("data.gamefabric_service_accounts.filtered", "items.0.labels.env", "test"),
					resource.TestCheckResourceAttr("data.gamefabric_service_accounts.filtered", "items.0.email", "user1@example.com"),
				),
			},
		},
	})
}
