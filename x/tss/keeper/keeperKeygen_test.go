package keeper

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func TestKeeper_StartKeygen_IdAlreadyInUse_ReturnError(t *testing.T) {
	for _, keyID := range randDistinctStr.Distinct().Take(100) {
		s := setup()
		keyInfo := types.KeyInfo{
			KeyID:   exported.KeyID(keyID),
			KeyRole: exported.MasterKey,
			KeyType: exported.Threshold,
		}
		err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
		assert.NoError(t, err)

		err = s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
		assert.Error(t, err)
	}
}

func TestKeeper_AssignNextMasterKey_StartKeygenAfterLockingPeriod_Unlocked(t *testing.T) {
	for _, currHeight := range randPosInt.Take(100) {
		s := setup()
		ctx := s.Ctx.WithBlockHeight(currHeight)

		// snapshotHeight + lockingPeriod <= currHeight
		lockingPeriod := rand2.I64Between(0, currHeight+1)
		snapshotHeight := rand2.I64Between(0, currHeight-lockingPeriod+1)
		assert.GreaterOrEqual(t, currHeight, snapshotHeight+lockingPeriod)

		keyID := randDistinctStr.Next()
		keyInfo := types.KeyInfo{
			KeyID:   exported.KeyID(keyID),
			KeyRole: exported.MasterKey,
			KeyType: exported.Threshold,
		}
		err := s.Keeper.StartKeygen(ctx, s.Voter, keyInfo, snap)
		assert.NoError(t, err)

		// time passes
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + rand2.I64Between(0, 2*lockingPeriod))

		sk, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		s.Keeper.SetKey(ctx, exported.KeyID(keyID), sk.PublicKey)
		chain := evm.Ethereum

		assert.NoError(t, s.Keeper.AssignNextKey(ctx, chain, exported.MasterKey, exported.KeyID(keyID)))
	}
}

func TestKeeper_AssignNextMasterKey_RotateMasterKey_NewKeyIsSet(t *testing.T) {
	// snapshotHeight + lockingPeriod <= currHeight
	currHeight := rand2.I64Between(0, 10000000)
	lockingPeriod := rand2.I64Between(0, currHeight+1)
	snapshotHeight := rand2.I64Between(0, currHeight-lockingPeriod+1)
	assert.GreaterOrEqual(t, currHeight, snapshotHeight+lockingPeriod)

	for i := 0; i < 100; i++ {
		chain := evm.Ethereum
		s := setup()
		time := time.Unix(time.Now().Unix(), 0)
		s.Ctx = s.Ctx.WithBlockHeight(currHeight)
		s.Ctx = s.Ctx.WithBlockTime(time)
		expectedKey := s.SetKey(t, s.Ctx, exported.MasterKey)
		expectedKey.RotatedAt = &time

		assert.NoError(t, s.Keeper.AssignNextKey(s.Ctx, chain, exported.MasterKey, expectedKey.ID))
		assert.NoError(t, s.Keeper.RotateKey(s.Ctx, chain, exported.MasterKey))

		actualKey, ok := s.Keeper.GetCurrentKey(s.Ctx, chain, exported.MasterKey)
		assert.True(t, ok)
		assert.Equal(t, expectedKey, actualKey)
	}
}

func TestKeeper_AssignNextMasterKey_RotateMasterKey_AssignNextSecondaryKey_RotateSecondaryKey(t *testing.T) {
	currHeight := rand2.I64Between(0, 10000000)

	chain := bitcoin.Bitcoin
	s := setup()
	time := time.Unix(time.Now().Unix(), 0)
	s.Ctx = s.Ctx.WithBlockHeight(currHeight)
	s.Ctx = s.Ctx.WithBlockTime(time)
	expectedMasterKey := s.SetKey(t, s.Ctx, exported.MasterKey)
	expectedSecondaryKey := s.SetKey(t, s.Ctx, exported.SecondaryKey)

	assert.NoError(t, s.Keeper.AssignNextKey(s.Ctx, chain, exported.MasterKey, expectedMasterKey.ID))
	assert.NoError(t, s.Keeper.RotateKey(s.Ctx, chain, exported.MasterKey))

	assert.NoError(t, s.Keeper.AssignNextKey(s.Ctx, chain, exported.SecondaryKey, expectedSecondaryKey.ID))
	assert.NoError(t, s.Keeper.RotateKey(s.Ctx, chain, exported.SecondaryKey))

	expectedMasterKey.Role = exported.MasterKey
	expectedMasterKey.RotatedAt = &time
	expectedSecondaryKey.Role = exported.SecondaryKey
	expectedSecondaryKey.RotatedAt = &time

	actualMasterKey, ok := s.Keeper.GetCurrentKey(s.Ctx, chain, exported.MasterKey)
	assert.True(t, ok)
	assert.Equal(t, expectedMasterKey, actualMasterKey)

	actualSecondaryKey, ok := s.Keeper.GetCurrentKey(s.Ctx, chain, exported.SecondaryKey)
	assert.True(t, ok)
	assert.Equal(t, expectedSecondaryKey, actualSecondaryKey)
}

