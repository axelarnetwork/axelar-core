package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/axelar/exported"
)

// Ensure MsgTrackAddress implements sdk.Msg interface
var _ sdk.Msg = &MsgVerifyTx{}

type MsgVerifyTx struct {
	Sender sdk.AccAddress
	Tx     exported.ExternalTx
}

func NewMsgVerifyTx(sender sdk.AccAddress, tx exported.ExternalTx) MsgVerifyTx {
	return MsgVerifyTx{
		Sender: sender,
		Tx:     tx,
	}
}

func (msg MsgVerifyTx) Route() string {
	return RouterKey
}

func (msg MsgVerifyTx) Type() string {
	return "VerifyTx"
}

func (msg MsgVerifyTx) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender")
	}
	if msg.Tx.IsInvalid() {
		return sdkerrors.Wrap(ErrInvalidExternalTx, msg.Tx.String())
	}

	return nil
}

func (msg MsgVerifyTx) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgVerifyTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
