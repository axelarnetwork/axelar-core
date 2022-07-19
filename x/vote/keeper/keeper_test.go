package keeper_test

import (
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	utilstestutils "github.com/axelarnetwork/axelar-core/utils/testutils"
	snapshottypes "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapshottestutils "github.com/axelarnetwork/axelar-core/x/snapshot/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/axelar-core/x/vote/types/mock"
	. "github.com/axelarnetwork/utils/test"
)

func setup() (sdk.Context, keeper.Keeper, *mock.SnapshotterMock, *mock.StakingKeeperMock, *mock.RewarderMock) {
	snapshotter := mock.SnapshotterMock{}
	staking := mock.StakingKeeperMock{}
	rewarder := mock.RewarderMock{}

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "vote")

	keeper := keeper.NewKeeper(
		encodingConfig.Codec,
		sdk.NewKVStoreKey(types.StoreKey),
		subspace,
		&snapshotter,
		&staking,
		&rewarder,
	)
	keeper.SetParams(ctx, types.DefaultParams())

	return ctx, keeper, &snapshotter, &staking, &rewarder
}

func TestKeeper(t *testing.T) {
	var (
		ctx         sdk.Context
		k           keeper.Keeper
		pollBuilder exported.PollBuilder
		snapshot    snapshottypes.Snapshot
	)

	givenKeeper := Given("vote keeper", func() {
		ctx, k, _, _, _ = setup()
	})

	t.Run("InitializePoll", testutils.Func(func(t *testing.T) {
		whenPollBuilderIsSet := When("poll builder is set", func() {
			votingThreshold := utilstestutils.RandThreshold()
			snapshot = snapshottestutils.Snapshot(uint64(rand.I64Between(1, 100)), votingThreshold)
			pollBuilder = exported.NewPollBuilder(rand.NormalizedStr(5), votingThreshold, snapshot, rand.PosI64()).
				RewardPoolName(rand.NormalizedStr(5)).
				MinVoterCount(rand.I64Between(0, int64(len(snapshot.Participants)))).
				GracePeriod(rand.I64Between(0, 10))
		})

		givenKeeper.
			When2(whenPollBuilderIsSet).
			Then("should successfully initialize the polls with different IDs", func(t *testing.T) {
				pollCount := rand.I64Between(1, 100)
				for i := 0; i < int(pollCount); i++ {
					actual, err := k.InitializePoll(ctx, pollBuilder)

					assert.NoError(t, err)
					assert.EqualValues(t, i, actual)

					_, ok := k.GetPoll(ctx, actual)
					assert.True(t, ok)
				}
			}).
			Run(t, 5)

		givenKeeper.
			When2(whenPollBuilderIsSet).
			When("poll is not valid", func() { pollBuilder = pollBuilder.MinVoterCount(1000) }).
			Then("should return an error", func(t *testing.T) {
				_, err := k.InitializePoll(ctx, pollBuilder)
				assert.Error(t, err)
			}).
			Run(t, 20)

		givenKeeper.
			When2(whenPollBuilderIsSet).
			Then("should bump up gas cost", func(t *testing.T) {
				gasBefore := ctx.GasMeter().GasConsumed()
				actual, err := k.InitializePoll(ctx, pollBuilder)
				gasAfter := ctx.GasMeter().GasConsumed()
				assert.NoError(t, err)

				voteCostPerMaintainer := storetypes.Gas(20000)
				gasBump := uint64(len(snapshot.GetParticipantAddresses())) * voteCostPerMaintainer
				assert.True(t, gasAfter-gasBefore > gasBump)

				_, ok := k.GetPoll(ctx, actual)
				assert.True(t, ok)

			}).
			Run(t, 20)
	}))
}
