package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

//go:generate moq -out ./mock/types.go -pkg mock . VoteRouter

var _ codectypes.UnpackInterfacesMessage = TalliedVote{}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m TalliedVote) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	return unpacker.UnpackAny(m.Data, &data)
}

// NewTalliedVote is the constructor for TalliedVote
func NewTalliedVote(pollID exported.PollID, data codec.ProtoMarshaler) TalliedVote {
	d, err := codectypes.NewAnyWithValue(data)
	if err != nil {
		panic(err)
	}

	return TalliedVote{
		PollID: pollID,
		Tally:  sdk.ZeroUint(),
		Data:   d,
	}
}

// TallyVote adds the given voting power to the tallied vote
func (m *TalliedVote) TallyVote(voter sdk.ValAddress, votingPower sdk.Uint, isLate bool) {
	if voter == nil {
		panic("voter cannot be nil")
	}

	if m.IsVoterLate == nil {
		m.IsVoterLate = make(map[string]bool)
	}

	m.IsVoterLate[voter.String()] = isLate
	m.Tally = m.Tally.Add(votingPower)
}
