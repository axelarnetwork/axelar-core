package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	store "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
	exportedMock "github.com/axelarnetwork/axelar-core/x/vote/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/vote/keeper"
	"github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/axelar-core/x/vote/types/mock"
)

func TestMsgServer_UpdateParams(t *testing.T) {
	enc := appParams.MakeEncodingConfig()
	subspace := paramstypes.NewSubspace(enc.Codec, enc.Amino, store.NewKVStoreKey("voteParams"), store.NewKVStoreKey("tvoteParams"), "vote")
	k := keeper.NewKeeper(enc.Codec, store.NewKVStoreKey(types.StoreKey), subspace, &mock.SnapshotterMock{}, &mock.StakingKeeperMock{}, &mock.RewarderMock{})
	ctx := sdk.NewContext(fake.NewMultiStore(), sdk.Context{}.BlockHeader(), false, log.NewTestLogger(t))
	server := keeper.NewMsgServerImpl(k)

	p := types.DefaultParams()
	p.EndBlockerLimit = p.EndBlockerLimit + 1
	_, err := server.UpdateParams(ctx, &types.UpdateParamsRequest{Authority: rand.AccAddr().String(), Params: p})
	assert.NoError(t, err)
	got := k.GetParams(ctx)
	assert.Equal(t, p, got)
}

func TestMsgServer_VoteGracePeriod(t *testing.T) {
	// Setup encoding and keeper
	enc := appParams.MakeEncodingConfig()
	types.RegisterLegacyAminoCodec(enc.Amino)
	types.RegisterInterfaces(enc.InterfaceRegistry)
	enc.InterfaceRegistry.RegisterImplementations((*codec.ProtoMarshaler)(nil), &evmtypes.VoteEvents{})
	subspace := paramstypes.NewSubspace(enc.Codec, enc.Amino, store.NewKVStoreKey("voteParams"), store.NewKVStoreKey("tvoteParams"), "vote")

	// Create voters
	voter1Proxy, voter2Proxy, voter3Proxy := rand.AccAddr(), rand.AccAddr(), rand.AccAddr()
	voter1Val, voter2Val, voter3Val := rand.ValAddr(), rand.ValAddr(), rand.ValAddr()

	// Mock snapshotter to return validator addresses for proxies
	snapshotter := &mock.SnapshotterMock{
		GetOperatorFunc: func(ctx sdk.Context, proxy sdk.AccAddress) sdk.ValAddress {
			switch proxy.String() {
			case voter1Proxy.String():
				return voter1Val
			case voter2Proxy.String():
				return voter2Val
			case voter3Proxy.String():
				return voter3Val
			}
			return nil
		},
	}

	k := keeper.NewKeeper(enc.Codec, store.NewKVStoreKey(types.StoreKey), subspace, snapshotter, &mock.StakingKeeperMock{}, &mock.RewarderMock{})

	// Register a mock vote handler for test-module
	voteHandler := &exportedMock.VoteHandlerMock{
		IsFalsyResultFunc:       func(codec.ProtoMarshaler) bool { return false },
		HandleResultFunc:        func(sdk.Context, codec.ProtoMarshaler) error { return nil },
		HandleExpiredPollFunc:   func(sdk.Context, exported.Poll) error { return nil },
		HandleFailedPollFunc:    func(sdk.Context, exported.Poll) error { return nil },
		HandleCompletedPollFunc: func(sdk.Context, exported.Poll) error { return nil },
	}
	k.SetVoteRouter(types.NewRouter().AddHandler("test-module", voteHandler))

	ctx := sdk.NewContext(fake.NewMultiStore(), sdk.Context{}.BlockHeader(), false, log.NewTestLogger(t)).WithBlockHeight(100)
	k.SetParams(ctx, types.DefaultParams())
	server := keeper.NewMsgServerImpl(k)

	// Create a poll with voters
	participants := []snapshot.Participant{
		snapshot.NewParticipant(voter1Val, math.NewUint(100)),
		snapshot.NewParticipant(voter2Val, math.NewUint(100)),
		snapshot.NewParticipant(voter3Val, math.NewUint(100)),
	}
	pollBuilder := exported.NewPollBuilder(
		"test-module",
		utils.NewThreshold(51, 100), // 51% threshold
		snapshot.NewSnapshot(time.Now(), 1, participants, math.NewUint(300)),
		ctx.BlockHeight()+100, // expires at block 200
	).GracePeriod(1) // 1 block grace period

	pollID, err := k.InitializePoll(ctx, pollBuilder)
	assert.NoError(t, err)

	voteData := &evmtypes.VoteEvents{Events: []evmtypes.Event{{}}}

	// Voter1 votes
	_, err = server.Vote(ctx, types.NewVoteRequest(voter1Proxy, pollID, voteData))
	assert.NoError(t, err)

	// Voter2 votes - this should complete the poll (2/3 = 66% > 51% threshold)
	_, err = server.Vote(ctx, types.NewVoteRequest(voter2Proxy, pollID, voteData))
	assert.NoError(t, err)

	// Verify poll is completed
	poll, ok := k.GetPoll(ctx, pollID)
	assert.True(t, ok)
	assert.Equal(t, exported.Completed, poll.GetState())

	// Voter3 votes after grace period (block 100 + 2 = 102, grace period ends at 101)
	ctx = ctx.WithBlockHeight(102)

	_, err = server.Vote(ctx, types.NewVoteRequest(voter3Proxy, pollID, voteData))
	assert.ErrorContains(t, err, "poll completed")
}
