package keeper

import (
	"fmt"
	"strings"
	"testing"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	rewardMock "github.com/axelarnetwork/axelar-core/x/reward/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/axelar-core/x/vote/types/mock"
)

func setup() (sdk.Context, Keeper, *mock.SnapshotterMock, *mock.StakingKeeperMock, *mock.RewarderMock) {
	snapshotter := mock.SnapshotterMock{}
	staking := mock.StakingKeeperMock{}
	rewarder := mock.RewarderMock{}

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encodingConfig := params.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(encodingConfig.Amino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	subspace := paramstypes.NewSubspace(encodingConfig.Codec, encodingConfig.Amino, sdk.NewKVStoreKey("paramsKey"), sdk.NewKVStoreKey("tparamsKey"), "vote")

	keeper := NewKeeper(
		encodingConfig.Codec,
		sdk.NewKVStoreKey(types.StoreKey),
		subspace,
		&snapshotter,
		&staking,
		&rewarder,
	)

	return ctx, keeper, &snapshotter, &staking, &rewarder
}

func initializeRandomPoll(ctx sdk.Context, keeper Keeper) exported.PollMetadata {
	voterCount := rand.I64Between(10, 100)
	voters := make([]exported.Voter, voterCount)
	totalVotingPower := sdk.ZeroInt()
	for i := range voters {
		voters[i] = exported.Voter{
			Validator:   rand.ValAddr(),
			VotingPower: rand.I64Between(1, 10),
		}
		totalVotingPower = totalVotingPower.AddRaw(voters[i].VotingPower)
	}

	pollKey := exported.PollKey{Module: randomNormalizedStr(5), ID: randomNormalizedStr(10)}
	keeper.initializePoll(ctx, pollKey, voters, totalVotingPower,
		exported.ExpiryAt(rand.PosI64()),
		exported.RewardPool(randomNormalizedStr(5)),
		exported.MinVoterCount(rand.I64Between(1, int64(len(voters)))),
		exported.Threshold(utils.NewThreshold(rand.I64Between(1, 101), 100)),
	)

	metadata, ok := keeper.getPollMetadata(ctx, pollKey)
	if !ok {
		panic(fmt.Errorf("poll metadata not found"))
	}

	pollStates := []exported.PollState{exported.Completed, exported.Failed, exported.Expired, exported.AllowOverride}
	poll := types.NewPoll(ctx, metadata, keeper.newPollStore(ctx, metadata.Key), keeper.rewarder)
	poll.State = pollStates[rand.I64Between(0, int64(len(pollStates)))]
	if poll.Is(exported.Completed) {
		poll.Result = &codectypes.Any{}
	}
	poll.SetMetadata(poll.PollMetadata)

	return poll.PollMetadata
}

func TestExportGenesisInitGenesis(t *testing.T) {
	ctx, keeper, _, _, rewarder := setup()
	keeper.InitGenesis(ctx, types.NewGenesisState(types.DefaultParams(), []exported.PollMetadata{}))

	rewarder.GetPoolFunc = func(ctx sdk.Context, name string) reward.RewardPool {
		return &rewardMock.RewardPoolMock{}
	}

	pollCount := rand.I64Between(10, 100)
	expectedPollMetadatas := make([]exported.PollMetadata, pollCount)
	for i := 0; i < int(pollCount); i++ {
		expectedPollMetadatas[i] = initializeRandomPoll(ctx, keeper)
	}

	expected := types.NewGenesisState(
		types.DefaultParams(),
		expectedPollMetadatas,
	)
	actual := keeper.ExportGenesis(ctx)

	assert.Equal(t, expected.Params, actual.Params)
	assert.ElementsMatch(t, expected.PollMetadatas, actual.PollMetadatas)
	assert.NoError(t, actual.Validate())

	ctx, keeper, _, _, _ = setup()
	keeper.InitGenesis(ctx, expected)
	actual = keeper.ExportGenesis(ctx)

	assert.Equal(t, expected.Params, actual.Params)
	assert.ElementsMatch(t, expected.PollMetadatas, actual.PollMetadatas)
	assert.NoError(t, actual.Validate())
}

func randomNormalizedStr(size int) string {
	return strings.ReplaceAll(utils.NormalizeString(rand.Str(size)), utils.DefaultDelimiter, "-")
}
