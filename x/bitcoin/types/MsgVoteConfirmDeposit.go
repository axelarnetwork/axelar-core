package types

import (
	"fmt"

	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"

	exported2 "github.com/axelarnetwork/axelar-core/x/broadcast/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var _ exported2.MsgWithSenderSetter = &MsgVoteConfirmOutpoint{}

// MsgVoteConfirmOutpoint represents a message to that votes on an outpoint
type MsgVoteConfirmOutpoint struct {
	Sender    sdk.AccAddress
	PollMeta  exported.PollMeta
	Outpoint  wire.OutPoint
	Confirmed bool
}

// Route returns the route for this message
func (msg MsgVoteConfirmOutpoint) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgVoteConfirmOutpoint) Type() string {
	return "VoteConfirmDeposit"
}

// ValidateBasic executes a stateless message validation
func (msg MsgVoteConfirmOutpoint) ValidateBasic() error {
	if msg.Sender == nil {
		return fmt.Errorf("missing sender")
	}
	return msg.PollMeta.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (msg MsgVoteConfirmOutpoint) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgVoteConfirmOutpoint) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// SetSender sets the message sender
// Deprecated
func (msg *MsgVoteConfirmOutpoint) SetSender(address sdk.AccAddress) {
	msg.Sender = address
}
