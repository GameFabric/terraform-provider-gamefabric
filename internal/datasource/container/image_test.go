package container_test

import (
	"testing"
	"time"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	containerv1 "github.com/gamefabric/gf-core/pkg/api/container/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestImage(t *testing.T) {
	img := &containerv1.Image{
		ImageObjectMeta: containerv1.ImageObjectMeta{
			ObjectMeta: metav1.ObjectMeta{Name: "test-image"},
			Branch:     "test-branch",
		},
		Spec: containerv1.ImageSpec{
			Image: "my-image",
			Tag:   "v1.0.0",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, img)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_image" "test1" {
  branch = "test-branch"
  image  = "my-image"
  tag    = "v1.0.0"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_image.test1", "image_target.name", "test-image"),
					resource.TestCheckResourceAttr("data.gamefabric_image.test1", "image_target.branch", "test-branch"),
				),
			},
		},
	})
}

func TestImage_Latest(t *testing.T) {
	now := time.Now()
	before := now.Add(-1 * time.Hour)

	img1 := &containerv1.Image{
		ImageObjectMeta: containerv1.ImageObjectMeta{
			ObjectMeta: metav1.ObjectMeta{
				Name:             "test-image1",
				CreatedTimestamp: before,
			},
			Branch: "test-branch",
		},
		Spec: containerv1.ImageSpec{
			Image: "my-image",
			Tag:   "v1.0.0",
		},
	}
	img2 := &containerv1.Image{
		ImageObjectMeta: containerv1.ImageObjectMeta{
			ObjectMeta: metav1.ObjectMeta{
				Name:             "test-image2",
				CreatedTimestamp: now,
			},
			Branch: "test-branch",
		},
		Spec: containerv1.ImageSpec{
			Image: "my-image",
			Tag:   "v1.0.1",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, img1, img2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_image" "test1" {
  branch = "test-branch"
  image  = "my-image"
  tag    = "latest"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_image.test1", "branch", "test-branch"),
					resource.TestCheckResourceAttr("data.gamefabric_image.test1", "image", "my-image"),
					resource.TestCheckResourceAttr("data.gamefabric_image.test1", "tag", "latest"),
					resource.TestCheckResourceAttr("data.gamefabric_image.test1", "image_target.name", "test-image2"),
					resource.TestCheckResourceAttr("data.gamefabric_image.test1", "image_target.branch", "test-branch"),
				),
			},
		},
	})
}
