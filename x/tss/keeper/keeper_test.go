package keeper

import (
	"crypto/ecdsa"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	tssMock "github.com/axelarnetwork/axelar-core/x/tss/types/mock"
	vote "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

var (
	val1       = newValidator(rand.ValAddr(), 10)
	val2       = newValidator(rand.ValAddr(), 10)
	val3       = newValidator(rand.ValAddr(), 10)
	val4       = newValidator(rand.ValAddr(), 10)
	validators = []snapshot.Validator{val1, val2, val3, val4}
	snap       = snapshot.Snapshot{
		Validators:      validators,
		Timestamp:       time.Now(),
		Height:          rand.I64Between(1, 1000000),
		TotalShareCount: sdk.NewInt(400),
		Counter:         rand.I64Between(0, 100000),
	}
	randDistinctStr = rand.Strings(3, 15).Distinct()
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
		InitializePollFunc: func(ctx sdk.Context, pollBuilder vote.PollBuilder) (vote.PollID, error) { return 0, nil },
	}
	snapshotter := &snapMock.SnapshotterMock{
		GetValidatorIllegibilityFunc: func(ctx sdk.Context, validator snapshot.SDKValidator) (snapshot.ValidatorIllegibility, error) {
			return snapshot.None, nil
		},
		GetSnapshotFunc: func(ctx sdk.Context, seqNo int64) (snapshot.Snapshot, bool) {
			if seqNo == snap.Counter {
				return snap, true
			}
			return snapshot.Snapshot{}, false
		},
	}

	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("storeKey"), sdk.NewKVStoreKey("tstorekey"), "tss")
	setup := &testSetup{
		Voter:       voter,
		Snapshotter: snapshotter,
		Ctx:         ctx,
		PrivateKey:  make(chan *ecdsa.PrivateKey, 1),
		Signature:   make(chan []byte, 1),
	}

	slasher := &mock.SlasherMock{
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
	rewarder := &tssMock.RewarderMock{}

	k := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("tss"), subspace, slasher, rewarder)
	k.SetParams(ctx, types.DefaultParams())

	setup.Keeper = k
	return setup
}

func (s *testSetup) SetKey(t *testing.T, ctx sdk.Context, keyRole exported.KeyRole, keyType exported.KeyType) exported.Key {
	keyID := exported.KeyID(randDistinctStr.Next())
	s.PrivateKey = make(chan *ecdsa.PrivateKey, 1)
	keyInfo := types.KeyInfo{
		KeyID:   keyID,
		KeyRole: keyRole,
		KeyType: keyType,
	}

	err := s.Keeper.StartKeygen(ctx, s.Voter, keyInfo, snap)
	assert.NoError(t, err)

	var key exported.Key
	switch keyType {
	case exported.Threshold:
		key = generateECDSAKey(keyID)
	case exported.Multisig:
		key = generateMultisigKey(keyID)
	}

	key.Role = keyRole
	key.Type = keyType

	s.Keeper.setKeyInfo(ctx, keyInfo)
	s.Keeper.SetKey(ctx, key)
	return key
}

func newValidator(address sdk.ValAddress, power int64) snapshot.Validator {
	return snapshot.NewValidator(&snapMock.SDKValidatorMock{
		GetOperatorFunc:       func() sdk.ValAddress { return address },
		GetConsensusPowerFunc: func(sdk.Int) int64 { return power },
		GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return address.Bytes(), nil },
		IsJailedFunc:          func() bool { return false },
		StringFunc:            func() string { return address.String() },
	}, power)
}

func randKeyIDs() []exported.KeyID {
	keyIDs := make([]exported.KeyID, 10)
	for i := range keyIDs {
		keyIDs[i] = exported.KeyID(rand.HexStr(int(rand.I64Between(exported.KeyIDLengthMin, exported.KeyIDLengthMax))))
	}
	return keyIDs
}

