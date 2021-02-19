package types

import (
	"fmt"

	"github.com/btcsuite/btcutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgSign represents a message to trigger the signing of a consolidation transaction
type MsgSign struct {
	Sender sdk.AccAddress
	Fee    btcutil.Amount
}

// NewMsgSign - MsgSign constructor
func NewMsgSign(sender sdk.AccAddress, fee btcutil.Amount) sdk.Msg {
	return MsgSign{Sender: sender, Fee: fee}
}

// Route returns the route for this message
func (msg MsgSign) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgSign) Type() string {
	return "Sign"
}

// ValidateBasic executes a stateless message validation
func (msg MsgSign) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.Fee <= 0 {
		return fmt.Errorf("fee must be a positive amount")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (msg MsgSign) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgSign) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
