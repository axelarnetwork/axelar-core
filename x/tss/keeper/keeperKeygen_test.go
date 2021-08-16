package keeper

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strconv"
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

		err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, snap)
		assert.NoError(t, err)

		err = s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, snap)
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

		s.SetLockingPeriod(lockingPeriod)

		keyID := randDistinctStr.Next()
		err := s.Keeper.StartKeygen(ctx, s.Voter, keyID, snap)
		assert.NoError(t, err)

		// time passes
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + rand2.I64Between(0, 2*lockingPeriod))

		sk, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		s.Keeper.SetKey(ctx, keyID, sk.PublicKey)
		chain := evm.Ethereum

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
		chain := evm.Ethereum
		s := setup()
		time := time.Unix(time.Now().Unix(), 0)
		s.Ctx = s.Ctx.WithBlockHeight(currHeight)
		s.Ctx = s.Ctx.WithBlockTime(time)
		s.SetLockingPeriod(lockingPeriod)
		expectedKey := s.SetKey(t, s.Ctx)
		expectedKey.Role = exported.MasterKey
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
	expectedMasterKey := s.SetKey(t, s.Ctx)
	expectedSecondaryKey := s.SetKey(t, s.Ctx)

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

func TestScheduleKeygenAtHeight(t *testing.T) {
	t.Run("testing schedule keygen", testutils.Func(func(t *testing.T) {
		s := setup()
		sender := rand2.AccAddr()
		policies := []exported.KeyShareDistributionPolicy{exported.WeightedByStake, exported.OnePerValidator}
		numReqs := int(rand2.I64Between(10, 30))
		currentHeight := s.Ctx.BlockHeight()
		expectedReqs := make([]types.StartKeygenRequest, numReqs)

		// schedule keygens
		for i := 0; i < numReqs; i++ {
			index := int(rand2.I64Between(0, int64(len(policies)-1)))
			keyID := rand2.StrBetween(5, 10)
			req := types.NewStartKeygenRequest(sender, keyID, int64(len(snap.Validators)), policies[index])
			expectedReqs[i] = *req
			height, err := s.Keeper.ScheduleKeygen(s.Ctx, *req)

			assert.NoError(t, err)
			assert.Equal(t, s.Keeper.GetParams(s.Ctx).AckWindowInBlocks+currentHeight, height)

			height, err = s.Keeper.ScheduleKeygen(s.Ctx, *req)
			assert.EqualError(t, err, fmt.Sprintf("keygen for key ID '%s' already set", req.NewKeyID))
		}

		// verify keygens from above
		s.Ctx = s.Ctx.WithBlockHeight(currentHeight + s.Keeper.GetParams(s.Ctx).AckWindowInBlocks)
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
			s.Keeper.DeleteScheduledKeygen(s.Ctx, req.NewKeyID)
			reqs := s.Keeper.GetAllKeygenRequestsAtCurrentHeight(s.Ctx)
			assert.Len(t, reqs, actualNumReqs-(i+1))
		}
	}).Repeat(20))
}

func TestScheduleKeygenEvents(t *testing.T) {
	t.Run("testing scheduled keygen events", testutils.Func(func(t *testing.T) {
		s := setup()
		currentHeight := s.Ctx.BlockHeight()
		keyID := rand2.Str(20)
		policies := []exported.KeyShareDistributionPolicy{exported.WeightedByStake, exported.OnePerValidator}
		index := int(rand2.I64Between(0, int64(len(policies)-1)))
		height, err := s.Keeper.ScheduleKeygen(s.Ctx, types.StartKeygenRequest{
			Sender:                     rand2.AccAddr(),
			NewKeyID:                   keyID,
			SubsetSize:                 rand2.I64Between(5, 10),
			KeyShareDistributionPolicy: policies[index],
		})
		assert.NoError(t, err)
		assert.Equal(t, s.Keeper.GetParams(s.Ctx).AckWindowInBlocks+currentHeight, height)

		assert.Len(t, s.Ctx.EventManager().ABCIEvents(), 1)
		assert.Equal(t, s.Ctx.EventManager().ABCIEvents()[0].Type, types.EventTypeAck)

		var heightFound, keyIDFound bool
		for _, attribute := range s.Ctx.EventManager().ABCIEvents()[0].Attributes {
			switch string(attribute.Key) {
			case types.AttributeKeyHeight:
				if string(attribute.Value) == strconv.FormatInt(height, 10) {
					heightFound = true
				}
			case types.AttributeKeyKeyID:
				if string(attribute.Value) == keyID {
					keyIDFound = true
				}
			}
		}

		assert.True(t, heightFound)
		assert.True(t, keyIDFound)
	}).Repeat(20))
}
