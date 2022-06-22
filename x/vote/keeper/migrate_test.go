package keeper

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestGetMigrationHandler(t *testing.T) {
	var (
		ctx     sdk.Context
		k       Keeper
		handler func(ctx sdk.Context) error
	)

	repeats := 20
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

				pollID := exported.PollID(rand.PosI64())
				poll := k.newPollStore(ctx, pollID)

				pollMeta := exported.PollMetadata{
					ID:    pollID,
					State: rand.Of(exported.Completed, exported.Failed, exported.Pending),
					Voters: slices.Map(voters, func(v sdk.ValAddress) exported.Voter {
						return exported.Voter{Validator: v, VotingPower: rand.I64Between(10, 100)}
					}),
				}

				switch pollMeta.State {
				case exported.Completed:
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
				case exported.Pending:
					poll.EnqueuePoll(pollMeta)
				}

				poll.SetMetadata(pollMeta)

				for _, voter := range voters {
					poll.KVStore.SetRaw(voterPrefix.AppendStr(poll.id.String()).AppendStr(voter.String()), []byte{0x01})
				}
			}
		})
	}

	givenTheMigrationHandler.
		When2(whenPollsExistWith(pollCount)).
		Then("should delete all polls", func(t *testing.T) {
			err := handler(ctx)
			assert.NoError(t, err)

			iter := k.getKVStore(ctx).Iterator(pollPrefix)
			defer utils.CloseLogError(iter, k.Logger(ctx))
			assert.False(t, iter.Valid())
		}).
		Run(t, repeats)

	givenTheMigrationHandler.
		When2(whenPollsExistWith(pollCount)).
		Then("should empty poll queue", func(t *testing.T) {
			err := handler(ctx)
			assert.NoError(t, err)

			assert.True(t, k.GetPollQueue(ctx).IsEmpty())
		}).
		Run(t, repeats)
}