func TestKeeper_AssignNextMasterKey_RotateMasterKey_MultipleTimes_PreviousKeysStillAvailable(t *testing.T) {
	for i := 0; i < 100; i++ {
		chain := evm.Ethereum
		s := setup()
		ctx := s.Ctx
		keys := make([]exported.Key, 10)

		for i := range keys {
			snapshotHeight := ctx.BlockHeight() + rand2.I64Between(0, 100)
			ctx = ctx.WithBlockHeight(snapshotHeight + rand2.I64Between(0, 100))

			key := s.SetKey(t, ctx, exported.MasterKey)
			keys[i] = key

			assert.NoError(t, s.Keeper.AssignNextKey(ctx, chain, exported.MasterKey, key.ID))
			assert.NoError(t, s.Keeper.RotateKey(ctx, chain, exported.MasterKey))
		}

		// sanity check that the latest key is the last that was set
		actualKey, ok := s.Keeper.GetCurrentKey(ctx, chain, exported.MasterKey)
		assert.True(t, ok)
		assert.Equal(t, keys[len(keys)-1].Value, actualKey.Value)

		for _, key := range keys {
			actualKey, ok = s.Keeper.GetKey(ctx, key.ID)

			assert.True(t, ok)
			assert.Equal(t, key.Value, actualKey.Value)
		}
	}
}

func TestGetKeygenParticipants(t *testing.T) {
	repeats := 20
	t.Run("should return keygen participants when keyID is found and keygen successful", testutils.Func(func(t *testing.T) {
		s := setup()
		snap = randSnapshot()
		keyID := rand2.StrBetween(5, 20)
		keyInfo := types.KeyInfo{
			KeyID:   exported.KeyID(keyID),
			KeyRole: exported.MasterKey,
			KeyType: exported.Threshold,
		}

		err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
		assert.NoError(t, err)

		participants := s.Keeper.GetParticipantsInKeygen(s.Ctx, exported.KeyID(keyID))
		assert.Equal(t, len(snap.Validators), len(participants))

		for _, v := range snap.Validators {
			assert.Contains(t,participants,v.GetSDKValidator().GetOperator())
		}
	}).Repeat(repeats))

	t.Run("should return empty list participants when keyID is not found", testutils.Func(func(t *testing.T) {
		s := setup()
		snap = randSnapshot()
		keyIDs := randDistinctStr.Distinct().Take(2)
		keyInfo := types.KeyInfo{
			KeyID:   exported.KeyID(keyIDs[0]),
			KeyRole: exported.MasterKey,
			KeyType: exported.Threshold,
		}

		err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
		assert.NoError(t, err)

		participants := s.Keeper.GetParticipantsInKeygen(s.Ctx, exported.KeyID(keyIDs[1]))
		assert.Equal(t, 0, len(participants))

	}).Repeat(repeats))
}

func TestMultisigKeygen(t *testing.T) {
	repeats := 20
	t.Run("should set keygen timeout when start multisig keygen", testutils.Func(func(t *testing.T) {
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
		timeout, ok := s.Keeper.GetMultisigPubKeyTimeout(s.Ctx, exported.KeyID(keyID))

		assert.True(t, ok)
		assert.Equal(t, expectedTimeoutBlock, timeout)
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
			// random pub keys
			var pubKeys [][]byte
			for i := int64(0); i < v.ShareCount; i++ {
				pk := btcec.PublicKey(generatePubKey())
				pubKeys = append(pubKeys, pk.SerializeCompressed())
			}
			pubKeysCount += int64(len(pubKeys))
			s.Keeper.SubmitPubKeys(s.Ctx, keyID, v.GetSDKValidator().GetOperator(), pubKeys...)
			assert.Equal(t, pubKeysCount, s.Keeper.GetMultisigPubKeyCount(s.Ctx, keyID))
			assert.True(t, s.Keeper.HasValidatorSubmittedMultisigPubKey(s.Ctx, keyID, v.GetSDKValidator().GetOperator()))
		}
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
