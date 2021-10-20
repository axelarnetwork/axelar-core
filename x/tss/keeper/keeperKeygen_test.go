package keeper

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"

	"github.com/axelarnetwork/axelar-core/testutils"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	bitcoin "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func TestKeeper_StartKeygen_IdAlreadyInUse_ReturnError(t *testing.T) {
	for _, keyID := range randDistinctStr.Distinct().Take(100) {
		s := setup()

		err := s.Keeper.StartKeygen(s.Ctx, s.Voter, exported.KeyID(keyID), exported.MasterKey, snap)
		assert.NoError(t, err)

		err = s.Keeper.StartKeygen(s.Ctx, s.Voter, exported.KeyID(keyID), exported.MasterKey, snap)
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
		err := s.Keeper.StartKeygen(ctx, s.Voter, exported.KeyID(keyID), exported.MasterKey, snap)
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

func TestScheduleKeygenAtHeight(t *testing.T) {
	t.Run("testing schedule keygen", testutils.Func(func(t *testing.T) {
		s := setup()
		sender := rand2.AccAddr()
		numReqs := int(rand2.I64Between(10, 30))
		currentHeight := s.Ctx.BlockHeight()
		expectedReqs := make([]types.StartKeygenRequest, numReqs)

		// schedule keygens
		for i := 0; i < numReqs; i++ {
			keyID := rand2.StrBetween(5, 10)
			req := types.NewStartKeygenRequest(sender, keyID, exported.MasterKey)
			expectedReqs[i] = *req
			height, err := s.Keeper.ScheduleKeygen(s.Ctx, *req)

			assert.NoError(t, err)
			assert.Equal(t, currentHeight, height)

			height, err = s.Keeper.ScheduleKeygen(s.Ctx, *req)
			assert.EqualError(t, err, fmt.Sprintf("keygen for key ID '%s' already set", req.KeyID))
		}

		// verify keygens from above
		//s.Ctx = s.Ctx.WithBlockHeight(currentHeight + s.Keeper.GetParams(s.Ctx).AckPeriodInBlocks)
		reqs := s.Keeper.GetAllKeygenRequestsAtCurrentHeight(s.Ctx)

		actualNumReqs := 0
		for _, expected := range expectedReqs {
			for _, actual := range reqs {
				if bytes.Equal(expected.GetSignBytes(), actual.GetSignBytes()) {
					actualNumReqs++
					break
				}
			}
		}
		assert.Len(t, expectedReqs, actualNumReqs)
		assert.Equal(t, numReqs, actualNumReqs)

		// check that we can delete scheduled keygens
		for i, req := range reqs {
			s.Keeper.DeleteScheduledKeygen(s.Ctx, req.KeyID)
			reqs := s.Keeper.GetAllKeygenRequestsAtCurrentHeight(s.Ctx)
			assert.Len(t, reqs, actualNumReqs-(i+1))
		}
	}).Repeat(20))
}
