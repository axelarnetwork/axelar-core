package types

import (
	fmt "fmt"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// SigRouter implements a sig router based on module name
type SigRouter interface {
	AddHandler(module string, handler exported.SigHandler) SigRouter
	HasHandler(module string) bool
	GetHandler(module string) exported.SigHandler
	Seal()
}

var _ SigRouter = (*router)(nil)

type router struct {
	handlers map[string]exported.SigHandler
	sealed   bool
}

// NewSigRouter is the contructor for sig router
func NewSigRouter() SigRouter {
	return &router{
		handlers: make(map[string]exported.SigHandler),
	}
}

// AddHandler registers a new handler for the given module; panics if the
// router is sealed, if the module is invalid, or if the module has been
// registered already.
func (r *router) AddHandler(module string, handler exported.SigHandler) SigRouter {
	if handler == nil {
		panic("nil handler received")
	}

	if r.sealed {
		panic("router already sealed")
	}

	funcs.MustNoErr(utils.ValidateString(module))

	if r.HasHandler(module) {
		panic(fmt.Sprintf("handler for module %s already registered", module))
	}

	r.handlers[module] = handler

	return r
}

// HasHandler returns true if the router has a handler registered for the
// given module
func (r router) HasHandler(module string) bool {
	_, ok := r.handlers[module]

	return ok
}

// GetHandler returns the handler for the given module.
func (r router) GetHandler(module string) exported.SigHandler {
	if !r.HasHandler(module) {
		panic(fmt.Sprintf("no handler for module %s registered", module))
	}

	return r.handlers[module]
}

// Seal prevents additional handlers from being added to the router
func (r *router) Seal() {
	r.sealed = true
}
