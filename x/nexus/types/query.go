package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GetAmount retrieves the parsed amount from the TransferFeeRequest
func (m TransferFeeRequest) GetAmount() (sdk.Coin, error) {
	amount, err := sdk.ParseCoinNormalized(m.Amount)
	if err != nil {
		return sdk.Coin{}, sdkerrors.Wrap(err, "invalid amount")
	}

	return amount, nil
}
