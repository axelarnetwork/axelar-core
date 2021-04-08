package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

// MsgSignDeployToken represents the message to sign a deploy token command for AxelarGateway
type MsgSignTransferOwnership struct {
	Sender   sdk.AccAddress
	NewOwner string
}

// Route implements sdk.Msg
func (msg MsgSignTransferOwnership) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (msg MsgSignTransferOwnership) Type() string {
	return "SignTransferOwnership"
}

// GetSignBytes  implements sdk.Msg
func (msg MsgSignTransferOwnership) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgSignTransferOwnership) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// ValidateBasic implements sdk.Msg
func (msg MsgSignTransferOwnership) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.NewOwner == "" {
		return fmt.Errorf("missing new owner address")
	}

	return nil
}

// NewMsgSignTransferOwnership is the constructor for MsgSignTransferOwnership
func NewMsgSignTransferOwnership(sender sdk.AccAddress, newOwner common.Address) sdk.Msg {
	return MsgSignTransferOwnership{Sender: sender, NewOwner: newOwner.Hex()}
}
