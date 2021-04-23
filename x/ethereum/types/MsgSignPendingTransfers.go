package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgSignPendingTransfers - MsgSignPendingTransfers constructor
func NewMsgSignPendingTransfers(sender sdk.AccAddress) *MsgSignPendingTransfers {
	return &MsgSignPendingTransfers{Sender: sender}
}

// Route returns the route for this message
func (m MsgSignPendingTransfers) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m MsgSignPendingTransfers) Type() string {
	return "SignPendingTransfers"
}

// ValidateBasic executes a stateless message validation
func (m MsgSignPendingTransfers) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m MsgSignPendingTransfers) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m MsgSignPendingTransfers) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
