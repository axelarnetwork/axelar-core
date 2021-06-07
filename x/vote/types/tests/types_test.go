package tests

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
)

func TestSdkInt_Marshaling(t *testing.T) {
	i := sdk.NewInt(75)
	cdc := testutils.MakeEncodingConfig().Amino

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
	encCfg := testutils.MakeEncodingConfig()
	cdc := encCfg.Marshaler

	data := tofnd.MessageOut_KeygenResult{KeygenResultData: &tofnd.MessageOut_KeygenResult_Pubkey{Pubkey: []byte("a public key")}}
	vote := types.NewTalliedVote(23, &data)

	bz := cdc.MustMarshalBinaryLengthPrefixed(&vote)
	var actual types.TalliedVote
	cdc.MustUnmarshalBinaryLengthPrefixed(bz, &actual)

	assert.Equal(t, vote, actual)

	bz = cdc.MustMarshalJSON(&vote)
	var actual2 types.TalliedVote
	cdc.MustUnmarshalJSON(bz, &actual2)

	assert.Equal(t, vote.Tally, actual2.Tally)
	assert.Equal(t, vote.Data.GetCachedValue(), actual2.Data.GetCachedValue())
}

func TestPoll_TallyNewVote(t *testing.T) {
	poll := types.Poll{
		Meta:                     exported.NewPollMeta("test", "test"),
		ValidatorSnapshotCounter: 0,
		Votes:                    []types.TalliedVote{types.NewTalliedVote(23, &gogoprototypes.BytesValue{Value: []byte("a public key")})},
		Result:                   nil,
	}

	vote := &poll.Votes[0]
	vote.Tally = vote.Tally.Add(sdk.NewInt(17))

	assert.Equal(t, sdk.NewInt(40), vote.Tally)
	assert.Equal(t, sdk.NewInt(40), poll.Votes[0].Tally)
}
