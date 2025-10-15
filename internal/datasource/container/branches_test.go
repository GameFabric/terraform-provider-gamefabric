package container_test

import (
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestBranches(t *testing.T) {
	t.Parallel()

	branch1 := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-branch-1",
			Labels: map[string]string{
				"baremetal": "true",
			},
		},
		Spec: containerv1.BranchSpec{
			DisplayName: "Test Branch 1",
			RetentionPolicyRules: []containerv1.BranchImageRetentionPolicyRule{
				{
					Name:       "default",
					KeepCount:  10,
					KeepDays:   30,
					ImageRegex: ".*",
					TagRegex:   ".*",
				},
			},
		},
	}

	branch2 := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-branch-2",
			Labels: map[string]string{
				"baremetal": "false",
			},
		},
		Spec: containerv1.BranchSpec{
			DisplayName: "Test Branch 2",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, branch1, branch2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_branches" "test" {
  label_filter = {
    baremetal = "true"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "label_filter.baremetal", "true"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.0.name", "test-branch-1"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.0.retention_policy_rules.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.0.retention_policy_rules.0.name", "default"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.0.retention_policy_rules.0.keep_count", "10"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.0.retention_policy_rules.0.keep_days", "30"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.0.retention_policy_rules.0.image_regex", ".*"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.0.retention_policy_rules.0.tag_regex", ".*"),
				),
			},
		},
	})
}

func TestBranches_FilterByName(t *testing.T) {
	t.Parallel()

	branch1 := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-branch-1",
		},
		Spec: containerv1.BranchSpec{
			DisplayName: "Test Branch 1",
		},
	}

	branch2 := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-branch-2",
		},
		Spec: containerv1.BranchSpec{
			DisplayName: "Test Branch 2",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, branch1, branch2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_branches" "test" {
  name = "test-branch-1"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "name", "test-branch-1"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.0.name", "test-branch-1"),
				),
			},
		},
	})
}

func TestBranches_FilterByDisplayName(t *testing.T) {
	t.Parallel()

	branch1 := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-branch-1",
		},
		Spec: containerv1.BranchSpec{
			DisplayName: "Test Branch 1",
		},
	}

	branch2 := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-branch-2",
		},
		Spec: containerv1.BranchSpec{
			DisplayName: "Test Branch 2",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, branch1, branch2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_branches" "test" {
  display_name = "Test Branch 2"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "display_name", "Test Branch 2"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.0.name", "test-branch-2"),
				),
			},
		},
	})
}

func TestBranches_AllowsGettingAll(t *testing.T) {
	t.Parallel()

	branch1 := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-branch-1",
		},
		Spec: containerv1.BranchSpec{
			DisplayName: "Test Branch 1",
		},
	}

	branch2 := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-branch-2",
		},
		Spec: containerv1.BranchSpec{
			DisplayName: "Test Branch 2",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, branch1, branch2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_branches" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.0.name", "test-branch-1"),
					resource.TestCheckResourceAttr("data.gamefabric_branches.test", "branches.1.name", "test-branch-2"),
				),
			},
		},
	})
}

func TestBranches_ErrorsOnDuplicateDisplayName(t *testing.T) {
	t.Parallel()

	branch1 := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-branch-1",
		},
		Spec: containerv1.BranchSpec{
			DisplayName: "My Branch",
		},
	}

	branch2 := &containerv1.Branch{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-branch-2",
		},
		Spec: containerv1.BranchSpec{
			DisplayName: "My Branch",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, branch1, branch2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_branches" "test" {
  display_name = "My Branch"
}
`,
				ExpectError: regexp.MustCompile(`Multiple Branches Found`),
			},
		},
	})
}
