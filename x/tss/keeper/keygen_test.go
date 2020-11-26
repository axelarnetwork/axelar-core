package keeper

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/gob"
	"math/big"
	"testing"
	"time"

	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/voting/exported"
)

func TestKeeper_StartKeygen_IdAlreadyInUse_ReturnError(t *testing.T) {
	s := setup(t)
	assert.NoError(t, s.keeper.StartKeygen(s.ctx, "keyID", 1, s.staker.GetAllValidators(s.ctx)))
	assert.Error(t, s.keeper.StartKeygen(s.ctx, "keyID", 1, s.staker.GetAllValidators(s.ctx)))
}

// Even if no session exists the keeper must not return an error, because we need to keep validators and
// non-participating nodes consistent (for non-participating nodes there should be no session)
func TestKeeper_KeygenMsg_NoSessionWithGivenID_Return(t *testing.T) {
	s := setup(t)

	assert.NoError(t, s.keeper.KeygenMsg(s.ctx, types.MsgKeygenTraffic{
		Sender:    s.broadcaster.Proxy,
		SessionID: "keyID",
		Payload:   &tssd.TrafficOut{},
	}))
}

func TestKeeper_StartKeyRefresh_NoKeyIDConflictPossibleWith_StartKeygen(t *testing.T) {
	s := setup(t)
	assert.NoError(t, s.keeper.StartKeygen(s.ctx, "master_bitcoin", 3, s.staker.GetAllValidators(s.ctx)))
	assert.NoError(t, s.keeper.StartKeygen(s.ctx, "bitcoin", 3, s.staker.GetAllValidators(s.ctx)))

	assert.NoError(t, s.keeper.StartKeyRefresh(s.ctx, "bitcoin", s.staker.GetAllValidators(s.ctx)))
	assert.NoError(t, s.keeper.StartKeyRefresh(s.ctx, "ethereum", s.staker.GetAllValidators(s.ctx)))
}

func TestKeeper_StartKeyRefresh_CallInitPoll(t *testing.T) {
	s := setup(t)
	assert.NoError(t, s.keeper.StartKeyRefresh(s.ctx, "bitcoin", s.staker.GetAllValidators(s.ctx)))

	poll := <-s.voter.initializedPoll
	expectedPoll := exported.PollMeta{
		Module: types.ModuleName,
		Type:   types.MsgMasterKeyRefresh{}.Type(),
		ID:     "master_bitcoin",
	}
	assert.Equal(t, expectedPoll, poll)
}

func TestKeeper_StartKeyRefresh_ProtocolCompletes_VoteOnResult(t *testing.T) {
	s := setup(t)
	assert.NoError(t, s.keeper.StartKeyRefresh(s.ctx, "bitcoin", s.staker.GetAllValidators(s.ctx)))
	pubkey := ecdsa.PublicKey{X: big.NewInt(5), Y: big.NewInt(3)}
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	assert.NoError(t, enc.Encode(pubkey))
	s.client.keygen.recv <- &tssd.MessageOut{Data: &tssd.MessageOut_KeygenResult{KeygenResult: buffer.Bytes()}}

	timeout, _ := context.WithTimeout(context.Background(), 100*time.Millisecond)
	select {
	case <-timeout.Done():
		assert.FailNow(t, "no vote received")
	case vote := <-s.voter.receivedVote:
		assert.Equal(t, vote.Data(), buffer.Bytes())
	}
}
