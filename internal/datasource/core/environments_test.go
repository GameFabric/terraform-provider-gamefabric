package core_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestEnvironments(t *testing.T) {
	t.Parallel()

	env := &corev1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dflt",
			Labels: map[string]string{
				"env": "prod",
			},
		},
		Spec: corev1.EnvironmentSpec{
			DisplayName: "My Env",
			Description: "My Env Description",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, env)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_environments" "test1" {
  label_filter = {
	env = "prod"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_environments.test1", "label_filter.env", "prod"),
					resource.TestCheckResourceAttr("data.gamefabric_environments.test1", "environments.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_environments.test1", "environments.0.name", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_environments.test1", "environments.0.labels.env", "prod"),
					resource.TestCheckResourceAttr("data.gamefabric_environments.test1", "environments.0.display_name", "My Env"),
					resource.TestCheckResourceAttr("data.gamefabric_environments.test1", "environments.0.description", "My Env Description"),
				),
			},
		},
	})
}

func TestEnvironments_AllowsGettingAll(t *testing.T) {
	t.Parallel()

	env := &corev1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dflt",
		},
		Spec: corev1.EnvironmentSpec{
			DisplayName: "My Env",
			Description: "My Env Description",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, env)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_environments" "test1" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_environments.test1", "environments.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_environments.test1", "environments.0.name", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_environments.test1", "environments.0.display_name", "My Env"),
					resource.TestCheckResourceAttr("data.gamefabric_environments.test1", "environments.0.description", "My Env Description"),
				),
			},
		},
	})
}
