package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgSignPendingTransfers represents a message to trigger the signing of a consolidation transaction
type MsgSignPendingTransfers struct {
	Sender sdk.AccAddress
	Fee    btcutil.Amount
}

// NewMsgSignPendingTransfers - MsgSignPendingTransfers constructor
func NewMsgSignPendingTransfers(sender sdk.AccAddress, fee btcutil.Amount) sdk.Msg {
	return MsgSignPendingTransfers{Sender: sender, Fee: fee}
}

// Route returns the route for this message
func (msg MsgSignPendingTransfers) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSignPendingTransfers) Type() string {
	return "SignPendingTransfers"
}

// ValidateBasic executes a stateless message validation
func (msg MsgSignPendingTransfers) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.Fee <= 0 {
		return fmt.Errorf("fee must be a positive amount")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (msg MsgSignPendingTransfers) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgSignPendingTransfers) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
