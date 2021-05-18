package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// NewConfirmTokenRequest creates a message of type ConfirmTokenRequest
func NewConfirmTokenRequest(sender sdk.AccAddress, txID common.Hash, symbol string) *ConfirmTokenRequest {
	return &ConfirmTokenRequest{
		Sender: sender,
		TxID:   txID.Hex(),
		Symbol: symbol,
	}
}

// Route implements sdk.Msg
func (m ConfirmTokenRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ConfirmTokenRequest) Type() string {
	return "ConfirmERC20TokenDeploy"
}

// ValidateBasic implements sdk.Msg
func (m ConfirmTokenRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ConfirmTokenRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m ConfirmTokenRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
