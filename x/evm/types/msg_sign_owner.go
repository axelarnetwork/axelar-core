package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// NewSignTransferOwnershipRequest is the constructor for SignTransferOwnershipRequest
func NewSignTransferOwnershipRequest(sender sdk.AccAddress, newOwner common.Address) *SignTransferOwnershipRequest {
	return &SignTransferOwnershipRequest{Sender: sender, NewOwner: Address(newOwner)}
}

// Route implements sdk.Msg
func (m SignTransferOwnershipRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m SignTransferOwnershipRequest) Type() string {
	return "SignTransferOwnership"
}

// GetSignBytes  implements sdk.Msg
func (m SignTransferOwnershipRequest) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements sdk.Msg
func (m SignTransferOwnershipRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// ValidateBasic implements sdk.Msg
func (m SignTransferOwnershipRequest) ValidateBasic() error {
	if m.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}

	return nil
}
