package multisig

import (
	"context"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/utils/log"
)

// ProcessSigningStarted handles event signing started
func (mgr *Mgr) ProcessSigningStarted(event *types.SigningStarted) error {
	pubKey, ok := event.PubKeys[mgr.participant.String()]
	if !ok {
		return nil
	}

	keyUID := fmt.Sprintf("%s_%d", event.GetKeyID().String(), 0)
	partyUID := mgr.participant.String()

	sig, err := mgr.sign(keyUID, event.GetPayloadHash(), pubKey)
	if err != nil {
		return err
	}

	log.Infof("operator %s sending signature for signing %d", partyUID, event.GetSigID())

	msg := types.NewSubmitSignatureRequest(mgr.ctx.FromAddress, event.GetSigID(), sig)
	if _, err := mgr.broadcaster.Broadcast(context.Background(), msg); err != nil {
		return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing submit signature message")
	}

	return nil
}
