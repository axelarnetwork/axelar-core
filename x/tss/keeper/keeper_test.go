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
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	tssMock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

var (
	val1       = newValidator(rand.ValAddr(), 100)
	val2       = newValidator(rand.ValAddr(), 100)
	val3       = newValidator(rand.ValAddr(), 100)
	val4       = newValidator(rand.ValAddr(), 100)
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
	Keeper      Keeper
	Voter       types.Voter
	Snapshotter *snapMock.SnapshotterMock
	Ctx         sdk.Context
	PrivateKey  chan *ecdsa.PrivateKey
	Signature   chan []byte
}

func setup() *testSetup {
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	encCfg := appParams.MakeEncodingConfig()
	voter := &tssMock.VoterMock{
		InitializePollFunc: func(sdk.Context, vote.PollKey, int64, ...vote.PollProperty) error { return nil },
	}
	snapshotter := &snapMock.SnapshotterMock{}

	subspace := params.NewSubspace(encCfg.Marshaler, encCfg.Amino, sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), "tss")
	setup := &testSetup{
		Voter:       voter,
		Snapshotter: snapshotter,
		Ctx:         ctx,
		PrivateKey:  make(chan *ecdsa.PrivateKey, 1),
		Signature:   make(chan []byte, 1),
	}

	slasher := &snapMock.SlasherMock{
		GetValidatorSigningInfoFunc: func(ctx sdk.Context, address sdk.ConsAddress) (slashingTypes.ValidatorSigningInfo, bool) {
			newInfo := slashingTypes.NewValidatorSigningInfo(
				address,
				int64(0),        // height at which validator was first a candidate OR was unjailed
				int64(3),        // index offset into signed block bit array. TODO: check if needs to be set correctly.
				time.Unix(0, 0), // jailed until
				false,           // tomstoned
				int64(0),        // missed blocks
			)

			return newInfo, true
		},
		SignedBlocksWindowFunc: func(sdk.Context) int64 { return 100 },
	}

	k := NewKeeper(encCfg.Amino, sdk.NewKVStoreKey("tss"), subspace, slasher)
	k.SetParams(ctx, types.DefaultParams())

	setup.Keeper = k
	return setup
}

func (s *testSetup) SetKey(t *testing.T, ctx sdk.Context, keyRole exported.KeyRole) tss.Key {
	keyID := exported.KeyID(randDistinctStr.Next())
	s.PrivateKey = make(chan *ecdsa.PrivateKey, 1)
	err := s.Keeper.StartKeygen(ctx, s.Voter, keyID, tss.MasterKey, snap)
	assert.NoError(t, err)

	sk, err := ecdsa.GenerateKey(btcec.S256(), cryptoRand.Reader)
	if err != nil {
		panic(err)
	}
	s.Keeper.SetKey(ctx, keyID, sk.PublicKey)
	s.Keeper.SetKeyRole(ctx, keyID, keyRole)

	return tss.Key{
		ID:    keyID,
		Value: sk.PublicKey,
		Role:  keyRole,
	}
}

func newValidator(address sdk.ValAddress, power int64) snapshot.Validator {
	return snapshot.NewValidator(&snapMock.SDKValidatorMock{
		GetOperatorFunc:       func() sdk.ValAddress { return address },
		GetConsensusPowerFunc: func() int64 { return power },
		GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return address.Bytes(), nil },
		IsJailedFunc:          func() bool { return false },
	}, power)
}

func TestAvailableOperator(t *testing.T) {
	t.Run("testing available operators", testutils.Func(func(t *testing.T) {
		s := setup()
		acks := []exported.AckType{exported.AckType_Keygen, exported.AckType_Keygen}
		repeats := int(rand.I64Between(5, 20))
		snapshotSeq := rand.I64Between(1, 100)

		for i := 0; i < repeats; i++ {
			id := rand.StrBetween(5, 10)
			index := int(rand.I64Between(0, int64(len(acks)-1)))
			ackType := acks[index]
			index = int(rand.I64Between(0, int64(len(snap.Validators)-1)))
			validator := snap.Validators[index].GetSDKValidator().GetOperator()
			snapshotSeq = snapshotSeq + rand.I64Between(1, 10)

			// not yet available
			assert.False(t, s.Keeper.IsOperatorAvailable(s.Ctx, id, ackType, validator))

			// available
			err := s.Keeper.SetAvailableOperator(s.Ctx, id, ackType, validator)
			assert.NoError(t, err)
			assert.True(t, s.Keeper.IsOperatorAvailable(s.Ctx, id, ackType, validator))

			// replaying
			err = s.Keeper.SetAvailableOperator(s.Ctx, id, ackType, validator)
			assert.EqualError(t, err, "validator already submitted its ack for the specified ID and type")

			// linked to counter
			assert.False(t, s.Keeper.OperatorIsAvailableForCounter(s.Ctx, snapshotSeq, validator))
			s.Keeper.LinkAvailableOperatorsToSnapshot(s.Ctx, id, ackType, snapshotSeq)
			assert.True(t, s.Keeper.OperatorIsAvailableForCounter(s.Ctx, snapshotSeq, validator))

			// delete available
			s.Keeper.DeleteAvailableOperators(s.Ctx, id, ackType)
			assert.False(t, s.Keeper.IsOperatorAvailable(s.Ctx, id, ackType, validator))
		}
	}).Repeat(20))
}

func TestActiveOldKeys(t *testing.T) {
	t.Run("testing locked rotation keys", testutils.Func(func(t *testing.T) {
		s := setup()
		chain := bitcoin.Bitcoin
		iterations := int(rand2.I64Between(2, 10) * s.Keeper.GetKeyUnbondingLockingKeyRotationCount(s.Ctx))
		role := exported.GetKeyRoles()[int(rand2.I64Between(0, int64(len(exported.GetKeyRoles()))))]
		var expectedKeys []exported.Key

		for i := 0; i < iterations; i++ {
			expectedMasterKey := s.SetKey(t, s.Ctx, role)
			assert.NoError(t, s.Keeper.AssignNextKey(s.Ctx, chain, role, expectedMasterKey.ID))
			assert.NoError(t, s.Keeper.RotateKey(s.Ctx, chain, role))
			expectedKeys = append(expectedKeys, expectedMasterKey)
		}

		keys, err := s.Keeper.GetOldActiveKeys(s.Ctx, chain, role)
		assert.NoError(t, err)
		assert.Len(t, keys, int(s.Keeper.GetKeyUnbondingLockingKeyRotationCount(s.Ctx)))

		count := 0
		for _, actual := range keys {
			for _, expected := range expectedKeys {
				if actual.ID == expected.ID && actual.Role == expected.Role {
					count++
				}
			}
		}
		assert.Equal(t, int(s.Keeper.GetKeyUnbondingLockingKeyRotationCount(s.Ctx)), count)

	}).Repeat(20))
}
