package keeper

import (
	"crypto/ecdsa"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/utils/slices"
)

func TestKeeper_StartKeygen_IdAlreadyInUse_ReturnError(t *testing.T) {
	for _, keyID := range randDistinctStr.Distinct().Take(100) {
		s := setup()
		keyInfo := types.KeyInfo{
			KeyID:   exported.KeyID(keyID),
			KeyRole: exported.MasterKey,
			KeyType: exported.Multisig,
		}
		err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
		assert.NoError(t, err)

		err = s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
		assert.Error(t, err)
	}
}

func TestKeeper_AssignNextMasterKey_RotateMasterKey_NewKeyIsSet(t *testing.T) {
	s := setup()

	for i := 0; i < 50; i++ {
		chain := evm.Ethereum
		s.Ctx = s.Ctx.WithBlockTime(time.Now())
		expectedKey := s.SetKey(t, s.Ctx, exported.MasterKey, chain.KeyType)
		timestamp := s.Ctx.BlockTime()
		expectedKey.RotatedAt = &timestamp
		expectedKey.Chain = chain.Name.String()
		expectedKey.RotationCount = int64(i + 1)
		expectedKey.SnapshotCounter = snap.Counter

		assert.NoError(t, s.Keeper.AssignNextKey(s.Ctx, chain, exported.MasterKey, expectedKey.ID))
		assert.NoError(t, s.Keeper.RotateKey(s.Ctx, chain, exported.MasterKey))

		actualKey, ok := s.Keeper.GetCurrentKey(s.Ctx, chain, exported.MasterKey)
		assert.True(t, ok)
		assert.Equal(t, expectedKey, actualKey)
	}
}

func TestKeeper_AssignNextMasterKey_RotateMasterKey_AssignNextSecondaryKey_RotateSecondaryKey(t *testing.T) {
	currHeight := rand2.I64Between(0, 10000000)

	chain := evm.Ethereum
	s := setup()
	s.Ctx = s.Ctx.WithBlockHeight(currHeight)
	s.Ctx = s.Ctx.WithBlockTime(time.Now())
	expectedMasterKey := s.SetKey(t, s.Ctx, exported.MasterKey, chain.KeyType)
	expectedSecondaryKey := s.SetKey(t, s.Ctx, exported.SecondaryKey, chain.KeyType)

	assert.NoError(t, s.Keeper.AssignNextKey(s.Ctx, chain, exported.MasterKey, expectedMasterKey.ID))
	assert.NoError(t, s.Keeper.RotateKey(s.Ctx, chain, exported.MasterKey))

	assert.NoError(t, s.Keeper.AssignNextKey(s.Ctx, chain, exported.SecondaryKey, expectedSecondaryKey.ID))
	assert.NoError(t, s.Keeper.RotateKey(s.Ctx, chain, exported.SecondaryKey))

	timestamp := s.Ctx.BlockTime()
	expectedMasterKey.Role = exported.MasterKey
	expectedMasterKey.RotatedAt = &timestamp
	expectedMasterKey.Chain = chain.Name.String()
	expectedMasterKey.RotationCount = 1
	expectedMasterKey.SnapshotCounter = snap.Counter
	expectedSecondaryKey.Role = exported.SecondaryKey
	expectedSecondaryKey.RotatedAt = &timestamp
	expectedSecondaryKey.Chain = chain.Name.String()
	expectedSecondaryKey.RotationCount = 1
	expectedSecondaryKey.SnapshotCounter = snap.Counter

	actualMasterKey, ok := s.Keeper.GetCurrentKey(s.Ctx, chain, exported.MasterKey)
	assert.True(t, ok)
	assert.Equal(t, expectedMasterKey, actualMasterKey)

	actualSecondaryKey, ok := s.Keeper.GetCurrentKey(s.Ctx, chain, exported.SecondaryKey)
	assert.True(t, ok)
	assert.Equal(t, expectedSecondaryKey, actualSecondaryKey)
}

func TestKeeper_AssignNextMasterKey_RotateMasterKey_MultipleTimes_PreviousKeysStillAvailable(t *testing.T) {
	for i := 0; i < 20; i++ {
		chain := evm.Ethereum
		s := setup()
		ctx := s.Ctx
		keys := make([]exported.Key, 10)

		for i := range keys {
			snapshotHeight := ctx.BlockHeight() + rand2.I64Between(0, 100)
			ctx = ctx.WithBlockHeight(snapshotHeight + rand2.I64Between(0, 100))

			key := s.SetKey(t, ctx, exported.MasterKey, chain.KeyType)
			keys[i] = key

			assert.NoError(t, s.Keeper.AssignNextKey(ctx, chain, exported.MasterKey, key.ID))
			assert.NoError(t, s.Keeper.RotateKey(ctx, chain, exported.MasterKey))
		}

		// sanity check that the latest key is the last that was set
		actualKey, ok := s.Keeper.GetCurrentKey(ctx, chain, exported.MasterKey)
		assert.True(t, ok)

		actualPubKey, _ := actualKey.GetECDSAPubKey()
		expectedPubKey, _ := keys[len(keys)-1].GetECDSAPubKey()
		assert.Equal(t, expectedPubKey, actualPubKey)

		for _, key := range keys {
			actualKey, ok = s.Keeper.GetKey(ctx, key.ID)
			pubKey, _ := key.GetECDSAPubKey()
			assert.True(t, ok)
			assert.Equal(t, pubKey, actualPubKey)
		}
	}
}

