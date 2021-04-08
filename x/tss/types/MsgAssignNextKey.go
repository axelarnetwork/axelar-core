package types

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgAssignNextKey represents a message to assign a new key
type MsgAssignNextKey struct {
	Sender  sdk.AccAddress
	Chain   string
	KeyID   string
	KeyRole exported.KeyRole
}

// NewMsgAssignNextKey constructor for MsgAssignNextKey
func NewMsgAssignNextKey(sender sdk.AccAddress, chain string, keyID string, keyRole exported.KeyRole) sdk.Msg {
	return MsgAssignNextKey{
		Sender:  sender,
		Chain:   chain,
		KeyID:   keyID,
		KeyRole: keyRole,
	}
}

// Route returns the route for this message
func (msg MsgAssignNextKey) Route() string { return RouterKey }

// Type returns the type of this message
func (msg MsgAssignNextKey) Type() string { return "AssignNextKey" }

// ValidateBasic performs a stateless validation of this message
func (msg MsgAssignNextKey) ValidateBasic() error {
	if msg.Sender == nil {
		return sdkerrors.ErrInvalidAddress
	}

	if msg.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	if msg.KeyID == "" {
		return fmt.Errorf("missing key ID")
	}

	if err := msg.KeyRole.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the bytes to sign for this message
func (msg MsgAssignNextKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgAssignNextKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
