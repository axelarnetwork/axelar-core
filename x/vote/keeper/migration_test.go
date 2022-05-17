package keeper_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported2"
	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
)

func TestName(t *testing.T) {
	vote := exported2.Vote{}
	for i := 0; i < 10; i++ {
		result, err := codectypes.NewAnyWithValue(&evmtypes.Event{
			Chain: rand.Str(10),
		})
		assert.NoError(t, err)
		vote.Results = append(vote.Results, result)
	}
	packed, err := codectypes.NewAnyWithValue(&vote)
	assert.NoError(t, err)

	meta := exported2.PollMetadata{Result: packed}

	cdc := app.MakeEncodingConfig().Codec

	bz := cdc.MustMarshalLengthPrefixed(&meta)
	meta = exported2.PollMetadata{} // reset
	cdc.MustUnmarshalLengthPrefixed(bz, &meta)
	newPackedVote := keeper.MigrateVoteData(cdc, meta.Result, log.TestingLogger())

	newVote, ok := newPackedVote.GetCachedValue().(*exported.Vote)
	assert.True(t, ok)
	events, err := evmtypes.UnpackEvents(cdc, newVote.Result)
	assert.NoError(t, err)
	assert.Len(t, events.Events, 10)
}
