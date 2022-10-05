package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/funcs"
)

var _ codectypes.UnpackInterfacesMessage = VoteRequest{}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m VoteRequest) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	return unpacker.UnpackAny(m.Vote, &data)
}

// NewVoteRequest creates a message of type VoteMsgRequest
func NewVoteRequest(sender sdk.AccAddress, id vote.PollID, vote codec.ProtoMarshaler) *VoteRequest {
	return &VoteRequest{
		Sender: sender,
		PollID: id,
		Vote:   funcs.Must(codectypes.NewAnyWithValue(vote)),
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

	if m.Vote == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "vote must not be nil")
	}

	vote := m.Vote.GetCachedValue()
	if vote == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "vote request contains no vote")
	}

	v, ok := vote.(utils.ValidatedProtoMarshaler)
	if !ok {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "vote request contains invalid vote")
	}

	if err := v.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
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
