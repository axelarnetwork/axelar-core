package keeper

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	utilstestutils "github.com/axelarnetwork/axelar-core/utils/testutils"
	snapshottestutils "github.com/axelarnetwork/axelar-core/x/snapshot/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
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
		k.SetParams(ctx, types.DefaultParams())
		handler = GetMigrationHandler(k)
	})

	whenPollsExistWith := func(pollCount int64) WhenStatement {
		return When("some polls exist", func() {
			for i := 0; i < int(pollCount); i++ {
				pollMeta := exported.PollMetadata{
					ID:       exported.PollID(rand.PosI64()),
					State:    rand.Of(exported.Completed, exported.Failed, exported.Pending),
					Snapshot: snapshottestutils.Snapshot(uint64(rand.I64Between(1, 10)), utilstestutils.RandThreshold()),
				}

				switch pollMeta.State {
				case exported.Completed:
					data := gogoprototypes.StringValue{Value: rand.Str(10)}
					d, err := codectypes.NewAnyWithValue(&data)
					if err != nil {
						panic(err)
					}

					pollMeta.Result = d
					pollMeta.CompletedAt = 0
				case exported.Pending:
					k.GetPollQueue(ctx).Enqueue(pollPrefix.AppendStr(pollMeta.ID.String()), &pollMeta)
				}

				k.setPollMetadata(ctx, pollMeta)

				for _, voter := range pollMeta.Snapshot.Participants {
					k.getKVStore(ctx).SetRaw(voterPrefix.AppendStr(pollMeta.ID.String()).AppendStr(voter.Address.String()), []byte{0x01})
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

	givenTheMigrationHandler.
		When("EndBlockerLimit param is not set", func() {
			k.paramSpace.Set(ctx, types.KeyEndBlockerLimit, int64(0))
		}).
		Then("should set EndBlockerLimit param", func(t *testing.T) {
			assert.Zero(t, k.GetParams(ctx).EndBlockerLimit)

			err := handler(ctx)
			assert.NoError(t, err)

			assert.Equal(t, types.DefaultParams().EndBlockerLimit, k.GetParams(ctx).EndBlockerLimit)
		}).
		Run(t)
}
