package billing_test

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

// TestCloudBudgetResource exercises the full CRUD lifecycle of a CloudBudget
// that uses explicit thresholds.
func TestCloudBudgetResource(t *testing.T) {
	t.Parallel()

	name := "example-budget"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCloudBudgetDestroy(t, cs),
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testCloudBudgetConfigThresholds(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "max_budget", "100"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "suspended", "false"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "receivers.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "receivers.0", "my-receiver"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "thresholds.#", "3"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "thresholds.0", "25"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "thresholds.1", "50"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "thresholds.2", "75"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "labels.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "labels.team", "ops"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "annotations.%", "1"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "annotations.note", "cost-control"),
				),
			},
			// ImportState.
			{
				ResourceName:            "gamefabric_cloudbudget.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"suspended"},
			},
			// Update: change receivers, max_budget, and thresholds.
			{
				Config: testCloudBudgetConfigThresholdsUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "max_budget", "200"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "receivers.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "receivers.0", "my-receiver"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "receivers.1", "another-receiver"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "thresholds.#", "2"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "thresholds.0", "50"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "thresholds.1", "90"),
				),
			},
		},
	})
}

// TestCloudBudgetResource_WithInterval exercises a CloudBudget that uses the
// interval-based threshold expansion instead of an explicit thresholds list.
func TestCloudBudgetResource_WithInterval(t *testing.T) {
	t.Parallel()

	name := "interval-budget"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCloudBudgetDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testCloudBudgetConfigInterval(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "name", name),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "max_budget", "500"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "receivers.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "receivers.0", "on-call"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "interval.start", "50"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "interval.step", "50"),
					resource.TestCheckNoResourceAttr("gamefabric_cloudbudget.test", "thresholds"),
				),
			},
			// ImportState.
			{
				ResourceName:            "gamefabric_cloudbudget.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"suspended"},
			},
			// Update: change interval step.
			{
				Config: testCloudBudgetConfigIntervalUpdated(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "interval.start", "100"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "interval.step", "100"),
				),
			},
		},
	})
}

// TestCloudBudgetResource_Suspended verifies that the suspended field can be
// toggled independently of other fields.
func TestCloudBudgetResource_Suspended(t *testing.T) {
	t.Parallel()

	name := "suspend-budget"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCloudBudgetDestroy(t, cs),
		Steps: []resource.TestStep{
			// Create with suspended = false (default).
			{
				Config: testCloudBudgetConfigMinimal(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "suspended", "false"),
				),
			},
			// Suspend.
			{
				Config: testCloudBudgetConfigSuspended(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "suspended", "true"),
				),
			},
			// Unsuspend.
			{
				Config: testCloudBudgetConfigSuspended(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "suspended", "false"),
				),
			},
		},
	})
}

// TestCloudBudgetResource_NameRequiresReplace verifies that changing the name
// forces a resource replacement (the old resource is destroyed and a new one
// is created under the new name).
func TestCloudBudgetResource_NameRequiresReplace(t *testing.T) {
	t.Parallel()

	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCloudBudgetDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: testCloudBudgetConfigMinimal("original-budget"),
				Check:  resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "name", "original-budget"),
			},
			{
				Config: testCloudBudgetConfigMinimal("renamed-budget"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "name", "renamed-budget"),
				),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesEmptyReceivers verifies that an empty
// receivers list is rejected at plan time.
func TestCloudBudgetResource_ValidatesEmptyReceivers(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config:      testCloudBudgetConfigEmptyReceivers("bad-budget"),
				ExpectError: regexp.MustCompile(`(?i)at least 1`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesName verifies that an invalid resource name
// is rejected at plan time.
func TestCloudBudgetResource_ValidatesName(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config:      testCloudBudgetConfigMinimal("Invalid_Name"),
				ExpectError: regexp.MustCompile(`(?i)name`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesLabels verifies that invalid label keys are
// rejected at plan time.
func TestCloudBudgetResource_ValidatesLabels(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config:      testCloudBudgetConfigInvalidLabels("bad-labels-budget"),
				ExpectError: regexp.MustCompile(`(?i)label key`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesAnnotations verifies that invalid annotation
// keys are rejected at plan time.
func TestCloudBudgetResource_ValidatesAnnotations(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config:      testCloudBudgetConfigInvalidAnnotations("bad-annotations-budget"),
				ExpectError: regexp.MustCompile(`(?i)annotation key`),
			},
		},
	})
}

func testCloudBudgetConfigThresholds(name string) string {
	return fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
  name       = "%s"
  max_budget = 100
  receivers  = ["my-receiver"]
  thresholds = [25, 50, 75]
  labels = {
    team = "ops"
  }
  annotations = {
    note = "cost-control"
  }
}`, name)
}

func testCloudBudgetConfigThresholdsUpdated(name string) string {
	return fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
  name       = "%s"
  max_budget = 200
  receivers  = ["my-receiver", "another-receiver"]
  thresholds = [50, 90]
}`, name)
}

func testCloudBudgetConfigInterval(name string) string {
	return fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
  name       = "%s"
  max_budget = 500
  receivers  = ["on-call"]
  interval = {
    start = 50
    step  = 50
  }
}`, name)
}

func testCloudBudgetConfigIntervalUpdated(name string) string {
	return fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
  name       = "%s"
  max_budget = 500
  receivers  = ["on-call"]
  interval = {
    start = 100
    step  = 100
  }
}`, name)
}

func testCloudBudgetConfigMinimal(name string) string {
	return fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
  name       = "%s"
  max_budget = 50
  receivers  = ["ops"]
  thresholds = [80]
}`, name)
}

func testCloudBudgetConfigSuspended(name string, suspended bool) string {
	return fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
  name       = "%s"
  max_budget = 50
  receivers  = ["ops"]
  thresholds = [80]
  suspended  = %t
}`, name, suspended)
}

func testCloudBudgetConfigEmptyReceivers(name string) string {
	return fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
  name       = "%s"
  max_budget = 100
  receivers  = []
  thresholds = [50]
}`, name)
}

func testCloudBudgetConfigInvalidLabels(name string) string {
	return fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
  name       = "%s"
  max_budget = 100
  receivers  = ["ops"]
  thresholds = [50]
  labels = {
    "_invalid" = "value"
  }
}`, name)
}

func testCloudBudgetConfigInvalidAnnotations(name string) string {
	return fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
  name       = "%s"
  max_budget = 100
  receivers  = ["ops"]
  thresholds = [50]
  annotations = {
    "-invalid" = "value"
  }
}`, name)
}

func testCloudBudgetDestroy(t *testing.T, cs clientset.Interface) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "gamefabric_cloudbudget" {
				continue
			}

			resp, err := cs.BillingV2Alpha1().CloudBudgets().Get(t.Context(), rs.Primary.ID, metav1.GetOptions{})
			if err == nil && resp.Name == rs.Primary.ID {
				return fmt.Errorf("cloudbudget still exists: %s", rs.Primary.ID)
			}
		}
		return nil
	}
}
