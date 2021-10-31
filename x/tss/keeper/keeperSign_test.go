package keeper

import (
	"crypto/ecdsa"
	rand3 "crypto/rand"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func TestStartSign_EnoughActiveValidators(t *testing.T) {
	s := setup()
	sigID := "sigID"
	keyID := exported.KeyID("keyID")
	msg := []byte("message")
	val1 := rand.ValAddr()
	val2 := rand.ValAddr()
	val3 := rand.ValAddr()
	val4 := rand.ValAddr()
	val5 := rand.ValAddr()

	snap := snapshot.Snapshot{
		Validators: []snapshot.Validator{
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val1 },
				GetConsensusPowerFunc: func(sdk.Int) int64 { return 140 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val1.Bytes(), nil },
				IsJailedFunc:          func() bool { return true },
				StringFunc:            func() string { return val1.String() },
			}, 140),
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val2 },
				GetConsensusPowerFunc: func(sdk.Int) int64 { return 130 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val2.Bytes(), nil },
				IsJailedFunc:          func() bool { return false },
				StringFunc:            func() string { return val2.String() },
			}, 130),
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val3 },
				GetConsensusPowerFunc: func(sdk.Int) int64 { return 120 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val3.Bytes(), nil },
				IsJailedFunc:          func() bool { return false },
				StringFunc:            func() string { return val3.String() },
			}, 120),
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val4 },
				GetConsensusPowerFunc: func(sdk.Int) int64 { return 110 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val4.Bytes(), nil },
				IsJailedFunc:          func() bool { return false },
				StringFunc:            func() string { return val4.String() },
			}, 110),
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val5 },
				GetConsensusPowerFunc: func(sdk.Int) int64 { return 100 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val5.Bytes(), nil },
				IsJailedFunc:          func() bool { return false },
				StringFunc:            func() string { return val5.String() },
			}, 100),
		},
		Timestamp:       time.Now(),
		Height:          rand2.I64Between(1, 1000000),
		TotalShareCount: sdk.NewInt(600),
		Counter:         rand2.I64Between(0, 100000),
	}
	snap.CorruptionThreshold = exported.ComputeAbsCorruptionThreshold(utils.Threshold{Numerator: 2, Denominator: 3}, snap.TotalShareCount)
	assert.Equal(t, int64(399), snap.CorruptionThreshold)
	s.Snapshotter.GetValidatorIllegibilityFunc = func(ctx sdk.Context, validator snapshot.SDKValidator) (snapshot.ValidatorIllegibility, error) {
		return snapshot.None, nil
	}
	s.Snapshotter.GetSnapshotFunc = func(ctx sdk.Context, seqNo int64) (snapshot.Snapshot, bool) {
		if seqNo == snap.Counter {
			return snap, true
		}
		return snapshot.Snapshot{}, false
	}

	height := s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx) * rand.I64Between(1, 10)
	height += rand.I64Between(0, s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx))
	s.Ctx = s.Ctx.WithBlockHeight(height)

	for _, val := range snap.Validators {
		s.Keeper.SetAvailableOperator(s.Ctx, val.GetSDKValidator().GetOperator(), keyID)
	}

	// start keygen to record the snapshot for each key
	keyInfo := types.KeyInfo{
		KeyID:   keyID,
		KeyRole: exported.MasterKey,
		KeyType: exported.Threshold,
	}
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
	assert.NoError(t, err)
	s.Keeper.SetKey(s.Ctx, exported.KeyID(keyID), generatePubKey())

	sigInfo := exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msg,
		SnapshotCounter: snap.Counter,
	}
	err = s.Keeper.StartSign(s.Ctx, sigInfo, s.Snapshotter, s.Voter)
	assert.NoError(t, err)

	participants, active, err := s.Keeper.SelectSignParticipants(s.Ctx, s.Snapshotter, sigInfo, snap)

	signingShareCount := sdk.ZeroInt()
	for _, p := range participants {
		signingShareCount = signingShareCount.AddRaw(p.ShareCount)
	}

	activeShareCount := sdk.ZeroInt()
	for _, v := range active {
		activeShareCount = activeShareCount.AddRaw(v.ShareCount)
	}

	assert.NoError(t, err)
	assert.True(t, signingShareCount.GTE(sdk.NewInt(snap.CorruptionThreshold)))
	assert.Equal(t, int64(600), activeShareCount.Int64())
	assert.Equal(t, int64(500), signingShareCount.Int64())
	assert.Equal(t, 4, len(participants))
	assert.Equal(t, 1, len(snap.Validators)-len(participants))
	assert.Equal(t, val1, participants[0].GetSDKValidator().GetOperator())
	assert.Equal(t, val2, participants[1].GetSDKValidator().GetOperator())
	assert.Equal(t, val3, participants[2].GetSDKValidator().GetOperator())
	assert.Equal(t, val4, participants[3].GetSDKValidator().GetOperator())
}

