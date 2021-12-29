package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
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
		Name:      utils.NormalizeString(name),
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
	if err := utils.ValidateString(m.Name, utils.DefaultDelimiter); err != nil {
		return sdkerrors.Wrap(err, "invalid chain name")
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
