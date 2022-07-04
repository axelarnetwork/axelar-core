package keeper

import (
	"crypto/ecdsa"
	rand3 "crypto/rand"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
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
		Height:          rand.I64Between(1, 1000000),
		TotalShareCount: sdk.NewInt(600),
		Counter:         rand.I64Between(0, 100000),
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
		KeyType: exported.Multisig,
	}
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
	assert.NoError(t, err)
	s.Keeper.SetKey(s.Ctx, generateECDSAKey(keyID))

	sigInfo := exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msg,
		SnapshotCounter: snap.Counter,
	}
	err = s.Keeper.StartSign(s.Ctx, sigInfo, s.Snapshotter, s.Voter)
	assert.NoError(t, err)

	participants, active, err := s.Keeper.selectSignParticipants(s.Ctx, s.Snapshotter, sigInfo, snap, keyInfo.KeyType)

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
	assert.Equal(t, int64(600), signingShareCount.Int64())
	assert.Equal(t, 5, len(participants))
	assert.Equal(t, 0, len(snap.Validators)-len(participants))
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
		Height:          rand.I64Between(1, 1000000),
		TotalShareCount: sdk.NewInt(200),
		Counter:         rand.I64Between(0, 100000),
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
		KeyType: exported.Multisig,
	}
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
	assert.NoError(t, err)
	s.Keeper.SetKey(s.Ctx, generateECDSAKey(keyID))

	sigInfo := exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msg,
		SnapshotCounter: snap.Counter,
	}
	err = s.Keeper.StartSign(s.Ctx, sigInfo, s.Snapshotter, s.Voter)
	assert.Error(t, err)

	participants, active, err := s.Keeper.selectSignParticipants(s.Ctx, s.Snapshotter, sigInfo, snap, keyInfo.KeyType)

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
		KeyType: exported.Multisig,
	}
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
	s.Keeper.SetKey(s.Ctx, generateECDSAKey(keyID))

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
		KeyType: exported.Multisig,
	}
	err = s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
	s.Keeper.SetKey(s.Ctx, generateECDSAKey(keyID))

	assert.NoError(t, err)
	err = s.Keeper.StartSign(s.Ctx, exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msgToSign,
		SnapshotCounter: snap.Counter,
	}, s.Snapshotter, s.Voter)
	assert.EqualError(t, err, "sig ID 'sigID' has been used before")
}

