package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewAssignKeyRequest constructor for AssignKeyRequest
func NewAssignKeyRequest(sender sdk.AccAddress, chain string, keyID string, keyRole exported.KeyRole) *AssignKeyRequest {
	return &AssignKeyRequest{
		Sender:  sender,
		Chain:   chain,
		KeyID:   keyID,
		KeyRole: keyRole,
	}
}

// Route returns the route for this message
func (m AssignKeyRequest) Route() string { return RouterKey }

// Type returns the type of this message
func (m AssignKeyRequest) Type() string { return "AssignNextKey" }

// ValidateBasic performs a stateless validation of this message
func (m AssignKeyRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(ErrTss, "sender must be set")
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
func (m AssignKeyRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m AssignKeyRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
