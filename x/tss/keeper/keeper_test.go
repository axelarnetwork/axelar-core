package keeper

import (
	"context"
	"strconv"
	"testing"

	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/mock"
	stExported "github.com/axelarnetwork/axelar-core/x/staking/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
)

func TestKeeper_IsKeyRefreshLocked_Locked(t *testing.T) {
	s := setup(t)

	for _, currHeight := range testutils.RandIntsBetween(0, 100000).Take(100) {
		ctx := s.ctx.WithBlockHeight(int64(currHeight))

		// snapshotHeight + lockingPeriod > currHeight
		lockingPeriod := testutils.RandIntsBetween(0, currHeight).Next()
		snapshotHeight := testutils.RandIntsBetween(currHeight-lockingPeriod+1, currHeight).Next()

		p := types.DefaultParams()
		p.LockingPeriod = int64(lockingPeriod)
		s.keeper.SetParams(s.ctx, p)

		assert.True(t, s.keeper.IsKeyRefreshLocked(ctx, "", int64(snapshotHeight)))
	}
}

func TestKeeper_IsKeyRefreshLocked_Unlocked(t *testing.T) {
	s := setup(t)

	for _, currHeight := range testutils.RandIntsBetween(0, 100000).Take(100) {
		ctx := s.ctx.WithBlockHeight(int64(currHeight))

		// snapshotHeight + lockingPeriod <= currHeight
		lockingPeriod := testutils.RandIntsBetween(0, currHeight).Next()
		snapshotHeight := testutils.RandIntsBetween(0, currHeight-lockingPeriod+1).Next()

		p := types.DefaultParams()
		p.LockingPeriod = int64(lockingPeriod)
		s.keeper.SetParams(s.ctx, p)

		assert.False(t, s.keeper.IsKeyRefreshLocked(ctx, "", int64(snapshotHeight)))
	}
}

type testSetup struct {
	keeper      Keeper
	staker      mock.Staker
	voter       mockVoter
	broadcaster mock.Broadcaster
	ctx         sdk.Context
	client      mockTssClient
}

func setup(t *testing.T) testSetup {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	staker := newStaker()
	broadcaster := prepareBroadcaster(t, ctx, testutils.Codec(), staker.GetAllValidators(ctx), nil)
	subspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), "tss")
	voter := mockVoter{receivedVote: make(chan exported.MsgVote, 1000), initializedPoll: make(chan exported.PollMeta, 100)}
	client := mockTssClient{keygen: mockKeyGenClient{recv: make(chan *tssd.MessageOut, 1)}}
	k := NewKeeper(mock.NewKVStoreKey("tss"), testutils.Codec(), client, subspace, broadcaster)
	k.SetParams(ctx, types.DefaultParams())
	return testSetup{
		keeper:      k,
		staker:      staker,
		broadcaster: broadcaster,
		ctx:         ctx,
		client:      client,
		voter:       voter,
	}
}

func newStaker() mock.Staker {
	val1 := stExported.Validator{Address: sdk.ValAddress("validator1"), Power: 100}
	val2 := stExported.Validator{Address: sdk.ValAddress("validator2"), Power: 100}
	val3 := stExported.Validator{Address: sdk.ValAddress("validator3"), Power: 100}
	val4 := stExported.Validator{Address: sdk.ValAddress("validator4"), Power: 100}
	staker := mock.NewTestStaker(1, val1, val2, val3, val4)
	return staker
}

func prepareBroadcaster(t *testing.T, ctx sdk.Context, cdc *codec.Codec, validators []stExported.Validator, msgIn chan sdk.Msg) mock.Broadcaster {
	broadcaster := mock.NewBroadcaster(cdc, sdk.AccAddress("proxy0"), validators[0].Address, msgIn)

	for i, v := range validators {
		assert.NoError(t, broadcaster.RegisterProxy(ctx, v.Address, sdk.AccAddress("proxy"+strconv.Itoa(i))))
	}

	return broadcaster
}

type mockTssClient struct {
	keygen mockKeyGenClient
	sign   mockSignClient
}

func (tc mockTssClient) Keygen(_ context.Context, _ ...grpc.CallOption) (tssd.GG18_KeygenClient, error) {
	return tc.keygen, nil
}

func (tc mockTssClient) Sign(_ context.Context, _ ...grpc.CallOption) (tssd.GG18_SignClient, error) {
	return tc.sign, nil
}

func (tc mockTssClient) GetKey(_ context.Context, _ *tssd.Uid, _ ...grpc.CallOption) (*tssd.Bytes, error) {
	panic("implement me")
}

func (tc mockTssClient) GetSig(_ context.Context, _ *tssd.Uid, _ ...grpc.CallOption) (*tssd.Bytes, error) {
	panic("implement me")
}

type mockKeyGenClient struct {
	recv chan *tssd.MessageOut
}

func (kc mockKeyGenClient) Send(_ *tssd.MessageIn) error {
	return nil
}

func (kc mockKeyGenClient) Recv() (*tssd.MessageOut, error) {
	return <-kc.recv, nil
}

func (kc mockKeyGenClient) Header() (metadata.MD, error) {
	panic("implement me")
}

func (kc mockKeyGenClient) Trailer() metadata.MD {
	panic("implement me")
}

func (kc mockKeyGenClient) CloseSend() error {
	return nil
}

func (kc mockKeyGenClient) Context() context.Context {
	panic("implement me")
}

func (kc mockKeyGenClient) SendMsg(_ interface{}) error {
	panic("implement me")
}

func (kc mockKeyGenClient) RecvMsg(_ interface{}) error {
	panic("implement me")
}

type mockSignClient struct {
	recv chan *tssd.MessageOut
}

func (sc mockSignClient) Send(_ *tssd.MessageIn) error {
	return nil
}

func (sc mockSignClient) Recv() (*tssd.MessageOut, error) {
	return <-sc.recv, nil
}

func (sc mockSignClient) Header() (metadata.MD, error) {
	panic("implement me")
}

func (sc mockSignClient) Trailer() metadata.MD {
	panic("implement me")
}

func (sc mockSignClient) CloseSend() error {
	panic("implement me")
}

func (sc mockSignClient) Context() context.Context {
	panic("implement me")
}

func (sc mockSignClient) SendMsg(_ interface{}) error {
	panic("implement me")
}

func (sc mockSignClient) RecvMsg(_ interface{}) error {
	panic("implement me")
}

type mockVoter struct {
	receivedVote    chan exported.MsgVote
	initializedPoll chan exported.PollMeta
}

func (m mockVoter) InitPoll(_ sdk.Context, poll exported.PollMeta) error {
	m.initializedPoll <- poll
	return nil
}

func (m mockVoter) Vote(_ sdk.Context, vote exported.MsgVote) error {
	m.receivedVote <- vote
	return nil
}

func (m mockVoter) TallyVote(_ sdk.Context, _ exported.MsgVote) error {
	panic("implement me")
}

func (m mockVoter) Result(_ sdk.Context, _ exported.PollMeta) exported.Vote {
	panic("implement me")
}
