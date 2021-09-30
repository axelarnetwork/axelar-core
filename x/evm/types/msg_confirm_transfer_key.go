package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewConfirmTransferKeyRequest creates a message of type ConfirmTransferKeyRequest
func NewConfirmTransferKeyRequest(sender sdk.AccAddress, chain string, txID common.Hash, transferType TransferKeyType, keyID string) *ConfirmTransferKeyRequest {
	return &ConfirmTransferKeyRequest{
		Sender:       sender,
		Chain:        chain,
		TxID:         Hash(txID),
		TransferType: transferType,
		KeyID:        tss.KeyID(keyID),
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

	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	if err := m.TransferType.Validate(); err != nil {
		return err
	}

	if err := m.KeyID.Validate(); err != nil {
		return err
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
