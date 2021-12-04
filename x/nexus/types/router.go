package types

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// Router implements a AddressValidator router based on module name.
type Router interface {
	AddAddressValidator(module string, handler exported.AddressValidator) Router
	HasAddressValidator(module string) bool
	GetAddressValidator(module string) exported.AddressValidator
	Seal()
}

var _ Router = (*router)(nil)

type router struct {
	routes map[string]exported.AddressValidator
	sealed bool
}

// NewRouter creates a new Router interface instance
func NewRouter() Router {
	return &router{
		routes: make(map[string]exported.AddressValidator),
	}
}

// Seal prevents additional route handlers from  being added to the router.
func (r *router) Seal() {
	r.sealed = true
}

// AddAddressValidator registers a nexus handler for a given path and returns the handler.
// Panics if the router is sealed, module is an empty string, or if the module has been registered already.
func (r *router) AddAddressValidator(module string, handler exported.AddressValidator) Router {
	if r.sealed {
		panic("cannot add handler (router sealed)")
	}

	if module == "" {
		panic("module name cannot be an empty string")
	}

	if r.HasAddressValidator(module) {
		panic(fmt.Sprintf("handler for module %s has already been registered", module))
	}

	r.routes[module] = handler
	return r
}

// HasAddressValidator returns true if the router has an handler registered for the given module
func (r *router) HasAddressValidator(module string) bool {
	return r.routes[module] != nil
}

// GetAddressValidator returns a Handler for a given module.
func (r *router) GetAddressValidator(module string) exported.AddressValidator {
	if !r.HasAddressValidator(module) {
		panic(fmt.Sprintf("handler for module \"%s\" not registered", module))
	}

	return r.routes[module]
}
