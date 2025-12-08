package core_test

import (
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestSecret(t *testing.T) {
	t.Parallel()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-secret",
			Environment: "dflt",
			Labels: map[string]string{
				"team": "platform",
			},
		},
		Description: "Test secret",
		Data: map[string]string{
			"password":   "secret-value",
			"second-key": "another-secret-value",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, secret)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_secret" "test1" {
  name        = "test-secret"
  environment = "dflt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_secret.test1", "name", "test-secret"),
					resource.TestCheckResourceAttr("data.gamefabric_secret.test1", "environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_secret.test1", "description", "Test secret"),
					resource.TestCheckResourceAttr("data.gamefabric_secret.test1", "labels.%", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_secret.test1", "labels.team", "platform"),
					resource.TestCheckResourceAttr("data.gamefabric_secret.test1", "data.0", "password"),
					resource.TestCheckResourceAttr("data.gamefabric_secret.test1", "data.1", "second-key"),
				),
			},
		},
	})
}

func TestSecret_NotFound(t *testing.T) {
	t.Parallel()

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-secret",
			Environment: "dflt",
		},
		Description: "Test secret",
		Data: map[string]string{
			"password": "secret-value",
		},
		Status: corev1.SecretStatus{
			State: corev1.SecretStateSynced,
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, secret)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_secret" "test1" {
 name        = "nonexistent-secret"
 environment = "dflt"
}
`,
				ExpectError: regexp.MustCompile(`Secret Not Found`),
			},
		},
	})
}