func TestMultisigKeygen(t *testing.T) {
	repeats := 20
	t.Run("should set multisig keygen info when start multisig keygen", testutils.Func(func(t *testing.T) {
		s := setup()
		keyID := rand2.StrBetween(5, 20)
		keyInfo := types.KeyInfo{
			KeyID:   exported.KeyID(keyID),
			KeyRole: exported.MasterKey,
			KeyType: exported.Multisig,
		}

		err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, randSnapshot())
		assert.NoError(t, err)
		keyRequirement, _ := s.Keeper.GetKeyRequirement(s.Ctx, exported.MasterKey, exported.Multisig)
		expectedTimeoutBlock := s.Ctx.BlockHeight() + keyRequirement.KeygenTimeout

		keygenInfo, ok := s.Keeper.GetMultisigKeygenInfo(s.Ctx, exported.KeyID(keyID))
		assert.True(t, ok)
		assert.Equal(t, expectedTimeoutBlock, keygenInfo.GetTimeoutBlock())
		assert.Equal(t, 0, len(keygenInfo.GetKeys()))
		assert.Equal(t, int64(0), keygenInfo.Count())
		assert.False(t, s.Keeper.IsMultisigKeygenCompleted(s.Ctx, exported.KeyID(keyID)))

	}).Repeat(repeats))

	t.Run("should update pub key count and save pub keys when validator submits pub keys", testutils.Func(func(t *testing.T) {
		s := setup()
		keyID := exported.KeyID(rand2.StrBetween(5, 20))
		keyInfo := types.KeyInfo{
			KeyID:   keyID,
			KeyRole: exported.MasterKey,
			KeyType: exported.Multisig,
		}

		err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, randSnapshot())
		assert.NoError(t, err)

		pubKeysCount := int64(0)
		for _, v := range snap.Validators {
			// generate random pub keys for validator v
			pubKeys := slices.Expand(func(idx int) ecdsa.PublicKey { return generatePubKey() }, int(v.ShareCount))
			serializedPubKeys := slices.Map(pubKeys, func(pk ecdsa.PublicKey) []byte { pk2 := btcec.PublicKey(pk); return pk2.SerializeCompressed() })

			pubKeysCount += int64(len(pubKeys))
			s.Keeper.SubmitPubKeys(s.Ctx, keyID, v.GetSDKValidator().GetOperator(), serializedPubKeys...)

			keygenInfo, ok := s.Keeper.GetMultisigKeygenInfo(s.Ctx, keyID)
			assert.True(t, ok)
			assert.Equal(t, pubKeysCount, keygenInfo.Count())
			assert.True(t, keygenInfo.DoesParticipate(v.GetSDKValidator().GetOperator()))

			// verify that the validator's pub keys in the keeper are correct
			actualPubKeys, ok := s.Keeper.GetMultisigPubKeysByValidator(s.Ctx, keyID, v.GetSDKValidator().GetOperator())
			assert.True(t, ok)
			assert.Equal(t, pubKeys, actualPubKeys)
		}

		// fail to retrieve pub keys for a validator that wasn't in the snapshot
		val := newValidator(rand.ValAddr(), 100)
		pubKeys, ok := s.Keeper.GetMultisigPubKeysByValidator(s.Ctx, keyID, val.GetSDKValidator().GetOperator())
		assert.False(t, ok)
		assert.Nil(t, pubKeys)
	}).Repeat(repeats))
}

func randSnapshot() snapshot.Snapshot {
	var validators []snapshot.Validator

	totalShareCount := int64(0)
	for i := int64(0); i < rand2.I64Between(5, 50); i++ {
		share := rand2.I64Between(1, 10)
		validators = append(validators, newValidator(rand2.ValAddr(), share))
		totalShareCount += share
	}

	return snapshot.Snapshot{
		Validators:      validators,
		Timestamp:       time.Now(),
		Height:          rand2.I64Between(1, 1000000),
		TotalShareCount: sdk.NewInt(totalShareCount),
		Counter:         rand2.I64Between(0, 100000),
	}
}
