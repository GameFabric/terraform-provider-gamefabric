package authentication_test

import (
	"regexp"
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
			Name:   "user",
			Labels: map[string]string{"env": "test", "team": "devops"},
		},
		Spec: authv1.ServiceAccountSpec{
			Username: "user",
			Email:    "user@example.com",
		},
		Status: authv1.ServiceAccountStatus{
			State:    authv1.ServiceAccountStateApplied,
			Password: "s3cr3t",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, sa)

	checkServiceAccount := func(resourceName string) resource.TestCheckFunc {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(resourceName, "name", "user"),
			resource.TestCheckResourceAttr(resourceName, "email", "user@example.com"),
			resource.TestCheckResourceAttr(resourceName, "labels.env", "test"),
			resource.TestCheckResourceAttr(resourceName, "labels.team", "devops"),
		)
	}

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_service_account" "this" {
  name = "user"
}
`,
				Check: checkServiceAccount("data.gamefabric_service_account.this"),
			},
			{
				Config: `data "gamefabric_service_account" "this" {
  email = "user@example.com"
}
`,
				Check: checkServiceAccount("data.gamefabric_service_account.this"),
			},
			{
				Config: `data "gamefabric_service_account" "this" {
  name = "nonexistent"
}
`,
				ExpectError: regexp.MustCompile(`Failed to retrieve ServiceAccount`),
			},
		},
	})
}
