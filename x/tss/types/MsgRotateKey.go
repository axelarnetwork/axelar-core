package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewMsgRotateKey constructor for MsgAssignNextKey
func NewMsgRotateKey(sender sdk.AccAddress, chain string, keyRole exported.KeyRole) *MsgRotateKey {
	return &MsgRotateKey{
		Sender:  sender.String(),
		Chain:   chain,
		KeyRole: keyRole,
	}
}

// Route returns the route for this message
func (m MsgRotateKey) Route() string {
	return RouterKey
}

// Type returns the type of this message
func (m MsgRotateKey) Type() string {
	return "RotateKey"
}

// ValidateBasic performs a stateless validation of this message
func (m MsgRotateKey) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "malformed sender address")
	}
	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	if err := m.KeyRole.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the bytes to sign for this message
func (m MsgRotateKey) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m MsgRotateKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.GetSender()}
}

// GetSender returns the sender object
func (m MsgRotateKey) GetSender() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return addr
}
