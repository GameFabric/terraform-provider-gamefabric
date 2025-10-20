package core_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestConfigFile(t *testing.T) {
	t.Parallel()

	cfgFile := &corev1.ConfigFile{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test-config-file",
			Environment: "dflt",
		},
		Data: "some-configuration-data",
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, cfgFile)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_configfile" "test1" {
  name = "test-config-file"
  environment = "dflt"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_configfile.test1", "name", "test-config-file"),
					resource.TestCheckResourceAttr("data.gamefabric_configfile.test1", "environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_configfile.test1", "data", "some-configuration-data"),
				),
			},
		},
	})
}
