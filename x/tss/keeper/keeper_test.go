package keeper

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"strconv"
	"testing"

	"github.com/axelarnetwork/tssd/convert"
	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/btcsuite/btcd/btcec"
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
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

type testSetup struct {
	Keeper          Keeper
	Staker          mock.Snapshotter
	Voter           mockVoter
	Broadcaster     mock.Broadcaster
	Ctx             sdk.Context
	Client          mockTssClient
	RandDistinctStr testutils.RandDistinctStringGen
	RandPosInt      testutils.RandIntGen
}

func setup(t *testing.T) testSetup {
	ctx := sdk.NewContext(mock.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	snapshotter := newSnapshotter()
	broadcaster := prepareBroadcaster(t, ctx, testutils.Codec(), snapshotter.GetAllValidators())
	subspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), "tss")
	voter := mockVoter{receivedVote: make(chan vote.MsgVote, 1000), initializedPoll: make(chan vote.PollMeta, 100)}
	client := mockTssClient{keygen: mockKeyGenClient{recv: make(chan *tssd.MessageOut, 1)}}
	k := NewKeeper(testutils.Codec(), mock.NewKVStoreKey("tss"), client, subspace, broadcaster)
	k.SetParams(ctx, types.DefaultParams())

	return testSetup{
		Keeper:          k,
		Staker:          snapshotter,
		Broadcaster:     broadcaster,
		Ctx:             ctx,
		Client:          client,
		Voter:           voter,
		RandPosInt:      testutils.RandIntsBetween(0, 100000000),
		RandDistinctStr: testutils.RandStrings(3, 15).Distinct(),
	}
}

func (s testSetup) SetLockingPeriod(lockingPeriod int64) {
	p := types.DefaultParams()
	p.LockingPeriod = lockingPeriod
	s.Keeper.SetParams(s.Ctx, p)
}

func (s testSetup) SetKeygenResult(pk ecdsa.PublicKey) {
	bz, _ := convert.PubkeyToBytes(pk)
	s.Client.keygen.recv <- &tssd.MessageOut{Data: &tssd.MessageOut_KeygenResult{KeygenResult: bz}}
}

func (s testSetup) SetKey(t *testing.T, ctx sdk.Context) (keyID string, keyChan ecdsa.PublicKey) {
	keyID = s.RandDistinctStr.Next()
	key := s.RandomPK()
	res, err := s.Keeper.StartKeygen(ctx, keyID, len(s.Staker.GetAllValidators())-1, s.Staker.GetAllValidators())
	assert.NoError(t, err)

	s.SetKeygenResult(key)

	publicKey := <-res
	s.Keeper.SetKey(ctx, keyID, publicKey)
	return keyID, publicKey
}

func (s testSetup) RandomPK() ecdsa.PublicKey {
	return ecdsa.PublicKey{
		Curve: btcec.S256(),
		X:     big.NewInt(s.RandPosInt.Next()),
		Y:     big.NewInt(s.RandPosInt.Next()),
	}
}

func (s testSetup) Teardown() {
	s.RandPosInt.Stop()
	s.RandDistinctStr.Stop()
}

func newSnapshotter() mock.Snapshotter {
	val1 := snapshot.Validator{Address: sdk.ValAddress("validator1"), Power: 100}
	val2 := snapshot.Validator{Address: sdk.ValAddress("validator2"), Power: 100}
	val3 := snapshot.Validator{Address: sdk.ValAddress("validator3"), Power: 100}
	val4 := snapshot.Validator{Address: sdk.ValAddress("validator4"), Power: 100}
	staker := mock.NewTestStaker(1, val1, val2, val3, val4)
	return staker
}

func prepareBroadcaster(t *testing.T, ctx sdk.Context, cdc *codec.Codec, validators []snapshot.Validator) mock.Broadcaster {
	broadcaster := mock.NewBroadcaster(cdc, validators[0].Address, func(msg sdk.Msg) (result <-chan mock.Result) {
		return make(chan mock.Result)
	})

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
	receivedVote    chan vote.MsgVote
	initializedPoll chan vote.PollMeta
}

func (m mockVoter) InitPoll(_ sdk.Context, poll vote.PollMeta) error {
	m.initializedPoll <- poll
	return nil
}

func (m mockVoter) Vote(_ sdk.Context, vote vote.MsgVote) error {
	m.receivedVote <- vote
	return nil
}

func (m mockVoter) TallyVote(_ sdk.Context, _ vote.MsgVote) error {
	panic("implement me")
}

func (m mockVoter) Result(_ sdk.Context, _ vote.PollMeta) vote.Vote {
	panic("implement me")
}
