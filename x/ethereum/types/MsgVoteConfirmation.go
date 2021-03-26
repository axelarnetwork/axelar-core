package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// MsgVoteConfirmation represents a message to that votes on a tx confirmation
type MsgVoteConfirmation struct {
	Sender    sdk.AccAddress
	PollMeta  exported.PollMeta
	TxID      common.Hash
	Confirmed bool
}

// Poll returns the poll this message votes on
func (msg MsgVoteConfirmation) Poll() exported.PollMeta {
	return msg.PollMeta
}

// Data returns the data this message is voting for
func (msg *MsgVoteConfirmation) Data() exported.VotingData {
	return msg.Confirmed
}

// Route returns the route for this message
func (msg MsgVoteConfirmation) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgVoteConfirmation) Type() string {
	return "VoteConfirmation"
}

// ValidateBasic executes a stateless message validation
func (msg MsgVoteConfirmation) ValidateBasic() error {
	if msg.Sender == nil {
		return fmt.Errorf("missing sender")
	}
	return msg.PollMeta.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (msg MsgVoteConfirmation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgVoteConfirmation) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
