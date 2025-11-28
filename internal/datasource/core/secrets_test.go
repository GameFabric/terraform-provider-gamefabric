package core_test

import (
	"testing"
	"time"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"k8s.io/utils/ptr"
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
		Status: corev1.SecretStatus{
			State:          corev1.SecretStateSynced,
			LastDataChange: ptr.To(time.Now()),
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
		Status: corev1.SecretStatus{
			State: corev1.SecretStateSynced,
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
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.status", "Synced"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.name", "secret-2"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.description", "Second test secret"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.status", "Synced"),
				),
			},
		},
	})
}

func TestSecrets_WithAllLabels(t *testing.T) {
	t.Parallel()

	secret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "secret-prod-1",
			Environment: "dflt",
			Labels: map[string]string{
				"team": "platform",
				"env":  "prod",
			},
		},
		Description: "Production secret",
		Data: map[string]string{
			"password": "prod-value",
		},
		Status: corev1.SecretStatus{
			State: corev1.SecretStateSynced,
		},
	}

	secret2 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "secret-staging-1",
			Environment: "dflt",
			Labels: map[string]string{
				"team": "platform",
				"env":  "staging",
			},
		},
		Description: "Staging secret",
		Data: map[string]string{
			"password": "staging-value",
		},
		Status: corev1.SecretStatus{
			State: corev1.SecretStateSynced,
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
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.name", "secret-prod-1"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.description", "Production secret"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.status", "Synced"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.name", "secret-staging-1"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.description", "Staging secret"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.status", "Synced"),
				),
			},
		},
	})
}

func TestSecrets_WithDescription(t *testing.T) {
	t.Parallel()

	secret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "secret-with-desc",
			Environment: "dflt",
			Labels: map[string]string{
				"team": "platform",
			},
		},
		Description: "Production database password with audit trail",
		Data: map[string]string{
			"password": "secret-value",
		},
		Status: corev1.SecretStatus{
			State: corev1.SecretStateSynced,
		},
	}

	secret2 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "secret-no-desc",
			Environment: "dflt",
			Labels: map[string]string{
				"team": "platform",
			},
		},
		Data: map[string]string{
			"key": "value",
		},
		Status: corev1.SecretStatus{
			State: corev1.SecretStateSynced,
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
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.name", "secret-no-desc"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.name", "secret-with-desc"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.description", "Production database password with audit trail"),
				),
			},
		},
	})
}

func TestSecrets_SortedByName(t *testing.T) {
	t.Parallel()

	secretZ := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "z-secret",
			Environment: "dflt",
		},
		Data: map[string]string{
			"key": "value-z",
		},
		Status: corev1.SecretStatus{
			State: corev1.SecretStateSynced,
		},
	}

	secretA := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "a-secret",
			Environment: "dflt",
		},
		Data: map[string]string{
			"key": "value-a",
		},
		Status: corev1.SecretStatus{
			State: corev1.SecretStateSynced,
		},
	}

	secretM := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "m-secret",
			Environment: "dflt",
		},
		Data: map[string]string{
			"key": "value-m",
		},
		Status: corev1.SecretStatus{
			State: corev1.SecretStateSynced,
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, secretZ, secretA, secretM)

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
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.#", "3"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.0.name", "a-secret"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.1.name", "m-secret"),
					resource.TestCheckResourceAttr("data.gamefabric_secrets.test1", "secrets.2.name", "z-secret"),
				),
			},
		},
	})
}
