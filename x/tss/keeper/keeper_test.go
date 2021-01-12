package keeper

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"strconv"
	"testing"
	"time"

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

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/bitcoin/types/mock"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
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
		Round:      testutils.RandIntBetween(0, 100000),
	}
	randPosInt      = testutils.RandIntsBetween(0, 100000000)
	randDistinctStr = testutils.RandStrings(3, 15).Distinct()
)

type testSetup struct {
	Keeper      Keeper
	Broadcaster fake.Broadcaster
	Ctx         sdk.Context
	PrivateKey  chan *ecdsa.PrivateKey
	Signature   chan []byte
}

func setup(t *testing.T) *testSetup {
	ctx := sdk.NewContext(fake.NewMultiStore(), abci.Header{}, false, log.TestingLogger())
	broadcaster := prepareBroadcaster(t, ctx, testutils.Codec(), validators)
	subspace := params.NewSubspace(testutils.Codec(), sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), "tss")
	setup := &testSetup{
		Broadcaster: broadcaster,
		Ctx:         ctx,
		PrivateKey:  make(chan *ecdsa.PrivateKey, 1),
		Signature:   make(chan []byte, 1),
	}

	client := &tssdMock.TSSDClientMock{
		KeygenFunc: func(context.Context, ...grpc.CallOption) (tssd.GG18_KeygenClient, error) {
			return &tssdMock.TSSDKeyGenClientMock{
				SendFunc: func(*tssd.MessageIn) error {
					k, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
					setup.PrivateKey <- k
					return nil
				},
				RecvFunc: func() (*tssd.MessageOut, error) {
					key := <-setup.PrivateKey
					bz, _ := convert.PubkeyToBytes(key.PublicKey)
					setup.PrivateKey <- key
					return &tssd.MessageOut{Data: &tssd.MessageOut_KeygenResult{KeygenResult: bz}}, nil
				},
				CloseSendFunc: func() error { return nil },
			}, nil
		},
		SignFunc: func(context.Context, ...grpc.CallOption) (tssd.GG18_SignClient, error) {
			return &tssdMock.TSSDSignClientMock{
				SendFunc: func(in *tssd.MessageIn) error {
					k := <-setup.PrivateKey
					r, s, _ := ecdsa.Sign(rand.Reader, k, in.Data.(*tssd.MessageIn_SignInit).SignInit.MessageToSign)
					bz, _ := convert.SigToBytes(r.Bytes(), s.Bytes())
					setup.Signature <- bz
					return nil
				},
				RecvFunc: func() (*tssd.MessageOut, error) {
					return &tssd.MessageOut{Data: &tssd.MessageOut_SignResult{SignResult: <-setup.Signature}}, nil
				},
				CloseSendFunc: func() error { return nil },
			}, nil
		}}
	voter := &mock.VoterMock{InitPollFunc: func(ctx sdk.Context, poll exported.PollMeta) error {
		return nil
	}}
	k := NewKeeper(testutils.Codec(), sdk.NewKVStoreKey("tss"), client, subspace, voter, broadcaster)
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
