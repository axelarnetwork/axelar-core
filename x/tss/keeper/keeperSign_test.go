package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	rand2 "github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func TestStartSign_NoEnoughActiveValidators(t *testing.T) {
	s := setup(t)
	sigID := "sigID"
	keyID := "keyID"
	msg := []byte("message")

	snap := snapshot.Snapshot{
		Validators: []snapshot.Validator{
			&snapMock.ValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return sdk.ValAddress("validator1") },
				GetConsensusPowerFunc: func() int64 { return 100 },
				GetConsAddrFunc:       func() sdk.ConsAddress { return sdk.ValAddress("validator1").Bytes() },
				IsJailedFunc:          func() bool { return true },
			},
			&snapMock.ValidatorMock{
				GetOperatorFunc:       func() sdk.ValAddress { return sdk.ValAddress("validator2") },
				GetConsensusPowerFunc: func() int64 { return 100 },
				GetConsAddrFunc:       func() sdk.ConsAddress { return sdk.ValAddress("validator2").Bytes() },
				IsJailedFunc:          func() bool { return false },
			},
		},
		Timestamp:            time.Now(),
		Height:               rand2.I64Between(1, 1000000),
		TotalPower:           sdk.NewInt(200),
		ValidatorsTotalPower: sdk.NewInt(200),
		Counter:              rand2.I64Between(0, 100000),
	}

	// start keygen to record the snapshot for each key
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, 1, snap)
	assert.NoError(t, err)
	err = s.Keeper.StartSign(s.Ctx, s.Voter, keyID, sigID, msg, snap)
	assert.EqualError(t, err, "not enough active validators are online: threshold [1], online [1]")
}

func TestKeeper_StartSign_IdAlreadyInUse_ReturnError(t *testing.T) {
	s := setup(t)
	sigID := "sigID"
	keyID := "keyID1"
	msgToSign := []byte("message")

	// start keygen to record the snapshot for each key
	err := s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, 1, snap)
	assert.NoError(t, err)
	err = s.Keeper.StartSign(s.Ctx, s.Voter, keyID, sigID, msgToSign, exported.Snapshot{})
	assert.NoError(t, err)

	keyID = "keyID2"
	msgToSign = []byte("second message")
	err = s.Keeper.StartKeygen(s.Ctx, s.Voter, keyID, 1, snap)
	assert.NoError(t, err)
	err = s.Keeper.StartSign(s.Ctx, s.Voter, keyID, sigID, msgToSign, exported.Snapshot{})
	assert.Error(t, err)
}

func TestKeeper_SignMsg_NoSessionWithGivenID_Error(t *testing.T) {
	s := setup(t)

	assert.Error(t, s.Keeper.SignMsg(s.Ctx, types.MsgSignTraffic{
		Sender:    s.Broadcaster.GetProxy(s.Ctx, s.Broadcaster.LocalPrincipal),
		SessionID: "sigID",
		Payload:   &tofnd.TrafficOut{},
	}))
}
