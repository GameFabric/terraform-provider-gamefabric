package core_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSecrets(t *testing.T) {
	t.Parallel()

	secret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "secret-1",
			Environment: "dflt",
			Labels: map[string]string{
				"team": "platform",
				"env":  "prod",
			},
		},
		Description: "First test secret",
		Data: map[string]string{
			"password": "value1",
		},
	}

	secret2 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "secret-2",
			Environment: "dflt",
			Labels: map[string]string{
				"team": "platform",
				"env":  "staging",
			},
		},
		Description: "Second test secret",
		Data: map[string]string{
			"api-key": "value2",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, secret1, secret2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_secrets" "test1" {
  environment = "dflt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.name", "secret-1"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.description", "First test secret"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.labels.env", "prod"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.labels.team", "platform"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.data.0", "password"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.name", "secret-2"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.description", "Second test secret"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.labels.env", "staging"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.labels.team", "platform"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.data.0", "api-key"),
				),
			},
		},
	})
}
