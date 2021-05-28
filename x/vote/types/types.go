package types

import (
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ types.UnpackInterfacesMessage = TalliedVote{}
var _ types.UnpackInterfacesMessage = Poll{}

func NewTalliedVote(tally int64, data exported.VotingData) TalliedVote {
	d, err := codectypes.NewAnyWithValue(data)
	if err != nil {
		panic(err)
	}

	return TalliedVote{
		Tally: sdk.NewInt(tally),
		Data:  d,
	}
}

func NewPoll(meta exported.PollMeta, validatorSnapshotCounter int64) Poll {
	return Poll{
		Meta:                     meta,
		ValidatorSnapshotCounter: validatorSnapshotCounter,
	}
}

func (m Poll) GetResult() interface{} {
	if m.Result == nil {
		return nil
	}

	return m.Result.GetCachedValue()
}

func (m Poll) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for i := range m.Votes {
		if err := m.Votes[i].UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}

	if m.Result == nil {
		return nil
	}

	var result exported.VotingData
	return unpacker.UnpackAny(m.Result, &result)
}

func (m TalliedVote) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var data exported.VotingData
	return unpacker.UnpackAny(m.Data, &data)
}
