package protection_test

import (
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestProtocols(t *testing.T) {
	proto1 := &protectionv1.Protocol{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-protocol1",
		},
		Spec: protectionv1.ProtocolSpec{
			DisplayName: "My Protocol 1",
			Description: "My Protection Protocol 1",
			Protocol:    protectionv1.NetworkProtocolTCP,
		},
	}
	proto2 := &protectionv1.Protocol{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-protocol2",
		},
		Spec: protectionv1.ProtocolSpec{
			DisplayName: "My Protocol 2",
			Description: "My Protection Protocol 2",
			Protocol:    protectionv1.NetworkProtocolUDP,
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, proto1, proto2)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_protection_protocols" "test1" {
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocols.test1", "protocols.#", "2"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocols.test1", "protocols.0.name", "test-protocol1"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocols.test1", "protocols.0.display_name", "My Protocol 1"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocols.test1", "protocols.0.description", "My Protection Protocol 1"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocols.test1", "protocols.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocols.test1", "protocols.1.name", "test-protocol2"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocols.test1", "protocols.1.display_name", "My Protocol 2"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocols.test1", "protocols.1.description", "My Protection Protocol 2"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocols.test1", "protocols.1.protocol", "UDP"),
				),
			},
		},
	})
}
