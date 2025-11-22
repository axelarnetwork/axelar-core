package types

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewProxiedValidator is the constructor of ProxiedValidator
func NewProxiedValidator(validator sdk.ValAddress, proxy sdk.AccAddress, active bool) ProxiedValidator {
	return ProxiedValidator{
		Validator: validator,
		Proxy:     proxy,
		Active:    active,
	}
}

// Validate returns an error if the validator proxy is not valid; nil otherwise
func (m ProxiedValidator) Validate() error {
	if err := sdk.VerifyAddressFormat(m.Validator); err != nil {
		return errorsmod.Wrap(err, "invalid validator")
	}

	if err := sdk.VerifyAddressFormat(m.Proxy); err != nil {
		return errorsmod.Wrap(err, "invalid proxy")
	}

	if m.Validator.Equals(m.Proxy) {
		return errors.New("validator cannot be the same as proxy")
	}

	return nil
}