func TestAvailableOperator(t *testing.T) {
	t.Run("operator available", testutils.Func(func(t *testing.T) {
		s := setup()
		keyIDs := randKeyIDs()

		eventHeight := s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx) * rand.I64Between(1, 10)
		height := eventHeight + rand.I64Between(1, s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx))
		repeats := int(rand.I64Between(5, 20))

		for i := 0; i < repeats; i++ {
			availableValidator := rand.ValAddr()
			s.Ctx = s.Ctx.WithBlockHeight(height)

			// not yet available
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx), i)
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx, keyIDs...), i)

			// available
			s.Keeper.SetAvailableOperator(s.Ctx, availableValidator, keyIDs...)
			s.Ctx = s.Ctx.WithBlockHeight(eventHeight + s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx))
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx), i+1)
			assert.True(t, s.Keeper.IsOperatorAvailable(s.Ctx, availableValidator))
			assert.Contains(t, s.Keeper.GetAvailableOperators(s.Ctx), availableValidator)
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx, keyIDs...), i+1)
			assert.True(t, s.Keeper.IsOperatorAvailable(s.Ctx, availableValidator, keyIDs...))
			assert.Contains(t, s.Keeper.GetAvailableOperators(s.Ctx, keyIDs...), availableValidator)

			// unknown keys
			unknownKeys := randKeyIDs()
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx, unknownKeys...), 0)
			assert.NotContains(t, s.Keeper.GetAvailableOperators(s.Ctx, unknownKeys...), availableValidator)
			assert.False(t, s.Keeper.IsOperatorAvailable(s.Ctx, availableValidator, unknownKeys...))
		}
	}).Repeat(20))

	t.Run("operator available (edge case #1)", testutils.Func(func(t *testing.T) {
		s := setup()
		keyIDs := randKeyIDs()

		eventHeight := s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx) * rand.I64Between(1, 10)
		height := eventHeight
		repeats := int(rand.I64Between(5, 20))

		for i := 0; i < repeats; i++ {
			availableValidator := rand.ValAddr()
			s.Ctx = s.Ctx.WithBlockHeight(height)

			// not yet available
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx), i)
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx, keyIDs...), i)

			// available
			s.Keeper.SetAvailableOperator(s.Ctx, availableValidator, keyIDs...)
			s.Ctx = s.Ctx.WithBlockHeight(eventHeight + s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx))
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx), i+1)
			assert.True(t, s.Keeper.IsOperatorAvailable(s.Ctx, availableValidator))
			assert.Contains(t, s.Keeper.GetAvailableOperators(s.Ctx), availableValidator)
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx, keyIDs...), i+1)
			assert.True(t, s.Keeper.IsOperatorAvailable(s.Ctx, availableValidator, keyIDs...))
			assert.Contains(t, s.Keeper.GetAvailableOperators(s.Ctx, keyIDs...), availableValidator)

			// unknown keys
			unknownKeys := randKeyIDs()
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx, unknownKeys...), 0)
			assert.NotContains(t, s.Keeper.GetAvailableOperators(s.Ctx, unknownKeys...), availableValidator)
			assert.False(t, s.Keeper.IsOperatorAvailable(s.Ctx, availableValidator, unknownKeys...))
		}
	}).Repeat(20))

	t.Run("operator available (edge case #2)", testutils.Func(func(t *testing.T) {
		s := setup()
		keyIDs := randKeyIDs()

		eventHeight := s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx) * rand.I64Between(1, 10)
		repeats := int(rand.I64Between(5, 20))

		for i := 0; i < repeats; i++ {
			availableValidator := rand.ValAddr()
			s.Ctx = s.Ctx.WithBlockHeight(eventHeight)

			// not yet available
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx), i)
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx, keyIDs...), i)

			// available
			s.Keeper.SetAvailableOperator(s.Ctx, availableValidator, keyIDs...)
			s.Ctx = s.Ctx.WithBlockHeight(eventHeight)
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx), i+1)
			assert.True(t, s.Keeper.IsOperatorAvailable(s.Ctx, availableValidator))
			assert.Contains(t, s.Keeper.GetAvailableOperators(s.Ctx), availableValidator)
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx, keyIDs...), i+1)
			assert.True(t, s.Keeper.IsOperatorAvailable(s.Ctx, availableValidator, keyIDs...))
			assert.Contains(t, s.Keeper.GetAvailableOperators(s.Ctx, keyIDs...), availableValidator)

			// unknown keys
			unknownKeys := randKeyIDs()
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx, unknownKeys...), 0)
			assert.NotContains(t, s.Keeper.GetAvailableOperators(s.Ctx, unknownKeys...), availableValidator)
			assert.False(t, s.Keeper.IsOperatorAvailable(s.Ctx, availableValidator, unknownKeys...))
		}
	}).Repeat(20))

	t.Run("operator unavailable", testutils.Func(func(t *testing.T) {
		s := setup()
		keyIDs := randKeyIDs()

		eventHeight := s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx) * rand.I64Between(1, 10)
		height := eventHeight - (rand.I64Between(1, 100) + s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx))
		repeats := int(rand.I64Between(5, 20))

		t.Logf("next event height %d, current height %d", eventHeight, height)

		for i := 0; i < repeats; i++ {
			unavailableValidator := rand.ValAddr()
			s.Ctx = s.Ctx.WithBlockHeight(height)

			// never available
			s.Keeper.SetAvailableOperator(s.Ctx, unavailableValidator, keyIDs...)
			s.Ctx = s.Ctx.WithBlockHeight(eventHeight)
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx), 0)
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx, keyIDs...), 0)
			assert.False(t, s.Keeper.IsOperatorAvailable(s.Ctx, unavailableValidator))
			assert.False(t, s.Keeper.IsOperatorAvailable(s.Ctx, unavailableValidator, keyIDs...))
		}
	}).Repeat(20))

	t.Run("operator unavailable (edge case)", testutils.Func(func(t *testing.T) {
		s := setup()
		keyIDs := randKeyIDs()

		eventHeight := s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx) * rand.I64Between(1, 10)
		height := eventHeight - (1 + s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx))
		repeats := int(rand.I64Between(5, 20))

		for i := 0; i < repeats; i++ {
			unavailableValidator := rand.ValAddr()
			s.Ctx = s.Ctx.WithBlockHeight(height)

			// never available
			s.Keeper.SetAvailableOperator(s.Ctx, unavailableValidator, keyIDs...)
			s.Ctx = s.Ctx.WithBlockHeight(eventHeight)
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx), 0)
			assert.Len(t, s.Keeper.GetAvailableOperators(s.Ctx, keyIDs...), 0)
			assert.False(t, s.Keeper.IsOperatorAvailable(s.Ctx, unavailableValidator))
			assert.False(t, s.Keeper.IsOperatorAvailable(s.Ctx, unavailableValidator, keyIDs...))
		}
	}).Repeat(20))
}

func TestActiveOldKeys(t *testing.T) {
	t.Run("testing locked rotation keys", testutils.Func(func(t *testing.T) {
		s := setup()
		chain := evm.Ethereum
		iterations := int(rand.I64Between(2, 10) * s.Keeper.GetKeyUnbondingLockingKeyRotationCount(s.Ctx))
		params := types.DefaultParams()
		params.MaxSignQueueSize = int64(iterations)
		s.Keeper.SetParams(s.Ctx, params)

		// exclude KeyRole external
		role := exported.GetKeyRoles()[int(rand.I64Between(0, int64(len(exported.GetKeyRoles()))-1))]
		var expectedKeys []exported.Key

		for i := 0; i < iterations; i++ {
			expectedMasterKey := s.SetKey(t, s.Ctx, role, exported.Multisig)
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
