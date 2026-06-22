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
				Config: fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
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
				}`, name),
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
				Config: fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
				  name       = "%s"
				  max_budget = 200
				  receivers  = ["my-receiver", "another-receiver"]
				  thresholds = [50, 90]
				}`, name),
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
				Config: fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
				  name       = "%s"
				  max_budget = 500
				  receivers  = ["on-call"]
				  interval = {
					start = 50
					step  = 50
				  }
				}`, name),
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
				Config: fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
				  name       = "%s"
				  max_budget = 500
				  receivers  = ["on-call"]
				  interval = {
					start = 100
					step  = 100
				  }
				}`, name),
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
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-budget"
				  max_budget = 100
				  receivers  = []
				  thresholds = [50]
				}`,
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
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-labels-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  thresholds = [50]
				  labels = {
					"_invalid" = "value"
				  }
				}`,
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
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-annotations-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  thresholds = [50]
				  annotations = {
					"-invalid" = "value"
				  }
				}`,
				ExpectError: regexp.MustCompile(`(?i)annotation key`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesMixedThresholdsAbsoluteFirst verifies that
// mixing absolute and percentage thresholds is rejected when the first entry is
// an absolute amount.
func TestCloudBudgetResource_ValidatesMixedThresholdsAbsoluteFirst(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-threshold-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  thresholds = [50, "75%"]
				}`,
				ExpectError: regexp.MustCompile(`(?i)same format`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesMixedThresholdsPercentFirst verifies that
// mixing percentage and absolute thresholds is rejected when the first entry is
// a percentage.
func TestCloudBudgetResource_ValidatesMixedThresholdsPercentFirst(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-threshold-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  thresholds = ["50%", 75]
				}`,
				ExpectError: regexp.MustCompile(`(?i)same format`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesZeroAbsoluteThreshold verifies that an
// absolute threshold of 0 is rejected at plan time.
func TestCloudBudgetResource_ValidatesZeroAbsoluteThreshold(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-threshold-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  thresholds = [0]
				}`,
				ExpectError: regexp.MustCompile(`(?i)greater than 0`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesZeroPercentThreshold verifies that a
// percentage threshold of 0% is rejected at plan time.
func TestCloudBudgetResource_ValidatesZeroPercentThreshold(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-threshold-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  thresholds = ["0%"]
				}`,
				ExpectError: regexp.MustCompile(`(?i)greater than 0`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesHundredPercentThreshold verifies that a
// percentage threshold of >= 100% is rejected at plan time.
func TestCloudBudgetResource_ValidatesHundredPercentThreshold(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-threshold-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  thresholds = ["100%"]
				}`,
				ExpectError: regexp.MustCompile(`(?i)less than 100`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesThresholdsNotAscending verifies that
// thresholds not in ascending order are rejected at plan time.
func TestCloudBudgetResource_ValidatesThresholdsNotAscending(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-threshold-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  thresholds = [50, 25]
				}`,
				ExpectError: regexp.MustCompile(`(?i)ascending`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesThresholdsDuplicateValue verifies that
// duplicate threshold values are rejected at plan time.
func TestCloudBudgetResource_ValidatesThresholdsDuplicateValue(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-threshold-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  thresholds = [50, 50]
				}`,
				ExpectError: regexp.MustCompile(`(?i)ascending`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesNonNumericPercentThreshold verifies that a
// non-numeric percentage threshold (e.g. "foo%") is rejected at plan time.
func TestCloudBudgetResource_ValidatesNonNumericPercentThreshold(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-threshold-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  thresholds = ["foo%"]
				}`,
				ExpectError: regexp.MustCompile(`(?i)(valid integer|invalid)`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesNegativePercentThreshold verifies that a
// negative percentage threshold is rejected at plan time.
func TestCloudBudgetResource_ValidatesNegativePercentThreshold(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-threshold-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  thresholds = ["-1%"]
				}`,
				ExpectError: regexp.MustCompile(`(?i)greater than 0`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesTooManyReceivers verifies that more than the
// maximum number of receivers is rejected at plan time.
func TestCloudBudgetResource_ValidatesTooManyReceivers(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-receivers-budget"
				  max_budget = 100
				  receivers  = ["a", "b", "c", "d", "e", "f"]
				  thresholds = [50]
				}`,
				ExpectError: regexp.MustCompile(`(?i)(exceed|most)`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesDuplicateReceiver verifies that duplicate
// receiver entries are rejected at plan time.
func TestCloudBudgetResource_ValidatesDuplicateReceiver(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-receivers-budget"
				  max_budget = 100
				  receivers  = ["ops", "ops"]
				  thresholds = [50]
				}`,
				ExpectError: regexp.MustCompile(`(?i)duplicate`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesEmptyReceiverEntry verifies that an empty
// string within the receivers list is rejected at plan time.
func TestCloudBudgetResource_ValidatesEmptyReceiverEntry(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-receivers-budget"
				  max_budget = 100
				  receivers  = ["ops", ""]
				  thresholds = [50]
				}`,
				ExpectError: regexp.MustCompile(`(?i)empty`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesMixedIntervalFormat verifies that mixing
// absolute and percentage formats across start and step is rejected.
func TestCloudBudgetResource_ValidatesMixedIntervalFormat(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-interval-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  interval = {
					start = 50
					step  = "10%"
				  }
				}`,
				ExpectError: regexp.MustCompile(`(?i)same`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesIntervalStepExceedsFiftyPercent verifies
// that a percentage step greater than 50% is rejected.
func TestCloudBudgetResource_ValidatesIntervalStepExceedsFiftyPercent(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-interval-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  interval = {
					start = "10%"
					step  = "51%"
				  }
				}`,
				ExpectError: regexp.MustCompile(`(?i)50`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesIntervalZeroStep verifies that an interval
// step of 0 is rejected at plan time.
func TestCloudBudgetResource_ValidatesIntervalZeroStep(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-interval-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  interval = {
					start = 10
					step  = 0
				  }
				}`,
				ExpectError: regexp.MustCompile(`(?i)greater than 0`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesIntervalNegativeStep verifies that a
// negative interval step is rejected at plan time.
func TestCloudBudgetResource_ValidatesIntervalNegativeStep(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-interval-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  interval = {
					start = 10
					step  = -5
				  }
				}`,
				ExpectError: regexp.MustCompile(`(?i)greater than 0`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesIntervalNonNumericStart verifies that a
// non-numeric percentage start (e.g. "abc%") is rejected at plan time.
func TestCloudBudgetResource_ValidatesIntervalNonNumericStart(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-interval-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  interval = {
					start = "abc%"
					step  = "10%"
				  }
				}`,
				ExpectError: regexp.MustCompile(`(?i)(valid integer|invalid)`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesIntervalNonNumericStep verifies that a
// non-numeric percentage step (e.g. "abc%") is rejected at plan time.
func TestCloudBudgetResource_ValidatesIntervalNonNumericStep(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_cloudbudget" "test" {
				  name       = "bad-interval-budget"
				  max_budget = 100
				  receivers  = ["ops"]
				  interval = {
					start = "10%"
					step  = "abc%"
				  }
				}`,
				ExpectError: regexp.MustCompile(`(?i)(valid integer|invalid)`),
			},
		},
	})
}

// TestCloudBudgetResource_ValidatesIntervalStepAtFiftyPercentAllowed verifies
// that a percentage step of exactly 50% is accepted.
func TestCloudBudgetResource_ValidatesIntervalStepAtFiftyPercentAllowed(t *testing.T) {
	t.Parallel()

	name := "valid-interval-budget"
	pf, cs := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		CheckDestroy:             testCloudBudgetDestroy(t, cs),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`resource "gamefabric_cloudbudget" "test" {
				  name       = "%s"
				  max_budget = 100
				  receivers  = ["ops"]
				  interval = {
					start = "10%%"
					step  = "50%%"
				  }
				}`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "interval.start", "10%"),
					resource.TestCheckResourceAttr("gamefabric_cloudbudget.test", "interval.step", "50%"),
				),
			},
		},
	})
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
