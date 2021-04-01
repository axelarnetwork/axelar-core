package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// MsgVoteConfirmToken represents a message that votes on a token deploy
type MsgVoteConfirmToken struct {
	Sender    sdk.AccAddress
	Poll      exported.PollMeta
	TxID      string
	Symbol    string
	Confirmed bool
}

// Route returns the route for this message
func (msg MsgVoteConfirmToken) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (msg MsgVoteConfirmToken) Type() string {
	return "VoteConfirmation"
}

// ValidateBasic executes a stateless message validation
func (msg MsgVoteConfirmToken) ValidateBasic() error {
	if msg.Sender == nil {
		return fmt.Errorf("missing sender")
	}
	if msg.TxID == "" {
		return fmt.Errorf("tx ID missing")
	}
	if msg.Symbol == "" {
		return fmt.Errorf("symbol missing")
	}
	return msg.Poll.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (msg MsgVoteConfirmToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (msg MsgVoteConfirmToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
