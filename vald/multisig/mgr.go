package multisig

import (
	"context"
	"fmt"
	"time"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/sdk-utils/broadcast"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
)

// Mgr represents an object that manages all communication with the multisig process
type Mgr struct {
	client      Client
	ctx         sdkclient.Context
	participant sdk.ValAddress
	broadcaster broadcast.Broadcaster
	timeout     time.Duration
}

// NewMgr is the constructor of mgr
func NewMgr(client Client, ctx sdkclient.Context, participant sdk.ValAddress, broadcaster broadcast.Broadcaster, timeout time.Duration) *Mgr {
	return &Mgr{
		client:      client,
		ctx:         ctx,
		participant: participant,
		broadcaster: broadcaster,
		timeout:     timeout,
	}
}

func (mgr Mgr) isParticipant(p sdk.ValAddress) bool {
	return mgr.participant.Equals(p)
}

func (mgr Mgr) generateKey(keyUID string) (exported.PublicKey, error) {
	grpcCtx, cancel := context.WithTimeout(context.Background(), mgr.timeout)
	defer cancel()

	res, err := mgr.client.Keygen(grpcCtx, &tofnd.KeygenRequest{
		KeyUid:   keyUID,
		PartyUid: mgr.participant.String(),
	})
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed generating key")
	}

	switch res.GetKeygenResponse().(type) {
	case *tofnd.KeygenResponse_PubKey:
		return res.GetPubKey(), nil
	case *tofnd.KeygenResponse_Error:
		return nil, fmt.Errorf(res.GetError())
	default:
		panic(fmt.Errorf("unknown multisig keygen response %T", res.GetKeygenResponse()))
	}
}

func (mgr Mgr) sign(keyUID string, payloadHash exported.Hash, pubKey []byte) (types.Signature, error) {
	grpcCtx, cancel := context.WithTimeout(context.Background(), mgr.timeout)
	defer cancel()

	res, err := mgr.client.Sign(grpcCtx, &tofnd.SignRequest{
		KeyUid:    keyUID,
		MsgToSign: payloadHash,
		PartyUid:  mgr.participant.String(),
		PubKey:    pubKey,
	})
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "failed signing")
	}

	switch res.GetSignResponse().(type) {
	case *tofnd.SignResponse_Signature:
		return res.GetSignature(), nil
	case *tofnd.SignResponse_Error:
		return nil, fmt.Errorf(res.GetError())
	default:
		panic(fmt.Errorf("unknown multisig sign response %T", res.GetSignResponse()))
	}
}
