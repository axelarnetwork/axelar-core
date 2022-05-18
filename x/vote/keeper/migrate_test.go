package keeper

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	exported0_17 "github.com/axelarnetwork/axelar-core/x/vote/exported-0.17"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestGetMigrationHandler(t *testing.T) {
	var (
		ctx     sdk.Context
		k       Keeper
		handler func(ctx sdk.Context) error
	)

	repeats := 1
	pollCount := rand.I64Between(100, 200)

	givenTheMigrationHandler := Given("the migration handler", func() {
		ctx, k, _, _, _ = setup()
		handler = GetMigrationHandler(k)
	})

	whenPollsExistWith := func(pollCount int64) WhenStatement {
		return When("some polls exist", func() {
			for i := 0; i < int(pollCount); i++ {
				voterCount := rand.I64Between(1, 10)
				voters := make([]sdk.ValAddress, voterCount)

				pollKey := exported.NewPollKey(rand.Str(5), rand.HexStr(64))
				poll := k.newPollStore(ctx, pollKey)

				pollMeta := exported.PollMetadata{
					Key:   pollKey,
					State: rand.Of(exported.Completed, exported.Failed, exported.Pending),
					Voters: slices.Map(voters, func(v sdk.ValAddress) exported.Voter {
						return exported.Voter{Validator: v, VotingPower: rand.I64Between(10, 100)}
					}),
				}

				if pollMeta.State == exported.Completed {
					data := gogoprototypes.StringValue{Value: rand.Str(10)}
					d, err := codectypes.NewAnyWithValue(&data)
					if err != nil {
						panic(err)
					}

					vote := exported.Vote{Result: d}
					d, err = codectypes.NewAnyWithValue(&vote)
					if err != nil {
						panic(err)
					}

					pollMeta.Result = d
					pollMeta.CompletedAt = 0
				}

				poll.SetMetadata(pollMeta)

				for _, voter := range voters {
					poll.KVStore.SetRaw(voterPrefix.AppendStr(poll.key.String()).AppendStr(voter.String()), []byte{0x01})
				}
			}
		})
	}

	givenTheMigrationHandler.
		When2(whenPollsExistWith(pollCount)).
		Then("should delete all pending polls", func(t *testing.T) {
			err := handler(ctx)
			assert.NoError(t, err)

			var pendingPollMetadatas []exported.PollMetadata

			iter := k.getKVStore(ctx).Iterator(pollPrefix)
			defer utils.CloseLogError(iter, k.Logger(ctx))

			for ; iter.Valid(); iter.Next() {
				var pollMetadata exported.PollMetadata
				iter.UnmarshalValue(&pollMetadata)

				if !pollMetadata.Is(exported.Pending) {
					continue
				}

				pendingPollMetadatas = append(pendingPollMetadatas, pollMetadata)
			}

			assert.Less(t, len(k.getNonPendingPollMetadatas(ctx)), int(pollCount))
			assert.Empty(t, pendingPollMetadatas)
		}).
		Run(t, repeats)

	givenTheMigrationHandler.
		When2(whenPollsExistWith(pollCount)).
		Then("should migrate all completed polls", func(t *testing.T) {
			iter := k.getKVStore(ctx).Iterator(pollPrefix)
			for ; iter.Valid(); iter.Next() {
				var pollMetadata exported.PollMetadata
				iter.UnmarshalValue(&pollMetadata)

				if !pollMetadata.Is(exported.Completed) {
					continue
				}

				assert.Zero(t, pollMetadata.CompletedAt)

				poll := k.newPollStore(ctx, pollMetadata.Key)
				for _, voter := range pollMetadata.Voters {
					assert.Panics(t, func() { poll.HasVotedLate(voter.Validator) })
				}
			}
			utils.CloseLogError(iter, k.Logger(ctx))

			err := handler(ctx)
			assert.NoError(t, err)

			iter = k.getKVStore(ctx).Iterator(pollPrefix)
			defer utils.CloseLogError(iter, k.Logger(ctx))

			for ; iter.Valid(); iter.Next() {
				var pollMetadata exported.PollMetadata
				iter.UnmarshalValue(&pollMetadata)

				if !pollMetadata.Is(exported.Completed) {
					continue
				}

				assert.NotZero(t, pollMetadata.CompletedAt)

				poll := k.newPollStore(ctx, pollMetadata.Key)
				for _, voter := range pollMetadata.Voters {
					assert.False(t, poll.HasVotedLate(voter.Validator))
				}
			}
		}).
		Run(t, repeats)
}

func TestMigrateVotes(t *testing.T) {
	vote := exported0_17.Vote{}
	for i := 0; i < 10; i++ {
		result, err := codectypes.NewAnyWithValue(&evmtypes.Event{
			Chain: nexus.ChainName(rand.Str(10)),
		})
		assert.NoError(t, err)
		vote.Results = append(vote.Results, result)
	}
	packed, err := codectypes.NewAnyWithValue(&vote)
	assert.NoError(t, err)

	meta := exported0_17.PollMetadata{Result: packed}

	encodingConfig := appParams.MakeEncodingConfig()
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	cdc := encodingConfig.Codec

	bz := cdc.MustMarshalLengthPrefixed(&meta)
	meta = exported0_17.PollMetadata{} // reset
	cdc.MustUnmarshalLengthPrefixed(bz, &meta)
	newPackedVote := MigrateVoteData(cdc, nexus.ChainName(rand.Str(5)), meta.Result, log.TestingLogger())

	newVote, ok := newPackedVote.GetCachedValue().(*exported.Vote)
	assert.True(t, ok)
	events, err := evmtypes.UnpackEvents(cdc, newVote.Result)
	assert.NoError(t, err)
	assert.Len(t, events.Events, 10)
}