func TestStartSign_NoEnoughActiveValidators(t *testing.T) {
	s := setup()
	sigID := "sigID"
	keyID := exported.KeyID("keyID")
	msg := []byte("message")
	val1 := rand.ValAddr()
	val2 := rand.ValAddr()

	snap := snapshot.Snapshot{
		Validators: []snapshot.Validator{
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val1 },
				GetConsensusPowerFunc: func(sdk.Int) int64 { return 100 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val1.Bytes(), nil },
				IsJailedFunc:          func() bool { return true },
				StringFunc:            func() string { return val1.String() },
			}, 100),
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val2 },
				GetConsensusPowerFunc: func(sdk.Int) int64 { return 100 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val2.Bytes(), nil },
				IsJailedFunc:          func() bool { return false },
				StringFunc:            func() string { return val2.String() },
			}, 100),
		},
		Timestamp:       time.Now(),
		Height:          rand2.I64Between(1, 1000000),
		TotalShareCount: sdk.NewInt(200),
		Counter:         rand2.I64Between(0, 100000),
	}
	snap.CorruptionThreshold = exported.ComputeAbsCorruptionThreshold(utils.Threshold{Numerator: 2, Denominator: 3}, snap.TotalShareCount)
	s.Snapshotter.GetValidatorIllegibilityFunc = func(ctx sdk.Context, validator snapshot.SDKValidator) (snapshot.ValidatorIllegibility, error) {
		if validator.GetOperator().Equals(val1) {
			return snapshot.Jailed, nil
		}

		return snapshot.None, nil
	}
	s.Snapshotter.GetSnapshotFunc = func(ctx sdk.Context, seqNo int64) (snapshot.Snapshot, bool) {
		if seqNo == snap.Counter {
			return snap, true
		}
		return snapshot.Snapshot{}, false
	}

	height := s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx) * rand.I64Between(1, 10)
	height += rand.I64Between(0, s.Keeper.GetHeartbeatPeriodInBlocks(s.Ctx))
	s.Ctx = s.Ctx.WithBlockHeight(height)

	for _, val := range snap.Validators {
		s.Keeper.SetAvailableOperator(s.Ctx, val.GetSDKValidator().GetOperator(), keyID)
	}

	// start keygen to record the snapshot for each key
	keyInfo := types.KeyInfo{
		KeyID:   keyID,
		KeyRole: exported.MasterKey,
		KeyType: exported.Threshold,
	}
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
	assert.NoError(t, err)
	s.Keeper.SetKey(s.Ctx, exported.KeyID(keyID), generatePubKey())

	sigInfo := exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msg,
		SnapshotCounter: snap.Counter,
	}
	err = s.Keeper.StartSign(s.Ctx, sigInfo, s.Snapshotter, s.Voter)
	assert.Error(t, err)

	participants, active, err := s.Keeper.SelectSignParticipants(s.Ctx, s.Snapshotter, sigInfo, snap)

	signingShareCount := sdk.ZeroInt()
	for _, p := range participants {
		signingShareCount = signingShareCount.AddRaw(p.ShareCount)
	}
	activeShareCount := sdk.ZeroInt()
	for _, v := range active {
		activeShareCount = activeShareCount.AddRaw(v.ShareCount)
	}

	assert.NoError(t, err)
	assert.False(t, signingShareCount.GTE(sdk.NewInt(snap.CorruptionThreshold)))
	assert.Equal(t, int64(100), signingShareCount.Int64())
	assert.Equal(t, int64(100), activeShareCount.Int64())
	assert.Equal(t, 1, len(participants))
	assert.Equal(t, 1, len(snap.Validators)-len(participants))
	assert.Equal(t, val2, participants[0].GetSDKValidator().GetOperator())
}

func TestKeeper_StartSign_IdAlreadyInUse_ReturnError(t *testing.T) {
	s := setup()
	sigID := "sigID"
	keyID := exported.KeyID("keyID1")
	msgToSign := []byte("message")

	for _, val := range snap.Validators {
		s.Keeper.SetAvailableOperator(s.Ctx, val.GetSDKValidator().GetOperator(), keyID)
	}

	// start keygen to record the snapshot for each key
	keyInfo := types.KeyInfo{
		KeyID:   keyID,
		KeyRole: exported.MasterKey,
		KeyType: exported.Threshold,
	}
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
	s.Keeper.SetKey(s.Ctx, keyID, generatePubKey())

	assert.NoError(t, err)
	err = s.Keeper.StartSign(s.Ctx, exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msgToSign,
		SnapshotCounter: snap.Counter,
	}, s.Snapshotter, s.Voter)
	assert.NoError(t, err)

	keyID = "keyID2"
	msgToSign = []byte("second message")

	for _, val := range snap.Validators {
		s.Keeper.SetAvailableOperator(s.Ctx, val.GetSDKValidator().GetOperator(), keyID)
	}

	keyInfo = types.KeyInfo{
		KeyID:   keyID,
		KeyRole: exported.MasterKey,
		KeyType: exported.Threshold,
	}
	err = s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
	s.Keeper.SetKey(s.Ctx, exported.KeyID(keyID), generatePubKey())

	assert.NoError(t, err)
	err = s.Keeper.StartSign(s.Ctx, exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msgToSign,
		SnapshotCounter: snap.Counter,
	}, s.Snapshotter, s.Voter)
	assert.EqualError(t, err, "sig ID 'sigID' has been used before")
}

func generatePubKey() ecdsa.PublicKey {
	sk, err := ecdsa.GenerateKey(btcec.S256(), rand3.Reader)
	if err != nil {
		panic(err)
	}
	return sk.PublicKey
}
