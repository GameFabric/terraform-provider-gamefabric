package context

import (
	"github.com/gamefabric/gf-core/pkg/apiclient/clientset"
)

// Context is the provider context.
type Context struct {
	ClientSet clientset.Interface
}

// NewContext creates a new context with the given client set.
func NewContext(clientSet clientset.Interface) *Context {
	return &Context{
		ClientSet: clientSet,
	}
}
