package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewMsgAssignNextKey constructor for MsgAssignNextKey
func NewMsgAssignNextKey(sender sdk.AccAddress, chain string, keyID string, keyRole exported.KeyRole) *MsgAssignNextKey {
	return &MsgAssignNextKey{
		Sender:  sender.String(),
		Chain:   chain,
		KeyID:   keyID,
		KeyRole: keyRole,
	}
}

// Route returns the route for this message
func (m MsgAssignNextKey) Route() string { return RouterKey }

// Type returns the type of this message
func (m MsgAssignNextKey) Type() string { return "AssignNextKey" }

// ValidateBasic performs a stateless validation of this message
func (m MsgAssignNextKey) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "malformed sender address")
	}

	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	if m.KeyID == "" {
		return fmt.Errorf("missing key ID")
	}

	if err := m.KeyRole.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the bytes to sign for this message
func (m MsgAssignNextKey) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m MsgAssignNextKey) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.GetSender()}
}

// GetSender returns the sender object
func (m MsgAssignNextKey) GetSender() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return addr
}
