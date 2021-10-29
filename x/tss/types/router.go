package types

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// Router implements a tss Handler router.
type Router interface {
	AddRoute(module string, handler exported.Handler) Router
	HasRoute(module string) bool
	GetRoute(module string) exported.Handler
	Seal()
}

var _ Router = (*router)(nil)

type router struct {
	routes map[string]exported.Handler
	sealed bool
}

// NewRouter creates a new Router interface instance
func NewRouter() Router {
	return &router{
		routes: make(map[string]exported.Handler),
	}
}

// Seal prevents additional route handlers from  being added to the router.
func (r *router) Seal() {
	r.sealed = true
}

// AddRoute registers a tss handler for a given path and returns the handler.
// Panics if the router is sealed, module is an empty string, or if the module has been registered already.
func (r *router) AddRoute(module string, handler exported.Handler) Router {
	if r.sealed {
		panic("cannot add handler (router sealed)")
	}

	if module == "" {
		panic("module name cannot be an empty string")
	}

	if r.HasRoute(module) {
		panic(fmt.Sprintf("handler for module %s has already been registered", module))
	}

	r.routes[module] = handler
	return r
}

// HasRoute returns true if the router has an handler registered for the given module
func (r *router) HasRoute(module string) bool {
	return r.routes[module] != nil
}

// GetRoute returns a Handler for a given module.
func (r *router) GetRoute(module string) exported.Handler {
	if !r.HasRoute(module) {
		panic(fmt.Sprintf("handler for module \"%s\" not registered", module))
	}

	return r.routes[module]
}
