package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// NewVoteRequest creates a message of type VoteMsgRequest
func NewVoteRequest(sender sdk.AccAddress, id vote.PollID, vote vote.Vote) *VoteRequest {
	return &VoteRequest{
		Sender: sender,
		PollID: id,
		Vote:   vote,
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
