package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// Route implements sdk.Msg
func (m MsgSignTransferOwnership) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m MsgSignTransferOwnership) Type() string {
	return "SignTransferOwnership"
}

// GetSignBytes  implements sdk.Msg
func (m MsgSignTransferOwnership) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners implements sdk.Msg
func (m MsgSignTransferOwnership) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}

// ValidateBasic implements sdk.Msg
func (m MsgSignTransferOwnership) ValidateBasic() error {
	if m.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if m.NewOwner == "" {
		return fmt.Errorf("missing new owner address")
	}

	return nil
}

// NewMsgSignTransferOwnership is the constructor for MsgSignTransferOwnership
func NewMsgSignTransferOwnership(sender sdk.AccAddress, newOwner common.Address) *MsgSignTransferOwnership {
	return &MsgSignTransferOwnership{Sender: sender, NewOwner: newOwner.Hex()}
}
