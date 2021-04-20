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
		Sender:  sender,
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
	if m.Sender == nil || len(m.Sender) != sdk.AddrLen {
		return sdkerrors.Wrap(ErrTss, "sender must be set")
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
	return []sdk.AccAddress{m.Sender}
}
