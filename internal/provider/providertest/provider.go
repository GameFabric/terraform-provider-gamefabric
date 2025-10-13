package providertest

import (
	"testing"

	"github.com/gamefabric/gf-apicore/runtime"
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
	"github.com/gamefabric/gf-core/pkg/apiclient/fake"
	"github.com/gamefabric/terraform-provider-gamefabric/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ProtoV6ProviderFactories returns a map of provider factories for testing with the given runtime objects.
func ProtoV6ProviderFactories(t *testing.T, objs ...runtime.Object) (map[string]func() (tfprotov6.ProviderServer, error), clientset.Interface) {
	t.Helper()

	clientSet, err := fake.New(objs...)
	if err != nil {
		t.Fatalf("failed to create fake clientset: %v", err)
	}

	return map[string]func() (tfprotov6.ProviderServer, error){
		"gamefabric": providerserver.NewProtocol6WithError(provider.NewWithClientSet("test", clientSet)()),
	}, clientSet
}
