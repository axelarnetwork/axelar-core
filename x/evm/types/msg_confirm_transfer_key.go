package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewConfirmTransferKeyRequest creates a message of type ConfirmTransferKeyRequest
func NewConfirmTransferKeyRequest(sender sdk.AccAddress, chain string, txID common.Hash) *ConfirmTransferKeyRequest {
	return &ConfirmTransferKeyRequest{
		Sender: sender.String(),
		Chain:  nexus.ChainName(utils.NormalizeString(chain)),
		TxID:   Hash(txID),
	}
}

// Route implements sdk.Msg
func (m ConfirmTransferKeyRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ConfirmTransferKeyRequest) Type() string {
	return "ConfirmTransferKey"
}

// ValidateBasic implements sdk.Msg
func (m ConfirmTransferKeyRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid chain")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ConfirmTransferKeyRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// NewForceConfirmTransferKeyRequest creates a governance-only message to force confirm key transfer
func NewForceConfirmTransferKeyRequest(sender sdk.AccAddress, chain string) *ForceConfirmTransferKeyRequest {
	return &ForceConfirmTransferKeyRequest{
		Sender: sender.String(),
		Chain:  nexus.ChainName(utils.NormalizeString(chain)),
	}
}

// ValidateBasic implements sdk.Msg
func (m ForceConfirmTransferKeyRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, errorsmod.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid chain")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ForceConfirmTransferKeyRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}
