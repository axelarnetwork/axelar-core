package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetAmount retrieves the parsed amount from the TransferFeeRequest
func (m TransferFeeRequest) GetAmount() (sdk.Coin, error) {
	amount, err := sdk.ParseCoinNormalized(m.Amount)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(err, "invalid amount")
	}

	return amount, nil
}
