package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// MsgVoteConfirmDeposit represents a message that votes on a deposit
type MsgVoteConfirmDeposit struct {
	Sender    sdk.AccAddress
	PollMeta  exported.PollMeta
	TxID      string
	BurnAddr  string
	Confirmed bool
}

// Poll returns the poll this message votes on
func (msg MsgVoteConfirmDeposit) Poll() exported.PollMeta {
	return msg.PollMeta
}

// Data returns the data this message is voting for
func (msg *MsgVoteConfirmDeposit) Data() exported.VotingData {
	return msg.Confirmed
}

// Route returns the route for this message
func (msg MsgVoteConfirmDeposit) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgVoteConfirmDeposit) Type() string {
	return "VoteConfirmDeposit"
}

// ValidateBasic executes a stateless message validation
func (msg MsgVoteConfirmDeposit) ValidateBasic() error {
	if msg.Sender == nil {
		return fmt.Errorf("missing sender")
	}
	if msg.TxID == "" {
		return fmt.Errorf("tx ID missing")
	}
	if msg.BurnAddr == "" {
		return fmt.Errorf("burn address missing")
	}
	return msg.PollMeta.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (msg MsgVoteConfirmDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgVoteConfirmDeposit) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
