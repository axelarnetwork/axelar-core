package keeper

import (
	"testing"

	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

func TestKeeper_StartSign_IdAlreadyInUse_ReturnError(t *testing.T) {
	s := setup(t)
	sigID := "sigID"
	keyID := "keyID1"
	msgToSign := []byte("message")

	// start keygen to record the snapshot for each key
	_, err := s.Keeper.StartKeygen(s.Ctx, keyID, 1, snap)
	assert.NoError(t, err)
	err = s.Keeper.StartSign(s.Ctx, keyID, sigID, msgToSign)
	assert.NoError(t, err)

	keyID = "keyID2"
	msgToSign = []byte("second message")
	_, err = s.Keeper.StartKeygen(s.Ctx, keyID, 1, snap)
	assert.NoError(t, err)
	err = s.Keeper.StartSign(s.Ctx, keyID, sigID, msgToSign)
	assert.Error(t, err)
}

// Even if no session exists the keeper must not return an error, because we need to keep validators and
// non-participating nodes consistent (for non-participating nodes there should be no session)
func TestKeeper_SignMsg_NoSessionWithGivenID_Return(t *testing.T) {
	s := setup(t)

	assert.NoError(t, s.Keeper.SignMsg(s.Ctx, types.MsgSignTraffic{
		Sender:    s.Broadcaster.GetProxy(s.Ctx, s.Broadcaster.LocalPrincipal),
		SessionID: "sigID",
		Payload:   &tssd.TrafficOut{},
	}))
}
