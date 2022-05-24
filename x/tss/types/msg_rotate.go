package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewRotateKeyRequest constructor for RotateKeyRequest
func NewRotateKeyRequest(sender sdk.AccAddress, chain string, keyRole exported.KeyRole, keyID string) *RotateKeyRequest {
	return &RotateKeyRequest{
		Sender:  sender,
		Chain:   nexus.ChainName(utils.NormalizeString(chain)),
		KeyRole: keyRole,
		KeyID:   exported.KeyID(keyID),
	}
}

// Route returns the route for this message
func (m RotateKeyRequest) Route() string {
	return RouterKey
}

// Type returns the type of this message
func (m RotateKeyRequest) Type() string {
	return "RotateKey"
}

// ValidateBasic performs a stateless validation of this message
func (m RotateKeyRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(ErrTss, "sender must be set")
	}
	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if err := m.KeyRole.Validate(); err != nil {
		return err
	}

	if err := m.KeyID.Validate(); err != nil {
		return err
	}

	return nil
}

// GetSignBytes returns the bytes to sign for this message
func (m RotateKeyRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m RotateKeyRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
