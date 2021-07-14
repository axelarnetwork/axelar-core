package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteConfirmChainRequest creates a message of type ConfirmTokenRequest
func NewVoteConfirmChainRequest(
	sender sdk.AccAddress,
	name string,
	key vote.PollKey,
	confirmed bool) *VoteConfirmChainRequest {
	return &VoteConfirmChainRequest{
		Sender:    sender,
		Name:      name,
		PollKey:   key,
		Confirmed: confirmed,
	}
}

// Route returns the route for this message
func (m VoteConfirmChainRequest) Route() string {
	return RouterKey
}

// Type returns the type of the message
func (m VoteConfirmChainRequest) Type() string {
	return "VoteConfirmChain"
}

// ValidateBasic executes a stateless message validation
func (m VoteConfirmChainRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}
	if m.Name == "" {
		return fmt.Errorf("missing chain")
	}

	return m.PollKey.Validate()
}

// GetSignBytes returns the message bytes that need to be signed
func (m VoteConfirmChainRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the set of signers for this message
func (m VoteConfirmChainRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
