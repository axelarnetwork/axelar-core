package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

var _ sdk.Msg = &RotateKeyRequest{}

// NewRotateKeyRequest constructor for RotateKeyRequest
func NewRotateKeyRequest(sender sdk.AccAddress, chain nexus.ChainName, keyID exported.KeyID) *RotateKeyRequest {
	return &RotateKeyRequest{
		Sender: sender.String(),
		Chain:  chain,
		KeyID:  keyID,
	}
}

// ValidateBasic implements the sdk.Msg interface.
func (m RotateKeyRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	if err := m.KeyID.ValidateBasic(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	return nil
}
