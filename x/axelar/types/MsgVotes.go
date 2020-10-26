package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/broadcast/exported"
)

var _ exported.ValidatorMsg = &MsgBatchVote{}

type MsgBatchVote struct {
	Votes  []bool
	Sender sdk.AccAddress
}

func (msg *MsgBatchVote) SetSender(sender sdk.AccAddress) {
	msg.Sender = sender
}

func (msg MsgBatchVote) Route() string {
	return RouterKey
}

func (msg MsgBatchVote) Type() string {
	return "BatchVote"
}

func (msg MsgBatchVote) ValidateBasic() error {
	if msg.Votes == nil {
		return sdkerrors.Wrap(ErrInvalidVotes, "votes must not be nil")
	}
	return nil
}

func (msg MsgBatchVote) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgBatchVote) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

func NewMsgBatchVote(votes []bool) *MsgBatchVote {
	return &MsgBatchVote{Votes: votes}
}
