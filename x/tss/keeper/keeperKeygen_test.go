package keeper

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"

	"github.com/btcsuite/btcd/btcec"

	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
)

func TestKeeper_StartKeygen_IdAlreadyInUse_ReturnError(t *testing.T) {
	for _, keyID := range randDistinctStr.Distinct().Take(100) {
		s := setup(t)

		err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, 1, snap)
		assert.NoError(t, err)

		err = s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, 1, snap)
		assert.Error(t, err)
	}
}

func TestKeeper_AssignNextMasterKey_StartKeygenAfterLockingPeriod_Unlocked(t *testing.T) {
	for _, currHeight := range randPosInt.Take(100) {
		s := setup(t)
		ctx := s.Ctx.WithBlockHeight(currHeight)

		// snapshotHeight + lockingPeriod <= currHeight
		lockingPeriod := rand2.I64Between(0, currHeight+1)
		snapshotHeight := rand2.I64Between(0, currHeight-lockingPeriod+1)
		assert.GreaterOrEqual(t, currHeight, snapshotHeight+lockingPeriod)

		s.SetLockingPeriod(lockingPeriod)

		keyID := randDistinctStr.Next()
		err := s.Keeper.StartKeygen(ctx, s.Voter, keyID, len(validators)-1, snap)
		assert.NoError(t, err)

		// time passes
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + rand2.I64Between(0, 2*lockingPeriod))

		sk, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		s.Keeper.SetKey(ctx, keyID, sk.PublicKey)
		chain := eth.Ethereum

		assert.NoError(t, s.Keeper.AssignNextKey(ctx, chain, exported.MasterKey, keyID))
	}
}

func TestKeeper_AssignNextMasterKey_RotateMasterKey_NewKeyIsSet(t *testing.T) {
	// snapshotHeight + lockingPeriod <= currHeight
	currHeight := rand2.I64Between(0, 10000000)
	lockingPeriod := rand2.I64Between(0, currHeight+1)
	snapshotHeight := rand2.I64Between(0, currHeight-lockingPeriod+1)
	assert.GreaterOrEqual(t, currHeight, snapshotHeight+lockingPeriod)

	for i := 0; i < 100; i++ {
		chain := eth.Ethereum
		s := setup(t)
		ctx := s.Ctx.WithBlockHeight(currHeight)
		s.SetLockingPeriod(lockingPeriod)
		expectedKey := s.SetKey(t, ctx)

		assert.NoError(t, s.Keeper.AssignNextKey(ctx, chain, exported.MasterKey, expectedKey.ID))
		assert.NoError(t, s.Keeper.RotateKey(s.Ctx, chain, exported.MasterKey))

		actualKey, ok := s.Keeper.GetCurrentKey(s.Ctx, chain, exported.MasterKey)
		assert.True(t, ok)
		assert.Equal(t, expectedKey, actualKey)
	}
}

func TestKeeper_AssignNextMasterKey_RotateMasterKey_AssignNextSecondaryKey_RotateSecondaryKey(t *testing.T) {
	currHeight := rand2.I64Between(0, 10000000)

	chain := bitcoin.Bitcoin
	s := setup(t)
	ctx := s.Ctx.WithBlockHeight(currHeight)
	expectedMasterKey := s.SetKey(t, ctx)
	expectedSecondaryKey := s.SetKey(t, ctx)

	assert.NoError(t, s.Keeper.AssignNextKey(ctx, chain, exported.MasterKey, expectedMasterKey.ID))
	assert.NoError(t, s.Keeper.RotateKey(s.Ctx, chain, exported.MasterKey))

	assert.NoError(t, s.Keeper.AssignNextKey(ctx, chain, exported.SecondaryKey, expectedSecondaryKey.ID))
	assert.NoError(t, s.Keeper.RotateKey(s.Ctx, chain, exported.SecondaryKey))

	actualMasterKey, ok := s.Keeper.GetCurrentKey(s.Ctx, chain, exported.MasterKey)
	assert.True(t, ok)
	assert.Equal(t, expectedMasterKey, actualMasterKey)

	actualSecondaryKey, ok := s.Keeper.GetCurrentKey(s.Ctx, chain, exported.SecondaryKey)
	assert.True(t, ok)
	assert.Equal(t, expectedSecondaryKey, actualSecondaryKey)
}

func TestKeeper_AssignNextMasterKey_RotateMasterKey_MultipleTimes_PreviousKeysStillAvailable(t *testing.T) {
	for i := 0; i < 100; i++ {
		chain := eth.Ethereum
		s := setup(t)
		s.SetLockingPeriod(0)
		ctx := s.Ctx
		keys := make([]exported.Key, 10)

		for i := range keys {
			snapshotHeight := ctx.BlockHeight() + rand2.I64Between(0, 100)
			ctx = ctx.WithBlockHeight(snapshotHeight + rand2.I64Between(0, 100))

			key := s.SetKey(t, ctx)
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
