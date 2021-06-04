package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// NewConfirmDepositRequest creates a message of type ConfirmDepositRequest
func NewConfirmDepositRequest(sender sdk.AccAddress, chain string, txID common.Hash, amount sdk.Uint, burnerAddr common.Address) *ConfirmDepositRequest {
	return &ConfirmDepositRequest{
		Sender:        sender,
		Chain:         chain,
		TxID:          Hash(txID),
		Amount:        amount,
		BurnerAddress: Address(burnerAddr),
	}
}

// Route implements sdk.Msg
func (m ConfirmDepositRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ConfirmDepositRequest) Type() string {
	return "ConfirmERC20Deposit"
}

// ValidateBasic implements sdk.Msg
func (m ConfirmDepositRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ConfirmDepositRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m ConfirmDepositRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
