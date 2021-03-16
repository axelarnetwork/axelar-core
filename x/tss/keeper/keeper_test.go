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

	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
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
		Height:     rand2.I64Between(1, 1000000),
		TotalPower: sdk.NewInt(400),
		Counter:    rand2.I64Between(0, 100000),
	}
	randPosInt      = rand2.I64GenBetween(0, 100000000)
	randDistinctStr = rand2.Strings(3, 15).Distinct()
)

type testSetup struct {
	Keeper      Keeper
	Broadcaster fake.Broadcaster
	Voter       types.Voter
	Ctx         sdk.Context
	PrivateKey  chan *ecdsa.PrivateKey
	Signature   chan []byte
}

func setup(t *testing.T) *testSetup {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())

	broadcaster := prepareBroadcaster(t, ctx, testutils.Codec(), validators)
	voter := &mock.VoterMock{
		InitPollFunc:   func(ctx sdk.Context, poll exported.PollMeta) error { return nil },
		RecordVoteFunc: func(exported.MsgVote) {},
	}
	subspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), "tss")
	setup := &testSetup{
		Broadcaster: broadcaster,
		Voter:       voter,
		Ctx:         ctx,
		PrivateKey:  make(chan *ecdsa.PrivateKey, 1),
		Signature:   make(chan []byte, 1),
	}

	k := NewKeeper(testutils.Codec(), sdk.NewKVStoreKey("tss"), subspace, broadcaster)
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
	err := s.Keeper.StartKeygen(ctx, s.Voter, keyID, len(validators)-1, snap)
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
