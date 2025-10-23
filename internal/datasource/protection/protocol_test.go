package protection_test

import (
	"regexp"
	"testing"

	metav1 "github.com/gamefabric/gf-apicore/apis/meta/v1"
	protectionv1 "github.com/gamefabric/gf-core/pkg/api/protection/v1"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider/providertest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestProtocol(t *testing.T) {
	proto := &protectionv1.Protocol{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-protocol",
		},
		Spec: protectionv1.ProtocolSpec{
			DisplayName: "My Protocol",
			Description: "My Protection Protocol",
			Protocol:    protectionv1.NetworkProtocolTCP,
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, proto)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_protection_protocol" "test1" {
  name = "test-protocol"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocol.test1", "name", "test-protocol"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocol.test1", "display_name", "My Protocol"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocol.test1", "description", "My Protection Protocol"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocol.test1", "protocol", "TCP"),
				),
			},
			{
				Config: `data "gamefabric_protection_protocol" "test2" {
  display_name = "My Protocol"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocol.test2", "name", "test-protocol"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocol.test2", "display_name", "My Protocol"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocol.test2", "description", "My Protection Protocol"),
					resource.TestCheckResourceAttr("data.gamefabric_protection_protocol.test2", "protocol", "TCP"),
				),
			},
		},
	})
}

func TestProtocol_HandlesMultipleMatches(t *testing.T) {
	proto := &protectionv1.Protocol{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-protocol",
		},
		Spec: protectionv1.ProtocolSpec{
			DisplayName: "My Protocol",
		},
	}
	otherProto := &protectionv1.Protocol{
		ObjectMeta: metav1.ObjectMeta{
			Name: "other-region",
		},
		Spec: protectionv1.ProtocolSpec{
			DisplayName: "My Protocol",
		},
	}

	pf, _ := providertest.ProtoV6ProviderFactories(t, proto, otherProto)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_protection_protocol" "test" {
  display_name = "My Protocol"
}
`,
				ExpectError: regexp.MustCompile(`Multiple Protocols Found`),
			},
		},
	})
}

func TestProtocol_HandlesMultipleSelectors(t *testing.T) {
	pf, _ := providertest.ProtoV6ProviderFactories(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: pf,
		Steps: []resource.TestStep{
			{
				Config: `data "gamefabric_protection_protocol" "test" {
  name = "test-protocol"
  display_name = "My Protocol"
}
`,
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
		},
	})
}
