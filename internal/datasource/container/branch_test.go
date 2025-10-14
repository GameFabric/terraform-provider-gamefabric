package container_test

import (
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBranch(t *testing.T) {
	t.Parallel()

	branch := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{Name: "test-branch"},
		Spec: containerv1.BranchSpec{
			DisplayName: "My Branch",
			RetentionPolicyRules: []containerv1.BranchImageRetentionPolicyRule{
				{
					Name:       "rule1",
					ImageRegex: "a.*",
					TagRegex:   "b.*",
					KeepCount:  10,
					KeepDays:   30,
				},
			},
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, branch)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_branch" "test1" {
  name = "test-branch"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_branch.test1", "name", "test-branch"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test1", "display_name", "My Branch"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test1", "retention_policy_rules.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test1", "retention_policy_rules.0.name", "rule1"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test1", "retention_policy_rules.0.image_regex", "a.*"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test1", "retention_policy_rules.0.tag_regex", "b.*"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test1", "retention_policy_rules.0.keep_count", "10"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test1", "retention_policy_rules.0.keep_days", "30"),
				),
			},
			{
				Config: `data "gamefabric_branch" "test2" {
  display_name = "My Branch"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_branch.test2", "name", "test-branch"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test2", "display_name", "My Branch"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test2", "retention_policy_rules.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test2", "retention_policy_rules.0.name", "rule1"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test2", "retention_policy_rules.0.image_regex", "a.*"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test2", "retention_policy_rules.0.tag_regex", "b.*"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test2", "retention_policy_rules.0.keep_count", "10"),
					resource.TestCheckResourceAttr("data.gamefabric_branch.test2", "retention_policy_rules.0.keep_days", "30"),
				),
			},
		},
	})
}

func TestBranch_HandlesMultipleMatches(t *testing.T) {
	t.Parallel()

	branch := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{Name: "test-branch"},
		Spec: containerv1.BranchSpec{
			DisplayName: "My Branch",
		},
	}
	otherBranch := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{Name: "other-branch"},
		Spec: containerv1.BranchSpec{
			DisplayName: "My Branch",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, branch, otherBranch)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_branch" "test" {
  display_name = "My Branch"
}
`,
				ExpectError: regexp.MustCompile(`Multiple Branches Found`),
			},
		},
	})
}

func TestBranch_HandlesMultipleSelectors(t *testing.T) {
	t.Parallel()

	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_branch" "test" {
  name = "test-branch"
  display_name = "My Branch"
}
`,
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}
