package types

import (
	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewUpdateGovernanceKeyRequest is the constructor for UpdateGovernanceKeyRequest
func NewUpdateGovernanceKeyRequest(sender sdk.AccAddress, threshold int, pubKeys ...cryptotypes.PubKey) *UpdateGovernanceKeyRequest {
	govKey := multisig.NewLegacyAminoPubKey(threshold, pubKeys)

	return &UpdateGovernanceKeyRequest{
		Sender:        sender.String(),
		GovernanceKey: *govKey,
	}
}

// Route returns the route for this message
func (m UpdateGovernanceKeyRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m UpdateGovernanceKeyRequest) Type() string {
	return "UpdateGovernanceKey"
}

// ValidateBasic executes a stateless message validation
func (m UpdateGovernanceKeyRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if m.GovernanceKey.Threshold <= 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "threshold k of n multisignature: k <= 0")
	}

	if uint32(len(m.GovernanceKey.GetPubKeys())) < m.GovernanceKey.Threshold {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "threshold k of n multisignature: len(pubKeys) < k")
	}

	return nil
}

// GetSignBytes returns the message bytes that need to be signed
func (m UpdateGovernanceKeyRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m UpdateGovernanceKeyRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if err := m.GovernanceKey.UnpackInterfaces(unpacker); err != nil {
		return err
	}
	return nil
}
