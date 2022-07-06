package keeper

import (
	"fmt"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	utilstestutils "github.com/axelarnetwork/axelar-core/utils/testutils"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	snapshottestutils "github.com/axelarnetwork/axelar-core/x/snapshot/exported/testutils"
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
	evmtypes.RegisterInterfaces(encodingConfig.InterfaceRegistry)
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
	votingThreshold := utilstestutils.RandThreshold()
	snapshot := snapshottestutils.Snapshot(uint64(rand.I64Between(1, 100)), votingThreshold)
	pollID, err := keeper.InitializePoll(
		ctx,
		exported.NewPollBuilder(rand.NormalizedStr(5), votingThreshold, snapshot, rand.I64Between(1, 100)).
			RewardPoolName(rand.NormalizedStr(5)).
			MinVoterCount(rand.I64Between(0, int64(len(snapshot.Participants)))).
			GracePeriod(rand.I64Between(0, 10)),
	)
	if err != nil {
		panic(err)
	}

	metadata, ok := keeper.getPollMetadata(ctx, pollID)
	if !ok {
		panic(fmt.Errorf("poll metadata not found"))
	}

	metadata.State = rand.Of(exported.Completed, exported.Failed)
	if metadata.Is(exported.Completed) {
		metadata.Result = &codectypes.Any{}
		metadata.CompletedAt = rand.PosI64()
	}

	keeper.setPollMetadata(ctx, metadata)

	return metadata
}

func TestExportGenesisInitGenesis(t *testing.T) {
	ctx, keeper, _, _, _ := setup()
	keeper.InitGenesis(ctx, types.NewGenesisState(types.DefaultParams(), []exported.PollMetadata{}))

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
