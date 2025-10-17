package protection_test

import (
	"fmt"
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestGatewayPolicy(t *testing.T) {
	name := "unreal"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceGatewayPolicyDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testResourceGatewayPolicyConfigBasic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "display_name", "My Gateway Policy"),
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "labels.example", "label"),
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "annotations.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "annotations.example", "annotation"),
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "destination_cidrs.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "destination_cidrs.0", "1.2.3.4/32"),
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "destination_cidrs.1", "2.3.0.0/8"),
				),
			},
			{
				ResourceName:      "gamefabric_protection_gatewaypolicy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceGatewayPolicyConfigBasicWithDescription(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "display_name", "My Gateway Policy"),
					resource.TestCheckResourceAttr("gamefabric_protection_gatewaypolicy.test", "description", "My Gateway Policy Description"),
				),
			},
		},
	})
}

func TestGatewayPolicy_ValidatesCIDR(t *testing.T) {
	name := "unreal"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceGatewayPolicyDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config:      testResourceGatewayPolicyConfigInvalid(name),
				ExpectError: regexp.MustCompile(`invalid CIDR`),
			},
		},
	})
}

func TestGatewayPolicy_ValidatesLabels(t *testing.T) {
	name := "unreal"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceGatewayPolicyDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config:      testResourceGatewayPolicyConfigInvalid(name),
				ExpectError: regexp.MustCompile(`Label key .* is not valid`),
			},
		},
	})
}

func TestGatewayPolicy_ValidatesAnnotations(t *testing.T) {
	name := "unreal"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testResourceGatewayPolicyDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config:      testResourceGatewayPolicyConfigInvalid(name),
				ExpectError: regexp.MustCompile(`Annotation key .* is not valid`),
			},
		},
	})
}

func testResourceGatewayPolicyConfigBasic(name string) string {
	return fmt.Sprintf(`resource "gamefabric_protection_gatewaypolicy" "test" {
  name = "%s"
  display_name = "My Gateway Policy"
  labels = {
    "example" = "label"
  }
  annotations = {
    "example" = "annotation"
  }
  destination_cidrs = [
    "1.2.3.4/32", 
    "2.3.0.0/8",
  ]
}`, name)
}

func testResourceGatewayPolicyConfigBasicWithDescription(name string) string {
	return fmt.Sprintf(`resource "gamefabric_protection_gatewaypolicy" "test" {
  name = "%s"
  display_name = "My Gateway Policy"
  description = "My Gateway Policy Description"
  destination_cidrs = [
    "1.2.3.4/32", 
    "2.3.0.0/8",
  ]
}`, name)
}

func testResourceGatewayPolicyConfigInvalid(name string) string {
	return fmt.Sprintf(`resource "gamefabric_protection_gatewaypolicy" "test" {
  name = "%s"
  display_name = "My Gateway Policy"
  labels = {
    "_" = "label"
  }
  annotations = {
	"-" = "annotation"
  }
  destination_cidrs = [
    "test",
  ]
}`, name)
}

func testResourceGatewayPolicyDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_protection_gatewaypolicy" {
				continue
			}

			resp, err := cs.ProtectionV1().GatewayPolicies().Get(t.Context(), rs.Primary.ID, metav1.GetOptions{})
			if err == nil && resp.Name == rs.Primary.ID {
				return fmt.Errorf("gatewaypolicy still exists: %s", rs.Primary.ID)
			}
		}
		return nil
	}
}
