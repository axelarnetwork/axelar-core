package tests

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

func TestSdkInt_Marshaling(t *testing.T) {
	i := sdk.NewInt(75)
	cdc := testutils.Codec()

	bz := cdc.MustMarshalBinaryLengthPrefixed(i)
	var unmarshaled sdk.Int
	cdc.MustUnmarshalBinaryLengthPrefixed(bz, &unmarshaled)

	assert.Equal(t, unmarshaled, i)

	bz = cdc.MustMarshalJSON(i)
	var unmarshaled2 sdk.Int
	cdc.MustUnmarshalJSON(bz, &unmarshaled2)

	assert.Equal(t, unmarshaled2, i)

}

func TestTalliedVote_Marshaling(t *testing.T) {
	vote := types.TalliedVote{
		Tally: sdk.NewInt(23),
		Data:  []byte("a public key"),
	}
	cdc := testutils.Codec()

	bz := cdc.MustMarshalBinaryLengthPrefixed(vote)
	var unmarshaled types.TalliedVote
	cdc.MustUnmarshalBinaryLengthPrefixed(bz, &unmarshaled)

	assert.Equal(t, unmarshaled, vote)

	bz = cdc.MustMarshalJSON(vote)
	var unmarshaled2 types.TalliedVote
	cdc.MustUnmarshalJSON(bz, &unmarshaled2)

	assert.Equal(t, unmarshaled2, vote)
}

func TestPoll_TallyNewVote(t *testing.T) {
	poll := types.Poll{
		Meta:                   exported.PollMeta{},
		ValidatorSnapshotRound: 0,
		Votes: []types.TalliedVote{{
			Tally: sdk.NewInt(23),
			Data:  []byte("a public key")},
		},
		Result: nil,
	}

	vote := &poll.Votes[0]
	vote.Tally = vote.Tally.Add(sdk.NewInt(17))

	assert.Equal(t, sdk.NewInt(40), vote.Tally)
	assert.Equal(t, sdk.NewInt(40), poll.Votes[0].Tally)
}
