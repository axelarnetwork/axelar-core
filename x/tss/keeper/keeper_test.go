package keeper

import (
	"crypto/ecdsa"
	"crypto/rand"
	"strconv"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var (
	val1       = newValidator(sdk.ValAddress("validator1"), 100)
	val2       = newValidator(sdk.ValAddress("validator2"), 100)
	val3       = newValidator(sdk.ValAddress("validator3"), 100)
	val4       = newValidator(sdk.ValAddress("validator4"), 100)
	validators = []snapshot.Validator{val1, val2, val3, val4}
	snap       = snapshot.Snapshot{
		Validators: validators,
		Timestamp:  time.Now(),
		Height:     testutils.RandIntBetween(1, 1000000),
		TotalPower: sdk.NewInt(400),
		Counter:    testutils.RandIntBetween(0, 100000),
	}
	randPosInt      = testutils.RandIntsBetween(0, 100000000)
	randDistinctStr = testutils.RandStrings(3, 15).Distinct()
)

type testSetup struct {
	Keeper      Keeper
	Broadcaster fake.Broadcaster
	Snapshotter *snapMock.SnapshotterMock
	Ctx         sdk.Context
	PrivateKey  chan *ecdsa.PrivateKey
	Signature   chan []byte
}

func setup(t *testing.T) *testSetup {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	counter := int64(350)

	snapshotter := &snapMock.SnapshotterMock{
		GetSnapshotFunc: func(sdk.Context, int64) (snapshot.Snapshot, bool) {
			return snapshot.Snapshot{Validators: validators, TotalPower: sdk.NewInt(counter)}, true
		},
		GetLatestCounterFunc: func(ctx sdk.Context) int64 {
			return counter
		},
	}
	broadcaster := prepareBroadcaster(t, ctx, testutils.Codec(), validators)
	subspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), "tss")
	setup := &testSetup{
		Broadcaster: broadcaster,
		Snapshotter: snapshotter,
		Ctx:         ctx,
		PrivateKey:  make(chan *ecdsa.PrivateKey, 1),
		Signature:   make(chan []byte, 1),
	}

	voter := &mock.VoterMock{
		InitPollFunc:   func(ctx sdk.Context, poll exported.PollMeta) error { return nil },
		RecordVoteFunc: func(exported.MsgVote) {},
	}
	k := NewKeeper(testutils.Codec(), sdk.NewKVStoreKey("tss"), subspace, voter, broadcaster, snapshotter)
	k.SetParams(ctx, types.DefaultParams())

	setup.Keeper = k
	return setup
}

func (s *testSetup) SetLockingPeriod(lockingPeriod int64) {
	p := types.DefaultParams()
	p.LockingPeriod = lockingPeriod
	s.Keeper.SetParams(s.Ctx, p)
}

func (s *testSetup) SetKey(t *testing.T, ctx sdk.Context) (keyID string, keyChan ecdsa.PublicKey) {
	keyID = randDistinctStr.Next()
	s.PrivateKey = make(chan *ecdsa.PrivateKey, 1)
	err := s.Keeper.StartKeygen(ctx, keyID, len(validators)-1, snap)
	assert.NoError(t, err)

	sk, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	s.Keeper.SetKey(ctx, keyID, sk.PublicKey)
	return keyID, sk.PublicKey
}

func prepareBroadcaster(t *testing.T, ctx sdk.Context, cdc *codec.Codec, validators []snapshot.Validator) fake.Broadcaster {
	broadcaster := fake.NewBroadcaster(cdc, validators[0].GetOperator(), func(msg sdk.Msg) (result <-chan *fake.Result) {
		return make(chan *fake.Result)
	})

	for i, v := range validators {
		assert.NoError(t, broadcaster.RegisterProxy(ctx, v.GetOperator(), sdk.AccAddress("proxy"+strconv.Itoa(i))))
	}

	return broadcaster
}

func newValidator(address sdk.ValAddress, power int64) *snapMock.ValidatorMock {
	return &snapMock.ValidatorMock{
		GetOperatorFunc:       func() sdk.ValAddress { return address },
		GetConsensusPowerFunc: func() int64 { return power }}
}
