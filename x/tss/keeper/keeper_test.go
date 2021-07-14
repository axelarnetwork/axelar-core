package keeper

import (
	"crypto/ecdsa"
	cryptoRand "crypto/rand"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssMock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
	"github.com/axelarnetwork/axelar-core/x/vote/exported"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var (
	val1       = newValidator(sdk.ValAddress("validator1"), 100)
	val2       = newValidator(sdk.ValAddress("validator2"), 100)
	val3       = newValidator(sdk.ValAddress("validator3"), 100)
	val4       = newValidator(sdk.ValAddress("validator4"), 100)
	validators = []snapshot.Validator{val1, val2, val3, val4}
	snap       = snapshot.Snapshot{
		Validators:      validators,
		Timestamp:       time.Now(),
		Height:          rand2.I64Between(1, 1000000),
		TotalShareCount: sdk.NewInt(400),
		Counter:         rand2.I64Between(0, 100000),
	}
	randPosInt      = rand2.I64GenBetween(0, 100000000)
	randDistinctStr = rand2.Strings(3, 15).Distinct()
)

type testSetup struct {
	Keeper     Keeper
	Voter      types.Voter
	Ctx        sdk.Context
	PrivateKey chan *ecdsa.PrivateKey
	Signature  chan []byte
}

func setup() *testSetup {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encCfg := appParams.MakeEncodingConfig()
	voter := &tssMock.VoterMock{
		InitializePollFunc: func(sdk.Context, exported.PollKey, int64, ...exported.PollProperty) error { return nil },
	}

	subspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), "tss")
	setup := &testSetup{
		Voter:      voter,
		Ctx:        ctx,
		PrivateKey: make(chan *ecdsa.PrivateKey, 1),
		Signature:  make(chan []byte, 1),
	}

	slasher := &snapMock.SlasherMock{
		GetValidatorSigningInfoFunc: func(ctx sdk.Context, address sdk.ConsAddress) (snapshot.ValidatorInfo, bool) {
			newInfo := slashingTypes.NewValidatorSigningInfo(
				address,
				int64(0),        // height at which validator was first a candidate OR was unjailed
				int64(3),        // index offset into signed block bit array. TODO: check if needs to be set correctly.
				time.Unix(0, 0), // jailed until
				false,           // tomstoned
				int64(0),        // missed blocks
			)
			return snapshot.ValidatorInfo{ValidatorSigningInfo: newInfo}, true
		},
	}

	k := NewKeeper(encCfg.Amino, sdk.NewKVStoreKey("tss"), subspace, slasher)
	k.SetParams(ctx, types.DefaultParams())

	setup.Keeper = k
	return setup
}

func (s *testSetup) SetLockingPeriod(lockingPeriod int64) {
	p := types.DefaultParams()
	p.LockingPeriod = lockingPeriod
	s.Keeper.SetParams(s.Ctx, p)
}

func (s *testSetup) SetKey(t *testing.T, ctx sdk.Context) tss.Key {
	keyID := randDistinctStr.Next()
	s.PrivateKey = make(chan *ecdsa.PrivateKey, 1)
	err := s.Keeper.StartKeygen(ctx, s.Voter, keyID, snap)
	assert.NoError(t, err)

	sk, err := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
	if err != nil {
		panic(err)
	}
	s.Keeper.SetKey(ctx, keyID, sk.PublicKey)

	return tss.Key{
		ID:    keyID,
		Value: sk.PublicKey,
	}
}

func newValidator(address sdk.ValAddress, power int64) snapshot.Validator {
	return snapshot.NewValidator(&snapMock.SDKValidatorMock{
		GetOperatorFunc:       func() sdk.ValAddress { return address },
		GetConsensusPowerFunc: func() int64 { return power },
		GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return address.Bytes(), nil },
	}, power)
}

func TestComputeCorruptionThreshold(t *testing.T) {
	s := setup()
	defaultParams := types.DefaultParams()

	s.Keeper.SetParams(s.Ctx, defaultParams)
	assert.Equal(t, int64(5), s.Keeper.ComputeCorruptionThreshold(s.Ctx, sdk.NewInt(10)))

	defaultParams.CorruptionThreshold = utils.Threshold{Numerator: 99, Denominator: 100}
	s.Keeper.SetParams(s.Ctx, defaultParams)
	assert.Equal(t, int64(8), s.Keeper.ComputeCorruptionThreshold(s.Ctx, sdk.NewInt(10)))

	defaultParams.CorruptionThreshold = utils.Threshold{Numerator: 1, Denominator: 100}
	s.Keeper.SetParams(s.Ctx, defaultParams)
	assert.Equal(t, int64(-1), s.Keeper.ComputeCorruptionThreshold(s.Ctx, sdk.NewInt(10)))
}
