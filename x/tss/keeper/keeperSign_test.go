package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func TestKeeper_StartSign_IdAlreadyInUse_ReturnError(t *testing.T) {
	s := setup(t)
	sigID := "sigID"
	keyID := "keyID1"
	msgToSign := []byte("message")

	// start keygen to record the snapshot for each key
	err := s.Keeper.StartKeygen(s.Ctx, keyID, 1, snap)
	assert.NoError(t, err)
	err = s.Keeper.StartSign(s.Ctx, keyID, sigID, msgToSign)
	assert.NoError(t, err)

	keyID = "keyID2"
	msgToSign = []byte("second message")
	err = s.Keeper.StartKeygen(s.Ctx, keyID, 1, snap)
	assert.NoError(t, err)
	err = s.Keeper.StartSign(s.Ctx, keyID, sigID, msgToSign)
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
