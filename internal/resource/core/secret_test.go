package core_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestSecretResource(t *testing.T) {
	t.Parallel()

	name := "db-creds"
	nameWO := "db-creds-wo"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceSecretDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceSecretConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_secret.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "environment", "dflt"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "description", "Backend database credentials"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "labels.game", "my_first_game"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data.db_user", "dbuser123"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data.db_password", "super-secret-pass-123"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data.%", "2"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data_wo.%", "0"),
				),
			},
			{
				Config: testResourceSecretConfigUpdateData(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_secret.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "environment", "dflt"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "description", "Backend database credentials - changed"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "labels.game", "my_first_game-1"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data.db_user", "dbuser123"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data.db_password", "super-secret-pass-123"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data.one_more", "secret-value"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data.%", "3"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data_wo.%", "0"),
				),
			},
			{
				Config: testResourceSecretConfigDataWO(nameWO),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_secret.test", "name", nameWO),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "environment", "dflt"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "description", "Backend database credentials"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data.%", "0"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data_wo.%", "0"),
				),
			},
			{
				Config: testResourceSecretConfigUpdateDataWO(nameWO),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_secret.test", "name", nameWO),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "environment", "dflt"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "description", "Backend database credentials"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data.%", "0"),
					resource.TestCheckResourceAttr("gamefabric_secret.test", "data_wo.%", "0"),
				),
			},
		},
	})
}

// TestSecretResourceBothData verifies that having both data and data_wo triggers an error
func TestSecretResourceBothData(t *testing.T) {
	t.Parallel()

	nameBoth := "db-creds-both"
	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config:      testResourceSecretConfigBothData(nameBoth),
				ExpectError: regexp.MustCompile("Invalid Attribute Combination"),
			},
		},
	})
}

func testResourceSecretConfigBasic(name string) string {
	return fmt.Sprintf(`resource "gamefabric_secret" "test" {
  name        = "%s"
  environment = "dflt"
  description = "Backend database credentials"
  labels = {
    game = "my_first_game"
  }
  data = {
    db_user     = "dbuser123"
    db_password = "super-secret-pass-123"
  }
}`, name)
}

func testResourceSecretConfigUpdateData(name string) string {
	return fmt.Sprintf(`resource "gamefabric_secret" "test" {
  name        = "%s"
  environment = "dflt"
  description = "Backend database credentials - changed"
  labels = {
    game = "my_first_game-1"
  }
  data = {
    db_user     = "dbuser123"
    db_password = "super-secret-pass-123"
    one_more    = "secret-value"
  }
}`, name)
}

func testResourceSecretConfigUpdateDataWO(name string) string {
	return fmt.Sprintf(`resource "gamefabric_secret" "test" {
  name        = "%s"
  environment = "dflt"
  description = "Backend database credentials"
  labels = {
	game = "my_first_game"
  }
  data_wo = {
	db_user     = "different-user"
	db_password = "different-pass"
  }
  data_wo_version = 2
}
`, name)
}

func testResourceSecretConfigDataWO(name string) string {
	return fmt.Sprintf(`resource "gamefabric_secret" "test" {
  name        = "%s"
  environment = "dflt"
  description = "Backend database credentials"
  labels = {
    game = "my_first_game"
  }
  data_wo = {
    db_user     = "dbuser321"
    db_password = "super-secret-pass-321"
  }
  data_wo_version = 1
}`, name)
}

func testResourceSecretConfigBothData(name string) string {
	return fmt.Sprintf(`resource "gamefabric_secret" "test" {
  name        = "%s"
  environment = "dflt"
  description = "Updated description"
  labels = {
	game    = "my_first_game"
	version = "v2"
  }
  data = {
	db_user     = "dbuser123"
	db_password = "even-more-secret-pass-456"
  }
  data_wo = {
	db_user    = "dbuser123"
	db_password = "super-secret-pass-123"
  }
}`, name)
}

func testResourceSecretDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_secret" {
				continue
			}

			env, name, _ := strings.Cut(rs.Primary.ID, "/")
			resp, err := cs.CoreV1().Secrets(env).Get(t.Context(), name, metav1.GetOptions{})
			if err == nil && resp.Name == name {
				return fmt.Errorf("secret still exists: %s", rs.Primary.ID)
			}
		}
		return nil
	}
}
