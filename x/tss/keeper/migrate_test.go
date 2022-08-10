package keeper

import (
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigKeeper "github.com/axelarnetwork/axelar-core/x/multisig/keeper"
	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/tss/types/mock"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
	rand2 "github.com/axelarnetwork/utils/test/rand"
)

func TestGetMigrationHandler(t *testing.T) {
	encCfg := params.MakeEncodingConfig()
	paramsStoreKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	paramsTSstoreKey := sdk.NewKVStoreKey(paramstypes.TStoreKey)

	val1 := newValidator(rand.ValAddr(), 10)
	val2 := newValidator(rand.ValAddr(), 10)
	val3 := newValidator(rand.ValAddr(), 10)
	val4 := newValidator(rand.ValAddr(), 10)
	validators := []snapshot.Validator{val1, val2, val3, val4}
	snap := snapshot.Snapshot{
		Validators:      validators,
		Timestamp:       time.Now(),
		Height:          rand.I64Between(1, 1000000),
		TotalShareCount: sdk.NewInt(40),
		Counter:         rand.I64Between(0, 100000),
	}

	chains := []nexus.Chain{
		{
			Name:   "ethereum",
			Module: evm.ModuleName,
		},
		{
			Name:   "avalanche",
			Module: evm.ModuleName,
		},
		{
			Name:   "osmosis",
			Module: axelarnet.ModuleName,
		},
	}
	var (
		ctx          sdk.Context
		k            Keeper
		msk          multisigKeeper.Keeper
		handler      func(ctx sdk.Context) error
		expectedKeys []exported.Key
	)

	givenSetup := Given("a context", func() {
		ctx = rand.Context(fake.NewMultiStore())
		expectedKeys = []exported.Key{}
	}).
		Given("a tss keeper", func() {
			subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, paramsStoreKey, paramsTSstoreKey, types.ModuleName)
			k = NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(types.StoreKey), subspace, &mock.SlasherMock{}, &mock.RewarderMock{})

			k.SetParams(ctx, types.DefaultParams())
		}).
		Given("a multisig keeper", func() {
			subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, paramsStoreKey, paramsTSstoreKey, multisigTypes.ModuleName)
			msk = multisigKeeper.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(multisigTypes.StoreKey), subspace)
			msk.InitGenesis(ctx, multisigTypes.DefaultGenesisState())
		}).
		Given("a migration handler", func() {
			n := &mock.NexusMock{GetChainsFunc: func(ctx sdk.Context) []nexus.Chain { return chains }}

			snapshotter := &mock.SnapshotterMock{
				GetSnapshotFunc: func(ctx sdk.Context, seqNo int64) (snapshot.Snapshot, bool) { return snap, true },
			}
			handler = GetMigrationHandler(k, msk, n, snapshotter)
		})

	givenSetup.When("there are no keys to migrate", func() {}).
		Then("do not fail", func(t *testing.T) {
			assert.NoError(t, handler(ctx))
		}).Run(t)

	givenSetup.
		When("there is only a current key", func() {
			expectedKeys = append(expectedKeys, setKey(ctx, k, chains[1].Name, validators))

			funcs.MustNoErr(k.AssignNextKey(ctx, chains[1], exported.SecondaryKey, expectedKeys[0].ID))
			funcs.MustNoErr(k.RotateKey(ctx, chains[1], exported.SecondaryKey))
		}).
		Then("migrate the current key", func(t *testing.T) {
			assert.NoError(t, handler(ctx))

			keyIDs := msk.GetActiveKeyIDs(ctx, nexus.ChainName(expectedKeys[0].Chain))

			assert.Len(t, keyIDs, 1)
			keyID := keyIDs[0]
			currentKey, ok := funcs.MustOk(msk.GetKey(ctx, keyID)).(*multisigTypes.Key)
			assert.True(t, ok)
			assert.EqualValues(t, expectedKeys[0].ID, currentKey.ID)
			assert.Equal(t, multisig.Active, currentKey.State)
			assert.Equal(t, currentKey, funcs.MustOk(msk.GetCurrentKey(ctx, nexus.ChainName(expectedKeys[0].Chain))))
		}).Run(t)

	givenSetup.
		When("there is only a next key", func() {
			expectedKeys = append(expectedKeys, setKey(ctx, k, chains[0].Name, validators))

			funcs.MustNoErr(k.AssignNextKey(ctx, chains[0], exported.SecondaryKey, expectedKeys[0].ID))
		}).
		Then("migrate the next key", func(t *testing.T) {
			assert.NoError(t, handler(ctx))

			keyID := funcs.MustOk(msk.GetNextKeyID(ctx, nexus.ChainName(expectedKeys[0].Chain)))

			assert.EqualValues(t, expectedKeys[0].ID, keyID)
			nextKey, ok := funcs.MustOk(msk.GetKey(ctx, keyID)).(*multisigTypes.Key)
			assert.True(t, ok)
			assert.EqualValues(t, expectedKeys[0].ID, nextKey.ID)
			assert.Equal(t, multisig.Assigned, nextKey.State)
		}).Run(t)

	givenSetup.
		When("there are old, current and next keys", func() {
			for i := 0; i < 20; i++ {
				key := setKey(ctx, k, chains[2].Name, validators)
				expectedKeys = append(expectedKeys, key)
				funcs.MustNoErr(k.AssignNextKey(ctx, chains[2], exported.SecondaryKey, key.ID))
				funcs.MustNoErr(k.RotateKey(ctx, chains[2], exported.SecondaryKey))
			}
			expectedKeys = append(expectedKeys, setKey(ctx, k, chains[2].Name, validators))
			funcs.MustNoErr(k.AssignNextKey(ctx, chains[2], exported.SecondaryKey, expectedKeys[len(expectedKeys)-1].ID))
		}).
		Then("migrate all keys", func(t *testing.T) {
			assert.NoError(t, handler(ctx))

			keyIDs := msk.GetActiveKeyIDs(ctx, nexus.ChainName(expectedKeys[0].Chain))
			assert.Len(t, keyIDs, int(types.DefaultParams().UnbondingLockingKeyRotationCount+1))

			expectedActiveKeys := slices.Reverse(expectedKeys[len(expectedKeys)-1-len(keyIDs) : len(expectedKeys)-1])
			for i := 1; i < len(keyIDs); i++ {
				oldActiveKey, ok := funcs.MustOk(msk.GetKey(ctx, keyIDs[i])).(*multisigTypes.Key)
				assert.True(t, ok)
				assert.EqualValues(t, expectedActiveKeys[i].ID, oldActiveKey.ID)
				assert.Equal(t, multisig.Active, oldActiveKey.State)
				assert.Len(t, oldActiveKey.GetParticipants(), len(validators))
				assert.Len(t, oldActiveKey.Snapshot.Participants, len(validators))
			}

			currentKey := funcs.MustOk(msk.GetCurrentKey(ctx, nexus.ChainName(expectedActiveKeys[0].Chain))).(*multisigTypes.Key)
			assert.EqualValues(t, expectedActiveKeys[0].ID, currentKey.ID)
			assert.EqualValues(t, currentKey.ID, keyIDs[0])
			assert.Equal(t, multisig.Active, currentKey.State)
			assert.Len(t, currentKey.GetParticipants(), len(validators))
			assert.Len(t, currentKey.Snapshot.Participants, len(validators))

			keyID := funcs.MustOk(msk.GetNextKeyID(ctx, nexus.ChainName(expectedKeys[len(expectedKeys)-1].Chain)))
			nextKey, ok := funcs.MustOk(msk.GetKey(ctx, keyID)).(*multisigTypes.Key)
			assert.True(t, ok)
			assert.EqualValues(t, expectedKeys[len(expectedKeys)-1].ID, nextKey.ID)
			assert.Equal(t, multisig.Assigned, nextKey.State)

			assert.Len(t, nextKey.GetParticipants(), len(validators))

			assert.Len(t, nextKey.Snapshot.Participants, len(validators))
		}).Run(t)
}

func setKey(ctx sdk.Context, k Keeper, chain nexus.ChainName, validators []snapshot.Validator) exported.Key {
	var expectedKey exported.Key
	expectedKey = generateMultisigKey(exported.KeyID(rand2.AlphaStrBetween(5, 10)))
	expectedKey.Chain = string(chain)
	expectedKey.SnapshotCounter = rand.PosI64()
	k.setSnapshotCounterForKeyID(ctx, expectedKey.ID, expectedKey.SnapshotCounter)
	k.SetKey(ctx, expectedKey)
	k.SetMultisigKeygenInfo(ctx, types.MultisigInfo{
		ID:        string(expectedKey.ID),
		Timeout:   rand.PosI64(),
		TargetNum: 1,
	})
	for _, validator := range validators {
		pubKeys := slices.Expand(func(idx int) ecdsa.PublicKey { return generatePubKey() }, 2)
		serializedPubKeys := slices.Map(pubKeys, func(pk ecdsa.PublicKey) []byte { pk2 := btcec.PublicKey(pk); return pk2.SerializeCompressed() })
		k.SubmitPubKeys(ctx, expectedKey.ID, validator.GetSDKValidator().GetOperator(), serializedPubKeys...)
	}
	return expectedKey
}
