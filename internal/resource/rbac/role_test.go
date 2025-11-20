package rbac_test

import (
	"fmt"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestRoleResource(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testRoleResourceDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_role" "test" {
	name = "test-role"
	rules = [
		{
			api_groups   = ["*"]
			resources    = ["*"]
			verbs        = ["*"]
			environments = ["*"]
		}
	]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_role.test", "name", "test-role"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.api_groups.0", "*"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.resources.0", "*"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.verbs.0", "*"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.environments.0", "*"),
				),
			},
			{
				ResourceName:      "gamefabric_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: `resource "gamefabric_role" "test" {
	name = "test-role-optional"
	rules = [
		{
			api_groups   = ["*"]
			resources    = ["*"]
			verbs        = ["*"]
			environments = ["*"]
			scopes       = ["scope1", "scope2"]
			resource_names = ["resource1", "resource2"]
		}
	]
	labels = {
		"key1" = "value1"
		"key2" = "value2"
	}
	annotations = {
		"akey1" = "avalue1"
		"akey2" = "avalue2"
	}
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_role.test", "name", "test-role-optional"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.scopes.0", "scope1"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.scopes.1", "scope2"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.resource_names.0", "resource1"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.resource_names.1", "resource2"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "labels.key1", "value1"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "labels.key2", "value2"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "annotations.akey1", "avalue1"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "annotations.akey2", "avalue2"),
				),
			},
			{
				Config: `resource "gamefabric_role" "test" {
	name = "test-role-update"
	rules = [
		{
			api_groups   = ["groupone"]
			resources    = ["resourceone"]
			verbs        = ["get", "list"]
			environments = ["env1"]
		}
	]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_role.test", "name", "test-role-update"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.api_groups.0", "groupone"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.resources.0", "resourceone"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.verbs.0", "get"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.verbs.1", "list"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.environments.0", "env1"),
				),
			},
			{
				Config: `resource "gamefabric_role" "test" {
	name = "test-role-update"
	rules = [
		{
			api_groups   = ["groupone"]
			resources    = ["resourceone"]
			verbs        = ["get", "list"]
			environments = ["env1"]
			scopes       = ["scope1"]
			resource_names = ["res1"]
		}
	]
	labels = {
		"key1" = "value1"
	}
	annotations = {
		"akey1" = "avalue1"
	}
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_role.test", "name", "test-role-update"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.api_groups.0", "groupone"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.resources.0", "resourceone"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.verbs.0", "get"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.verbs.1", "list"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.environments.0", "env1"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.scopes.0", "scope1"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "rules.0.resource_names.0", "res1"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "labels.key1", "value1"),
					resource.TestCheckResourceAttr("gamefabric_role.test", "annotations.akey1", "avalue1"),
				),
			},
		},
	})
}

func testRoleResourceDestroy(t *testing.T, cs clientset.Interface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_role" {
				continue
			}

			resp, err := cs.RBACV1().Roles().Get(t.Context(), rs.Primary.ID, metav1.GetOptions{})
			if err == nil {
				return fmt.Errorf("role %q still exists: %+v", rs.Primary.ID, resp)
			}
		}
		return nil
	}
}
