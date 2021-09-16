package keeper

import (
	"bytes"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
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
	keyID := "keyID"
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
				GetConsensusPowerFunc: func() int64 { return 150 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val1.Bytes(), nil },
				IsJailedFunc:          func() bool { return true },
			}, 150),
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val2 },
				GetConsensusPowerFunc: func() int64 { return 100 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val2.Bytes(), nil },
				IsJailedFunc:          func() bool { return false },
			}, 100),
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val3 },
				GetConsensusPowerFunc: func() int64 { return 80 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val3.Bytes(), nil },
				IsJailedFunc:          func() bool { return false },
			}, 80),
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val4 },
				GetConsensusPowerFunc: func() int64 { return 70 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val4.Bytes(), nil },
				IsJailedFunc:          func() bool { return false },
			}, 70),
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val5 },
				GetConsensusPowerFunc: func() int64 { return 50 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val5.Bytes(), nil },
				IsJailedFunc:          func() bool { return false },
			}, 50),
		},
		Timestamp:       time.Now(),
		Height:          rand2.I64Between(1, 1000000),
		TotalShareCount: sdk.NewInt(450),
		Counter:         rand2.I64Between(0, 100000),
	}
	snap.CorruptionThreshold = exported.ComputeAbsCorruptionThreshold(utils.Threshold{Numerator: 2, Denominator: 3}, snap.TotalShareCount)
	assert.Equal(t, int64(300), snap.CorruptionThreshold)
	s.Snapshotter.GetValidatorIllegibilityFunc = func(ctx sdk.Context, validator snapshot.SDKValidator) (snapshot.ValidatorIllegibility, error) {
		return snapshot.None, nil
	}

	// start keygen to record the snapshot for each key
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, exported.MasterKey, snap)
	assert.NoError(t, err)

	height, err := s.Keeper.ScheduleSign(s.Ctx, exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msg,
		SnapshotCounter: snap.Counter,
	})
	assert.NoError(t, err)

	for _, val := range snap.Validators {
		err = s.Keeper.SetAvailableOperator(s.Ctx, sigID, exported.AckType_Sign, val.GetSDKValidator().GetOperator())
		assert.NoError(t, err)
	}

	s.Ctx = s.Ctx.WithBlockHeight(height)
	participants, activeShareCount, err := s.Keeper.SelectSignParticipants(s.Ctx, &s.Snapshotter, sigID, snap)

	signingShareCount := sdk.ZeroInt()
	for _, p := range participants {
		signingShareCount = signingShareCount.AddRaw(p.ShareCount)
	}

	assert.NoError(t, err)
	assert.True(t, signingShareCount.GTE(sdk.NewInt(snap.CorruptionThreshold)))
	assert.Equal(t, int64(450), activeShareCount.Int64())
	assert.Equal(t, int64(330), signingShareCount.Int64())
	assert.Equal(t, 3, len(participants))
	assert.Equal(t, 2, len(snap.Validators)-len(participants))
}

func TestStartSign_NoEnoughActiveValidators(t *testing.T) {
	s := setup()
	sigID := "sigID"
	keyID := "keyID"
	msg := []byte("message")
	val1 := rand.ValAddr()
	val2 := rand.ValAddr()

	snap := snapshot.Snapshot{
		Validators: []snapshot.Validator{
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val1 },
				GetConsensusPowerFunc: func() int64 { return 100 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val1.Bytes(), nil },
				IsJailedFunc:          func() bool { return true },
			}, 100),
			snapshot.NewValidator(&snapMock.SDKValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return val2 },
				GetConsensusPowerFunc: func() int64 { return 100 },
				GetConsAddrFunc:       func() (sdk.ConsAddress, error) { return val2.Bytes(), nil },
				IsJailedFunc:          func() bool { return false },
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

	// start keygen to record the snapshot for each key
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, exported.MasterKey, snap)
	assert.NoError(t, err)

	height, err := s.Keeper.ScheduleSign(s.Ctx, exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msg,
		SnapshotCounter: snap.Counter,
	})
	assert.NoError(t, err)

	for _, val := range snap.Validators {
		err = s.Keeper.SetAvailableOperator(s.Ctx, sigID, exported.AckType_Sign, val.GetSDKValidator().GetOperator())
		assert.NoError(t, err)
	}

	s.Ctx = s.Ctx.WithBlockHeight(height)
	participants, activeShareCount, err := s.Keeper.SelectSignParticipants(s.Ctx, &s.Snapshotter, sigID, snap)

	signingShareCount := sdk.ZeroInt()
	for _, p := range participants {
		signingShareCount = signingShareCount.AddRaw(p.ShareCount)
	}

	assert.NoError(t, err)
	assert.False(t, signingShareCount.GTE(sdk.NewInt(snap.CorruptionThreshold)))
	assert.Equal(t, int64(100), signingShareCount.Int64())
	assert.Equal(t, int64(100), activeShareCount.Int64())
}

