package core_test

import (
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestEnvironment(t *testing.T) {
	t.Parallel()

	env := &corev1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "dflt"},
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
				Config: `data "gamefabric_environment" "test1" {
  name = "dflt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_environment.test1", "name", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_environment.test1", "display_name", "My Env"),
					resource.TestCheckResourceAttr("data.gamefabric_environment.test1", "description", "My Env Description"),
				),
			},
			{
				Config: `data "gamefabric_environment" "test2" {
  display_name = "My Env"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_environment.test2", "name", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_environment.test2", "display_name", "My Env"),
					resource.TestCheckResourceAttr("data.gamefabric_environment.test2", "description", "My Env Description"),
				),
			},
		},
	})
}

func TestEnvironment_HandlesMultipleMatches(t *testing.T) {
	t.Parallel()

	env := &corev1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "dflt"},
		Spec: corev1.EnvironmentSpec{
			DisplayName: "My Env",
			Description: "My Env Description",
		},
	}
	otherEnv := &corev1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "other"},
		Spec: corev1.EnvironmentSpec{
			DisplayName: "My Env",
			Description: "My Env Description",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, env, otherEnv)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_environment" "test" {
  display_name = "My Env"
}
`,
				ExpectError: regexp.MustCompile(`Multiple Environments Found`),
			},
		},
	})
}

func TestEnvironment_HandlesMultipleSelectors(t *testing.T) {
	t.Parallel()

	env := &corev1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "dflt"},
		Spec: corev1.EnvironmentSpec{
			DisplayName: "My Env",
			Description: "My Env Description",
		},
	}
	otherEnv := &corev1.Environment{
		ObjectMeta: metav1.ObjectMeta{Name: "other"},
		Spec: corev1.EnvironmentSpec{
			DisplayName: "My Env",
			Description: "My Env Description",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, env, otherEnv)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_environment" "test" {
  name = "dflt"
  display_name = "My Env"
}
`,
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}
