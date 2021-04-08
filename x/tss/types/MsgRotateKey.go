package types

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgRotateKey represents a message to rotate a key
type MsgRotateKey struct {
	Sender  sdk.AccAddress
	Chain   string
	KeyRole exported.KeyRole
}

// NewMsgRotateKey constructor for MsgAssignNextKey
func NewMsgRotateKey(sender sdk.AccAddress, chain string, keyRole exported.KeyRole) sdk.Msg {
	return MsgRotateKey{
		Sender:  sender,
		Chain:   chain,
		KeyRole: keyRole,
	}
}

// Route returns the route for this message
func (msg MsgRotateKey) Route() string {
	return RouterKey
}

// Type returns the type of this message
func (msg MsgRotateKey) Type() string {
	return "RotateKey"
}

// ValidateBasic performs a stateless validation of this message
func (msg MsgRotateKey) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	if msg.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	if err := msg.KeyRole.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the bytes to sign for this message
func (msg MsgRotateKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgRotateKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
