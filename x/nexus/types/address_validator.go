package types

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// AddressValidators collects all registered address validators by module
type AddressValidators struct {
	validators map[string]exported.AddressValidator
	sealed     bool
}

// NewAddressValidators returns a new AddressValidators instance
func NewAddressValidators() *AddressValidators {
	return &AddressValidators{
		validators: make(map[string]exported.AddressValidator),
	}
}

// Seal prevents additional validators from being added
func (r *AddressValidators) Seal() {
	if r.sealed {
		panic("cannot seal address validator (validator already sealed)")
	}

	r.sealed = true
}

// IsSealed returns true if the validator is sealed
func (r *AddressValidators) IsSealed() bool {
	return r.sealed
}

// AddAddressValidator registers a validator for a given path
// panics if the validator is sealed, module is an empty string, or if the module has been registered already
func (r *AddressValidators) AddAddressValidator(module string, validator exported.AddressValidator) *AddressValidators {
	if r.sealed {
		panic("cannot add validator (validator sealed)")
	}

	if module == "" {
		panic("module name cannot be an empty string")
	}

	if r.hasAddressValidator(module) {
		panic(fmt.Sprintf("validator for module %s has already been registered", module))
	}

	r.validators[module] = validator
	return r
}

// hasAddressValidator returns true if a validator is registered for the given module
func (r *AddressValidators) hasAddressValidator(module string) bool {
	_, err := r.GetAddressValidator(module)
	return err == nil
}

// GetAddressValidator returns a validator for a given module
func (r *AddressValidators) GetAddressValidator(module string) (exported.AddressValidator, error) {
	validate, ok := r.validators[module]
	if !ok || validate == nil {
		return nil, fmt.Errorf("validator for module \"%s\" not registered", module)
	}
	return validate, nil
}
