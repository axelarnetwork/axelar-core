package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
)

// NewGovAccount performs a stateless check to ensure the Chain object has been initialized correctly
func NewGovAccount(addr sdk.AccAddress, role exported.Role) GovAccount {
	return GovAccount{Address: addr, Role: role}
}

// Validate performs a stateless check to ensure the Chain object has been initialized correctly
func (m GovAccount) Validate() error {
	if err := sdk.VerifyAddressFormat(m.Address); err != nil {
		return err
	}

	if err := m.Role.Validate(); err != nil {
		return err
	}

	return nil
}
