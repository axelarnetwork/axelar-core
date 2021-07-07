package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var _ types.UnpackInterfacesMessage = TalliedVote{}
var _ types.UnpackInterfacesMessage = Poll{}

// NewTalliedVote is the constructor for TalliedVote
func NewTalliedVote(tally int64, data codec.ProtoMarshaler) TalliedVote {
	d, err := codectypes.NewAnyWithValue(data)
	if err != nil {
		panic(err)
	}

	return TalliedVote{
		Tally: sdk.NewInt(tally),
		Data:  d,
	}
}

// NewPoll is the constructor for Poll
func NewPoll(meta exported.PollMeta, validatorSnapshotCounter int64, expiresAt int64, threshold utils.Threshold) Poll {
	return Poll{
		Meta:                     meta,
		ValidatorSnapshotCounter: validatorSnapshotCounter,
		ExpiresAt:                expiresAt,
		VotingThreshold:          threshold,
	}
}

// HasExpired returns true if the poll has expired; otherwise, false
func (m Poll) HasExpired(ctx sdk.Context) bool {
	return m.ExpiresAt > 0 && ctx.BlockHeight() >= m.ExpiresAt
}

// GetResult returns the poll result
func (m Poll) GetResult() codec.ProtoMarshaler {
	if m.Result == nil {
		return nil
	}

	return m.Result.GetCachedValue().(codec.ProtoMarshaler)
}

// GetVotedShareCount returns the total voted share count in poll
func (m Poll) GetVotedShareCount() sdk.Int {
	result := sdk.ZeroInt()

	for _, talliedVote := range m.Votes {
		result = result.Add(talliedVote.Tally)
	}

	return result
}

// GetHighestTalliedVote returns the vote with the highest tally if any vote exists
func (m Poll) GetHighestTalliedVote() (bool, TalliedVote) {
	var result TalliedVote
	found := false

	for i, talliedVote := range m.Votes {
		if i == 0 || talliedVote.Tally.GT(result.Tally) {
			found = true
			result = talliedVote
		}
	}

	return found, result
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m Poll) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for i := range m.Votes {
		if err := m.Votes[i].UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}

	if m.Result == nil {
		return nil
	}

	var result codec.ProtoMarshaler
	return unpacker.UnpackAny(m.Result, &result)
}

// UnpackInterfaces implements UnpackInterfacesMessage
func (m TalliedVote) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data codec.ProtoMarshaler
	return unpacker.UnpackAny(m.Data, &data)
}