func TestMultisigSign(t *testing.T) {
	repeats := 1
	t.Run("should set sign timeout when start multisig sign", testutils.Func(func(t *testing.T) {
		s := setup()
		sigID := "sigID"
		keyID := exported.KeyID(rand.StrBetween(5, 20))
		msgToSign := []byte("message")

		signInfo := exported.SignInfo{
			KeyID:           keyID,
			SigID:           sigID,
			Msg:             msgToSign,
			SnapshotCounter: snap.Counter,
		}

		for _, val := range snap.Validators {
			s.Keeper.SetAvailableOperator(s.Ctx, val.GetSDKValidator().GetOperator(), keyID)
		}

		// start keygen to record the snapshot for each key
		keyInfo := types.KeyInfo{
			KeyID:   keyID,
			KeyRole: exported.MasterKey,
			KeyType: exported.Multisig,
		}
		err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
		assert.NoError(t, err)

		for _, v := range snap.Validators {
			// random pub keys
			var pubKeys [][]byte
			for i := int64(0); i < v.ShareCount; i++ {
				pk := btcec.PublicKey(generatePubKey())
				pubKeys = append(pubKeys, pk.SerializeCompressed())
			}
			s.Keeper.SubmitPubKeys(s.Ctx, keyID, v.GetSDKValidator().GetOperator(), pubKeys...)
		}
		s.Keeper.SetKey(s.Ctx, generateMultisigKey(keyID))
		err = s.Keeper.StartSign(s.Ctx, signInfo, s.Snapshotter, s.Voter)
		assert.NoError(t, err)
		multisigSign, ok := s.Keeper.GetMultisigSignInfo(s.Ctx, sigID)
		assert.True(t, ok)
		assert.Equal(t, int64(0), multisigSign.Count())
		assert.Len(t, multisigSign.GetTargetSigKeyPairs(), 0)
		assert.False(t, multisigSign.IsCompleted())
		keyRequirement, _ := s.Keeper.GetKeyRequirement(s.Ctx, exported.MasterKey, exported.Multisig)
		expectedTimeoutBlock := s.Ctx.BlockHeight() + keyRequirement.SignTimeout
		assert.Equal(t, expectedTimeoutBlock, multisigSign.GetTimeoutBlock())

	}).Repeat(repeats))

	t.Run("should update sig count and save signatures when validator submits signatures", testutils.Func(func(t *testing.T) {
		s := setup()
		sigID := "sigID"
		keyID := exported.KeyID(rand.StrBetween(5, 20))
		msgToSign := []byte("message")
		snap = randSnapshot()
		signInfo := exported.SignInfo{
			KeyID:           keyID,
			SigID:           sigID,
			Msg:             msgToSign,
			SnapshotCounter: snap.Counter,
		}

		for _, val := range snap.Validators {
			s.Keeper.SetAvailableOperator(s.Ctx, val.GetSDKValidator().GetOperator(), keyID)
		}

		keyInfo := types.KeyInfo{
			KeyID:   keyID,
			KeyRole: exported.MasterKey,
			KeyType: exported.Multisig,
		}

		err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyInfo, snap)
		assert.NoError(t, err)

		for _, v := range snap.Validators {
			// random pub keys
			var pubKeys [][]byte
			for i := int64(0); i < v.ShareCount; i++ {
				pk := btcec.PublicKey(generatePubKey())
				pubKeys = append(pubKeys, pk.SerializeCompressed())
			}
			s.Keeper.SubmitPubKeys(s.Ctx, keyID, v.GetSDKValidator().GetOperator(), pubKeys...)
		}
		s.Keeper.SetKey(s.Ctx, generateMultisigKey(keyID))

		err = s.Keeper.StartSign(s.Ctx, signInfo, s.Snapshotter, s.Voter)
		assert.NoError(t, err)

		sigCount := int64(0)
		var expectedPairs []exported.SigKeyPair
		for _, v := range snap.Validators {
			// random sigs
			var pairs [][]byte
			for i := int64(0); i < v.ShareCount; i++ {
				privKey, _ := btcec.NewPrivateKey(btcec.S256())
				pk := privKey.PubKey()
				d := sha256.Sum256(msgToSign)
				sig, _ := privKey.Sign(d[:])
				pair := exported.SigKeyPair{PubKey: pk.SerializeCompressed(), Signature: sig.Serialize()}
				bz, _ := pair.Marshal()
				pairs = append(pairs, bz)
				expectedPairs = append(expectedPairs, exported.SigKeyPair{PubKey: pk.SerializeCompressed(), Signature: sig.Serialize()})
			}
			ok := s.Keeper.SubmitSignatures(s.Ctx, sigID, v.GetSDKValidator().GetOperator(), pairs...)
			sigCount += v.ShareCount
			assert.True(t, ok)

			multisigSign, ok := s.Keeper.GetMultisigSignInfo(s.Ctx, sigID)
			assert.True(t, ok)
			assert.Equal(t, sigCount, multisigSign.Count())
			assert.True(t, multisigSign.DoesParticipate(v.GetSDKValidator().GetOperator()))
			assert.Equal(t, expectedPairs[:snap.CorruptionThreshold+1], multisigSign.GetTargetSigKeyPairs())
		}

	}).Repeat(repeats))
}

func generatePubKey() ecdsa.PublicKey {
	sk, err := ecdsa.GenerateKey(btcec.S256(), rand3.Reader)
	if err != nil {
		panic(err)
	}
	return sk.PublicKey
}

func randSignInfo(snap snapshot.Snapshot) exported.SignInfo {
	sigID := rand.StrBetween(5, 20)
	keyID := exported.KeyID(rand.StrBetween(5, 20))
	msgToSign := []byte("message")
	snap = randSnapshot()
	return exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msgToSign,
		SnapshotCounter: snap.Counter,
	}
}

func generateECDSAKey(keyID exported.KeyID) exported.Key {
	sk, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		panic(err)
	}

	return exported.Key{
		ID:        keyID,
		PublicKey: &exported.Key_ECDSAKey_{ECDSAKey: &exported.Key_ECDSAKey{Value: sk.PubKey().SerializeCompressed()}},
	}
}

func generateMultisigKey(keyID exported.KeyID) exported.Key {
	keyNum := rand.I64Between(5, 10)
	var pks [][]byte
	for i := int64(0); i <= keyNum; i++ {
		sk, err := btcec.NewPrivateKey(btcec.S256())
		if err != nil {
			panic(err)
		}
		pks = append(pks, sk.PubKey().SerializeCompressed())
	}

	return exported.Key{
		ID:        keyID,
		PublicKey: &exported.Key_MultisigKey_{MultisigKey: &exported.Key_MultisigKey{Values: pks, Threshold: keyNum / 2}},
	}
}
