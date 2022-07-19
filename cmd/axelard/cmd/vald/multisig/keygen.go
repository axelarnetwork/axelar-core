package multisig

import (
	"context"
	"crypto/sha256"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/utils/slices"
)

// ProcessKeygenStarted handles event keygen started
func (mgr *Mgr) ProcessKeygenStarted(event *types.KeygenStarted) error {
	if !slices.Any(event.Participants, mgr.isParticipant) {
		return nil
	}

	keyUID := fmt.Sprintf("%s_%d", event.GetKeyID().String(), 0)
	partyUID := mgr.participant.String()

	pubKey, err := mgr.generateKey(keyUID, partyUID)
	if err != nil {
		return err
	}

	payloadHash := sha256.Sum256(mgr.ctx.FromAddress)
	sig, err := mgr.sign(keyUID, payloadHash[:], partyUID, pubKey)
	if err != nil {
		return err
	}

	mgr.logger.Info(fmt.Sprintf("operator %s sending public key for multisig key %s", partyUID, keyUID))

	msg := types.NewSubmitPubKeyRequest(mgr.ctx.FromAddress, event.GetKeyID(), pubKey, sig)
	if _, err := mgr.broadcaster.Broadcast(context.Background(), msg); err != nil {
		return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing submit pub key message")
	}

	return nil
}
