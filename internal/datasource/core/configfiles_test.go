package core_test

import (
	"testing"

	corev1 "github.com/gamefabric/gf-core/pkg/api/core/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestConfigFiles(t *testing.T) {
	cfgFile := &corev1.ConfigFile{
		// ObjectMeta
		Name:        "test-config-file",
		Environment: "dflt",
		Labels: map[string]string{
			"test": "lbl",
		},

		Data: "some-configuration-data",
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, cfgFile)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_configfiles" "test1" {
  environment = "dflt"
  label_filter = {
    test = "lbl"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_configfiles.test1", "environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_configfiles.test1", "label_filter.test", "lbl"),
					resource.TestCheckResourceAttr("data.gamefabric_configfiles.test1", "config_files.#", "1"),
					resource.TestCheckResourceAttr("data.gamefabric_configfiles.test1", "config_files.0.name", "test-config-file"),
					resource.TestCheckResourceAttr("data.gamefabric_configfiles.test1", "config_files.0.environment", "dflt"),
					resource.TestCheckResourceAttr("data.gamefabric_configfiles.test1", "config_files.0.data", "some-configuration-data"),
				),
			},
		},
	})
}
