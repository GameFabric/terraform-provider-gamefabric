package notification_test

import (
	"context"
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	billingv2alpha1 "github.com/gamefabric/gf-core/pkg/api/billing/v2alpha1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestReceiverResource_CRUD exercises create, read, update, and delete of a
// Receiver that is NOT referenced by any CloudBudget.
func TestReceiverResource_CRUD(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `resource "gamefabric_receiver" "test" {
  name     = "test-receiver"
  email_to = ["ops@example.com"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_receiver.test", "name", "test-receiver"),
					resource.TestCheckResourceAttr("gamefabric_receiver.test", "email_to.#", "1"),
					resource.TestCheckResourceAttr("gamefabric_receiver.test", "email_to.0", "ops@example.com"),
				),
			},
			{
				// Update email list.
				Config: `resource "gamefabric_receiver" "test" {
  name     = "test-receiver"
  email_to = ["ops@example.com", "alerts@example.com"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("gamefabric_receiver.test", "email_to.#", "2"),
				),
			},
			// The final step destroys the resource; no error expected because
			// no CloudBudget references this receiver.
		},
	})
}

// TestReceiverResource_DestroyBlockedByCloudBudget verifies that the provider
// blocks (plan-time error) the deletion of a Receiver that is still referenced
// by a CloudBudget.
//
// This covers the ordering bug where Terraform, after the user removes both
// the receiver reference from the CloudBudget config AND the receiver resource
// in one edit, may try to delete the Receiver before the CloudBudget is
// updated. The ModifyPlan hook detects this during terraform plan and surfaces
// a clear error asking the user to apply the two-step fix.
func TestReceiverResource_DestroyBlockedByCloudBudget(t *testing.T) {
	t.Parallel()

	// Pre-populate the fake API with only the CloudBudget that references the
	// receiver. The receiver itself is created by Terraform in step 1 so that
	// there is no "already exists" conflict. By the time step 2 runs (destroy),
	// the CloudBudget is already present and ModifyPlan can detect the reference.
	budget := &billingv2alpha1.CloudBudget{
		ObjectMeta: metav1.ObjectMeta{Name: "example-budget"},
		Spec: billingv2alpha1.CloudBudgetSpec{
			Receivers: []string{"blocked-receiver"},
			MaxBudget: 100,
		},
	}

	pf, cs := providertest.ProtoV6ProviderFactories(t, budget)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			// Step 1: Create the receiver via Terraform.
			{
				Config: `resource "gamefabric_receiver" "test" {
  name     = "blocked-receiver"
  email_to = ["blocked@example.com"]
}`,
				Check: resource.TestCheckResourceAttr("gamefabric_receiver.test", "name", "blocked-receiver"),
			},
			// Step 2: Remove the resource from config entirely; this triggers a
			// destroy plan. ModifyPlan should detect the CloudBudget reference and
			// return an error before any API call is made.
			{
				Config:      `# receiver removed – CloudBudget still references it`,
				ExpectError: regexp.MustCompile(`(?i)receiver still in use`),
			},
			// Step 3: Delete the CloudBudget from the fake API (simulating the
			// user having first removed the receiver reference and applied), then
			// remove the receiver resource. This lets the test framework clean up
			// without hitting the ModifyPlan guard again.
			{
				PreConfig: func() {
					_ = cs.BillingV2Alpha1().CloudBudgets().Delete(
						context.Background(), "example-budget", metav1.DeleteOptions{},
					)
				},
				Config: `# receiver removed – no more CloudBudget references`,
			},
		},
	})
}
