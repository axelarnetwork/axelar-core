package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteRequest creates a message of type VoteMsgRequest
func NewVoteRequest(sender sdk.AccAddress, pollKey vote.PollKey, vote vote.Vote, chain string) *VoteRequest {
	return &VoteRequest{
		Sender:  sender,
		PollKey: pollKey,
		Vote:    vote,
		Chain:   chain,
	}
}

// Route implements sdk.Msg
func (m VoteRequest) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (m VoteRequest) Type() string {
	return "Vote"
}

// ValidateBasic implements sdk.Msg
func (m VoteRequest) ValidateBasic() error {
	if err := sdk.VerifyAddressFormat(m.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, sdkerrors.Wrap(err, "sender").Error())
	}

	if err := m.PollKey.Validate(); err != nil {
		return sdkerrors.Wrap(err, "invalid poll key")
	}

	if err := utils.ValidateString(m.Chain); err != nil {
		return sdkerrors.Wrap(err, "invalid chain")
	}

	if m.Vote.Results == nil {
		return fmt.Errorf("missing vote results")
	}

	return nil
}

// GetSignBytes implements sdk.Msg
func (m VoteRequest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (m VoteRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{m.Sender}
}
