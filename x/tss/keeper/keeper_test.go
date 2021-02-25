package keeper

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"strconv"
	"testing"
	"time"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	tssdMock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
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
		GetSnapshotActiveValidatorsFunc: func(sdk.Context, int64) (snapshot.Snapshot, bool) {
			return snapshot.Snapshot{Validators: validators, TotalPower: sdk.NewInt(counter)}, true
		},
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

	client := &tssdMock.TofndClientMock{
		KeygenFunc: func(context.Context, ...grpc.CallOption) (tofnd.GG20_KeygenClient, error) {
			return &tssdMock.TofndKeyGenClientMock{
				SendFunc: func(*tofnd.MessageIn) error {
					k, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
					setup.PrivateKey <- k
					return nil
				},
				RecvFunc: func() (*tofnd.MessageOut, error) {
					key := <-setup.PrivateKey
					btcecPK := btcec.PublicKey(key.PublicKey)
					bz := btcecPK.SerializeCompressed()
					setup.PrivateKey <- key
					return &tofnd.MessageOut{Data: &tofnd.MessageOut_KeygenResult{KeygenResult: bz}}, nil
				},
				CloseSendFunc: func() error { return nil },
			}, nil
		},
		SignFunc: func(context.Context, ...grpc.CallOption) (tofnd.GG20_SignClient, error) {
			return &tssdMock.TofndSignClientMock{
				SendFunc: func(in *tofnd.MessageIn) error {
					k := <-setup.PrivateKey
					r, s, _ := ecdsa.Sign(rand.Reader, k, in.Data.(*tofnd.MessageIn_SignInit).SignInit.MessageToSign)
					btcecSig := btcec.Signature{R: r, S: s}
					bz := btcecSig.Serialize()
					setup.Signature <- bz
					return nil
				},
				RecvFunc: func() (*tofnd.MessageOut, error) {
					return &tofnd.MessageOut{Data: &tofnd.MessageOut_SignResult{SignResult: <-setup.Signature}}, nil
				},
				CloseSendFunc: func() error { return nil },
			}, nil
		}}
	voter := &mock.VoterMock{
		InitPollFunc:   func(ctx sdk.Context, poll exported.PollMeta) error { return nil },
		RecordVoteFunc: func(exported.MsgVote) {},
	}
	k := NewKeeper(testutils.Codec(), sdk.NewKVStoreKey("tss"), client, subspace, voter, broadcaster, snapshotter)
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
	res, err := s.Keeper.StartKeygen(ctx, keyID, len(validators)-1, snap)
	assert.NoError(t, err)

	publicKey := <-res
	s.Keeper.SetKey(ctx, keyID, publicKey)
	return keyID, publicKey
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
