package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewMsgSignPendingTransfers - MsgSignPendingTransfers constructor
func NewMsgSignPendingTransfers(sender sdk.AccAddress, fee btcutil.Amount) *MsgSignPendingTransfers {
	return &MsgSignPendingTransfers{Sender: sender.String(), Fee: int64(fee)}
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
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}
	if m.Fee <= 0 {
		return fmt.Errorf("fee must be a positive amount")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m MsgSignPendingTransfers) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m MsgSignPendingTransfers) GetSigners() []sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{from}
}
