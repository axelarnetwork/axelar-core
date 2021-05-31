package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Route returns the route for this message
func (m VoteConfirmDepositRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m VoteConfirmDepositRequest) Type() string {
	return "VoteConfirmDeposit"
}

// ValidateBasic executes a stateless message validation
func (m VoteConfirmDepositRequest) ValidateBasic() error {
	if m.Sender == nil || len(m.Sender) != sdk.AddrLen {
		return fmt.Errorf("missing sender")
	}

	return m.Poll.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (m VoteConfirmDepositRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m VoteConfirmDepositRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
