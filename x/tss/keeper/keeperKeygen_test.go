package keeper

import (
	"crypto/ecdsa"
	"crypto/rand"
	"testing"

	"github.com/btcsuite/btcd/btcec"

	eth "github.com/axelarnetwork/axelar-core/x/ethereum/exported"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func TestKeeper_StartKeygen_IdAlreadyInUse_ReturnError(t *testing.T) {
	for _, keyID := range randDistinctStr.Distinct().Take(100) {
		s := setup(t)
		err := s.Keeper.StartKeygen(s.Ctx, keyID, 1, snap)
		assert.NoError(t, err)
		err = s.Keeper.StartKeygen(s.Ctx, keyID, 1, snap)
		assert.Error(t, err)
	}
}

func TestKeeper_KeygenMsg_NoSessionWithGivenID_Error(t *testing.T) {
	for _, keyID := range randDistinctStr.Take(100) {
		s := setup(t)
		assert.Error(t, s.Keeper.KeygenMsg(s.Ctx, types.MsgKeygenTraffic{
			Sender:    s.Broadcaster.GetProxy(s.Ctx, s.Broadcaster.LocalPrincipal),
			SessionID: keyID,
			Payload:   &tofnd.TrafficOut{},
		}))
	}
}

func TestKeeper_AssignNextMasterKey_StartKeygenDuringLockingPeriod_Locked(t *testing.T) {
	for _, currHeight := range randPosInt.Take(100) {
		s := setup(t)
		ctx := s.Ctx.WithBlockHeight(currHeight)

		// snapshotHeight + lockingPeriod > currHeight
		lockingPeriod := testutils.RandIntBetween(0, currHeight+1)
		snapshotHeight := testutils.RandIntBetween(currHeight-lockingPeriod+1, currHeight+1)
		assert.Less(t, currHeight, snapshotHeight+lockingPeriod)

		s.SetLockingPeriod(lockingPeriod)

		keyID := randDistinctStr.Next()
		err := s.Keeper.StartKeygen(ctx, keyID, len(validators)-1, snap)
		assert.NoError(t, err)

		// time passes
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + testutils.RandIntBetween(0, 2*lockingPeriod))

		sk, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		s.Keeper.SetKey(ctx, keyID, sk.PublicKey)
		chain := eth.Ethereum

		assert.Errorf(
			t,
			s.Keeper.AssignNextMasterKey(ctx, chain, snapshotHeight, keyID),
			"snapshot: %d, lock: %d, height: %d, difference: %d",
			snapshotHeight,
			lockingPeriod,
			ctx.BlockHeight(),
			ctx.BlockHeight()-lockingPeriod-snapshotHeight,
		)
	}
}

func TestKeeper_AssignNextMasterKey_StartKeygenAfterLockingPeriod_Unlocked(t *testing.T) {
	for _, currHeight := range randPosInt.Take(100) {
		s := setup(t)
		ctx := s.Ctx.WithBlockHeight(currHeight)

		// snapshotHeight + lockingPeriod <= currHeight
		lockingPeriod := testutils.RandIntBetween(0, currHeight+1)
		snapshotHeight := testutils.RandIntBetween(0, currHeight-lockingPeriod+1)
		assert.GreaterOrEqual(t, currHeight, snapshotHeight+lockingPeriod)

		s.SetLockingPeriod(lockingPeriod)

		keyID := randDistinctStr.Next()
		err := s.Keeper.StartKeygen(ctx, keyID, len(validators)-1, snap)
		assert.NoError(t, err)

		// time passes
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + testutils.RandIntBetween(0, 2*lockingPeriod))

		sk, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		s.Keeper.SetKey(ctx, keyID, sk.PublicKey)
		chain := eth.Ethereum

		assert.NoError(t, s.Keeper.AssignNextMasterKey(ctx, chain, snapshotHeight, keyID))
	}
}

func TestKeeper_AssignNextMasterKey_RotateMasterKey_NewKeyIsSet(t *testing.T) {
	// snapshotHeight + lockingPeriod <= currHeight
	currHeight := testutils.RandIntBetween(0, 10000000)
	lockingPeriod := testutils.RandIntBetween(0, currHeight+1)
	snapshotHeight := testutils.RandIntBetween(0, currHeight-lockingPeriod+1)
	assert.GreaterOrEqual(t, currHeight, snapshotHeight+lockingPeriod)

	for i := 0; i < 100; i++ {
		chain := eth.Ethereum
		s := setup(t)
		ctx := s.Ctx.WithBlockHeight(currHeight)
		s.SetLockingPeriod(lockingPeriod)
		keyID, expectedKey := s.SetKey(t, ctx)

		assert.NoError(t, s.Keeper.AssignNextMasterKey(ctx, chain, snapshotHeight, keyID))
		assert.NoError(t, s.Keeper.RotateMasterKey(s.Ctx, chain))

		actualKey, ok := s.Keeper.GetCurrentMasterKey(s.Ctx, chain)
		assert.True(t, ok)
		assert.Equal(t, expectedKey, actualKey)
	}
}

func TestKeeper_AssignNextMasterKey_RotateMasterKey_MultipleTimes_PreviousKeysStillAvailable(t *testing.T) {
	for i := 0; i < 100; i++ {
		chain := eth.Ethereum
		s := setup(t)
		s.SetLockingPeriod(0)
		ctx := s.Ctx
		masterKeys := make([]ecdsa.PublicKey, 10)
		for i := range masterKeys {

			snapshotHeight := ctx.BlockHeight() + testutils.RandIntBetween(0, 100)
			ctx = ctx.WithBlockHeight(snapshotHeight + testutils.RandIntBetween(0, 100))

			keyID, pk := s.SetKey(t, ctx)
			masterKeys[i] = pk

			assert.NoError(t, s.Keeper.AssignNextMasterKey(ctx, chain, snapshotHeight, keyID))
			assert.NoError(t, s.Keeper.RotateMasterKey(ctx, chain))
		}

		// sanity check that the latest key is the last that was set
		actualKey, ok := s.Keeper.GetCurrentMasterKey(ctx, chain)
		assert.True(t, ok)
		assert.Equal(t, masterKeys[len(masterKeys)-1], actualKey)

		for i, key := range masterKeys {
			actualKey, ok = s.Keeper.GetPreviousMasterKey(ctx, chain, int64(len(masterKeys)-i-1))
			assert.True(t, ok)
			assert.Equal(t, key, actualKey)
		}
	}
}
