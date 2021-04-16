package types

import (
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewMsgVoteConfirmOutpoint - MsgVoteConfirmOutpoint constructor
func NewMsgVoteConfirmOutpoint(sender sdk.AccAddress, poll exported.PollMeta, outPoint wire.OutPoint, confirmed bool) *MsgVoteConfirmOutpoint {
	return &MsgVoteConfirmOutpoint{
		Sender:    sender.String(),
		Poll:      poll,
		OutPoint:  outPoint.String(),
		Confirmed: confirmed,
	}
}

// Route returns the route for this message
func (m MsgVoteConfirmOutpoint) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m MsgVoteConfirmOutpoint) Type() string {
	return "VoteConfirmDeposit"
}

// ValidateBasic executes a stateless message validation
func (m MsgVoteConfirmOutpoint) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid sender address (%s)", err)
	}

	if _, err := OutPointFromStr(m.OutPoint); err != nil {
		return sdkerrors.Wrap(err, "outpoint malformed")
	}
	return m.Poll.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (m MsgVoteConfirmOutpoint) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the set of signers for this message
func (m MsgVoteConfirmOutpoint) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.GetSender()}
}

// GetSender returns decoded sender Address
func (m MsgVoteConfirmOutpoint) GetSender() sdk.AccAddress {
	from, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err)
	}
	return from
}
