package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewRotateKeyRequest constructor for RotateKeyRequest
func NewRotateKeyRequest(sender sdk.AccAddress, chain string, keyRole exported.KeyRole, keyID string) *RotateKeyRequest {
	return &RotateKeyRequest{
		Sender:  sender.String(),
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
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(ErrTss, "sender must be set")
	}
	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid chain")
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