func TestKeeper_StartSign_IdAlreadyInUse_ReturnError(t *testing.T) {
	s := setup()
	sigID := "sigID"
	keyID := "keyID1"
	msgToSign := []byte("message")

	// start keygen to record the snapshot for each key
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, exported.MasterKey, snap)
	assert.NoError(t, err)
	_, err = s.Keeper.ScheduleSign(s.Ctx, exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msgToSign,
		SnapshotCounter: snap.Counter,
	})
	assert.NoError(t, err)

	keyID = "keyID2"
	msgToSign = []byte("second message")
	err = s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, exported.MasterKey, snap)
	assert.NoError(t, err)
	_, err = s.Keeper.ScheduleSign(s.Ctx, exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msgToSign,
		SnapshotCounter: snap.Counter,
	})
	assert.EqualError(t, err, "sig ID 'sigID' has been used before")
}

func TestScheduleSignAtHeight(t *testing.T) {
	t.Run("testing schedule sign", testutils.Func(func(t *testing.T) {
		s := setup()
		numSigns := int(rand2.I64Between(10, 30))
		currentHeight := s.Ctx.BlockHeight()
		expectedInfos := make([]exported.SignInfo, numSigns)
		snapshotSeq := rand2.I64Between(20, 50)

		// schedule signs
		for i := 0; i < numSigns; i++ {
			info := exported.SignInfo{
				KeyID:           rand2.StrBetween(5, 10),
				SigID:           rand2.StrBetween(10, 20),
				Msg:             []byte(rand2.StrBetween(20, 50)),
				SnapshotCounter: snapshotSeq + int64(i),
			}
			expectedInfos[i] = info
			height, err := s.Keeper.ScheduleSign(s.Ctx, info)

			assert.NoError(t, err)
			assert.Equal(t, s.Keeper.GetParams(s.Ctx).AckWindowInBlocks+currentHeight, height)
		}

		// verify signs from above
		s.Ctx = s.Ctx.WithBlockHeight(currentHeight + s.Keeper.GetParams(s.Ctx).AckWindowInBlocks)
		infos := s.Keeper.GetAllSignInfosAtCurrentHeight(s.Ctx)

		actualNumInfos := 0
		for _, expected := range expectedInfos {
			for _, actual := range infos {
				bz1, err := actual.Marshal()
				assert.NoError(t, err)
				bz2, err := expected.Marshal()
				assert.NoError(t, err)

				if bytes.Equal(bz1, bz2) {
					actualNumInfos++
					break
				}
			}
		}
		assert.Len(t, expectedInfos, actualNumInfos)
		assert.Equal(t, numSigns, actualNumInfos)

		// check that we can delete scheduled signs
		for i, info := range infos {
			s.Keeper.DeleteScheduledSign(s.Ctx, info.SigID)
			infos := s.Keeper.GetAllSignInfosAtCurrentHeight(s.Ctx)
			assert.Len(t, infos, actualNumInfos-(i+1))
		}
	}).Repeat(20))
}

func TestScheduleSignEvents(t *testing.T) {
	t.Run("testing scheduled sign events", testutils.Func(func(t *testing.T) {
		s := setup()
		currentHeight := s.Ctx.BlockHeight()
		keyID := rand2.Str(20)
		sigID := rand2.Str(20)
		height, err := s.Keeper.ScheduleSign(s.Ctx, exported.SignInfo{
			KeyID:           keyID,
			SigID:           sigID,
			Msg:             rand.Bytes(20),
			SnapshotCounter: snap.Counter,
		})
		assert.NoError(t, err)
		assert.Equal(t, s.Keeper.GetParams(s.Ctx).AckWindowInBlocks+currentHeight, height)

		assert.Len(t, s.Ctx.EventManager().ABCIEvents(), 1)
		assert.Equal(t, s.Ctx.EventManager().ABCIEvents()[0].Type, types.EventTypeAck)

		var heightFound, keyIDFound, sigIDFound bool
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
			case types.AttributeKeySigID:
				if string(attribute.Value) == sigID {
					sigIDFound = true
				}
			}
		}

		assert.True(t, heightFound)
		assert.True(t, keyIDFound)
		assert.True(t, sigIDFound)
	}).Repeat(20))
}
