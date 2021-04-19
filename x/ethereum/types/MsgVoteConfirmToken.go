package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Route returns the route for this message
func (m MsgVoteConfirmToken) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m MsgVoteConfirmToken) Type() string {
	return "VoteConfirmToken"
}

// ValidateBasic executes a stateless message validation
func (m MsgVoteConfirmToken) ValidateBasic() error {
	if m.Sender == nil {
		return fmt.Errorf("missing sender")
	}
	if m.TxID == "" {
		return fmt.Errorf("tx ID missing")
	}
	if m.Symbol == "" {
		return fmt.Errorf("symbol missing")
	}
	return m.Poll.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (m MsgVoteConfirmToken) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m MsgVoteConfirmToken) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
