package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
)

// NewPool is the constructor of Pool
func NewPool(name string) Pool {
	return Pool{
		Name:    utils.NormalizeString(name),
		Rewards: []Pool_Reward{},
	}
}

// ValidateBasic returns an error if the Pool is not valid; nil otherwise
func (m Pool) ValidateBasic() error {
	if err := utils.ValidateString(m.Name); err != nil {
		return sdkerrors.Wrap(err, "invalid name")
	}

	validatorSeen := make(map[string]bool)
	for _, reward := range m.Rewards {
		validatorAddr := reward.Validator.String()
		if validatorSeen[validatorAddr] {
			return fmt.Errorf("duplicate validator %s found in pool %s", validatorAddr, m.Name)
		}

		if err := sdk.VerifyAddressFormat(reward.Validator); err != nil {
			return fmt.Errorf("invalid validator %s found in pool %s", validatorAddr, m.Name)
		}

		if reward.Coins == nil || reward.Coins.Empty() {
			return fmt.Errorf("empty rewards found for validator %s in pool %s", validatorAddr, m.Name)
		}

		if err := reward.Coins.Validate(); err != nil {
			return sdkerrors.Wrapf(err, "invalid rewards for validator %s found in pool %s", validatorAddr, m.Name)
		}

		validatorSeen[validatorAddr] = true
	}

	return nil
}

// ValidateBasic returns an error if the Refund is not valid; nil otherwise
func (m Refund) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Payer); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "payer").Error())
	}

	if err := m.Fees.Validate(); err != nil {
		return err
	}

	return nil
}
