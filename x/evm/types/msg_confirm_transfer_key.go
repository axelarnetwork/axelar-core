package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/axelarnetwork/axelar-core/utils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// NewConfirmTransferKeyRequest creates a message of type ConfirmTransferKeyRequest
func NewConfirmTransferKeyRequest(sender sdk.AccAddress, chain string, txID common.Hash) *ConfirmTransferKeyRequest {
	return &ConfirmTransferKeyRequest{
		Sender: sender,
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
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.Chain.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ConfirmTransferKeyRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m ConfirmTransferKeyRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
