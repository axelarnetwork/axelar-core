package keeper

import (
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

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

	// start keygen to record the snapshot for each key
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, snap)
	assert.NoError(t, err)

	height, err := s.Keeper.ScheduleSign(s.Ctx, exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msg,
		SnapshotCounter: snap.Counter,
	})

	for _, val := range snap.Validators {
		err = s.Keeper.SetAvailableOperator(s.Ctx, sigID, exported.AckType_Sign, val.GetSDKValidator().GetOperator())
		assert.NoError(t, err)
	}

	s.Ctx = s.Ctx.WithBlockHeight(height)
	s.Keeper.SetSignParticipants(s.Ctx, sigID, snap.Validators)

	threshold, ok := s.Keeper.GetCorruptionThreshold(s.Ctx, keyID)
	assert.True(t, ok)

	ok = s.Keeper.MeetsThreshold(s.Ctx, sigID, threshold)
	assert.False(t, ok)
	assert.Equal(t, int64(100), s.Keeper.GetTotalShareCount(s.Ctx, sigID))
}

func TestKeeper_StartSign_IdAlreadyInUse_ReturnError(t *testing.T) {
	s := setup()
	sigID := "sigID"
	keyID := "keyID1"
	msgToSign := []byte("message")

	// start keygen to record the snapshot for each key
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, snap)
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
	err = s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, snap)
	assert.NoError(t, err)
	_, err = s.Keeper.ScheduleSign(s.Ctx, exported.SignInfo{
		KeyID:           keyID,
		SigID:           sigID,
		Msg:             msgToSign,
		SnapshotCounter: snap.Counter,
	})
	assert.EqualError(t, err, "sigID 'sigID' has been used before")
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
