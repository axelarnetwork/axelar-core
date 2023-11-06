package types

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// AddressValidator implements a AddressValidator based on module name.
type AddressValidator interface {
	AddAddressValidator(module string, validator exported.AddressValidator) AddressValidator
	HasAddressValidator(module string) bool
	GetAddressValidator(module string) exported.AddressValidator
	Seal()
}

var _ AddressValidator = (*addressValidator)(nil)

type addressValidator struct {
	validators map[string]exported.AddressValidator
	sealed     bool
}

// NewAddressValidator creates a new AddressValidator interface instance
func NewAddressValidator() AddressValidator {
	return &addressValidator{
		validators: make(map[string]exported.AddressValidator),
	}
}

// Seal prevents additional validators from being added
func (r *addressValidator) Seal() {
	r.sealed = true
}

// AddAddressValidator registers a validator for a given path
// panics if the validator is sealed, module is an empty string, or if the module has been registered already
func (r *addressValidator) AddAddressValidator(module string, validator exported.AddressValidator) AddressValidator {
	if r.sealed {
		panic("cannot add validator (validator sealed)")
	}

	if module == "" {
		panic("module name cannot be an empty string")
	}

	if r.HasAddressValidator(module) {
		panic(fmt.Sprintf("validator for module %s has already been registered", module))
	}

	r.validators[module] = validator
	return r
}

// HasAddressValidator returns true if a validator is registered for the given module
func (r *addressValidator) HasAddressValidator(module string) bool {
	return r.validators[module] != nil
}

// GetAddressValidator returns a validator for a given module
func (r *addressValidator) GetAddressValidator(module string) exported.AddressValidator {
	if !r.HasAddressValidator(module) {
		panic(fmt.Sprintf("validator for module \"%s\" not registered", module))
	}

	return r.validators[module]
}
