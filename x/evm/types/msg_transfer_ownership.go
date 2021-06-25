package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// NewConfirmTransferOwnershipRequest creates a message of type ConfirmTransferOwnershipRequest
func NewConfirmTransferOwnershipRequest(sender sdk.AccAddress, chain string, txID common.Hash, newOwnerAddr common.Address) *ConfirmTransferOwnershipRequest {
	return &ConfirmTransferOwnershipRequest{
		Sender:          sender,
		Chain:           chain,
		TxID:            Hash(txID),
		NewOwnerAddress: Address(newOwnerAddr),
	}
}

// Route implements sdk.Msg
func (m ConfirmTransferOwnershipRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m ConfirmTransferOwnershipRequest) Type() string {
	return "ConfirmTransferOwnershipRequest"
}

// ValidateBasic implements sdk.Msg
func (m ConfirmTransferOwnershipRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if m.Chain == "" {
		return fmt.Errorf("missing chain")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m ConfirmTransferOwnershipRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m ConfirmTransferOwnershipRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
