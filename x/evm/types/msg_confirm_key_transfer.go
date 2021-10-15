package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
)

// NewConfirmKeyTransferRequest creates a message of type ConfirmTransferKeyRequest
func NewConfirmKeyTransferRequest(sender sdk.AccAddress, chain string, txID common.Hash, transferType KeyTransferType, keyID string) *ConfirmKeyTransferRequest {
	return &ConfirmKeyTransferRequest{
		Sender:       sender,
		Chain:        chain,
		TxID:         Hash(txID),
		TransferType: transferType,
		KeyID:        tss.KeyID(keyID),
	}
}

// Route implements sdk.Msg
func (m ConfirmKeyTransferRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ConfirmKeyTransferRequest) Type() string {
	return "ConfirmTransferKey"
}

// ValidateBasic implements sdk.Msg
func (m ConfirmKeyTransferRequest) ValidateBasic() error {
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
func (m ConfirmKeyTransferRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m ConfirmKeyTransferRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
