package tss

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// ProcessSignStart starts the communication with the sign protocol
func (mgr *Mgr) ProcessSignStart(attributes []sdk.Attribute) error {
	keyID, sigID, participants, payload := parseSignStartParams(attributes)
	_, ok := indexOf(participants, mgr.principalAddr)
	if !ok {
		// do not participate
		return nil
	}

	stream, cancel, err := mgr.startSign(keyID, sigID, participants, payload)
	if err != nil {
		return err
	}
	mgr.setSignStream(sigID, stream)

	// use error channel to coordinate errors during communication with keygen protocol
	errChan := make(chan error, 3)
	intermediateMsgs, result, streamErrChan := handleStream(stream, cancel, mgr.Logger)
	go func() {
		err, ok := <-streamErrChan
		if ok {
			errChan <- err
		}
	}()
	go func() {
		err := mgr.handleIntermediateSignMsgs(sigID, intermediateMsgs)
		if err != nil {
			errChan <- err
		}
	}()
	go func() {
		err := mgr.handleSignResult(sigID, result)
		if err != nil {
			errChan <- err
		} else {
			// this is the last part of the keygen, so if there are no errors here return nil
			errChan <- nil
		}
	}()
	return <-errChan
}

// ProcessSignMsg forwards blockchain messages to the sign protocol
func (mgr *Mgr) ProcessSignMsg(attributes []sdk.Attribute) error {
	sigID, from, payload := parseMsgParams(attributes)
	msgIn, err := prepareTrafficIn(mgr.principalAddr, from, sigID, payload, mgr.Logger)
	if err != nil {
		return err
	}
	// this message is not meant for this tofnd instance
	if msgIn == nil {
		return nil
	}

	stream, ok := mgr.getSignStream(sigID)
	if !ok {
		mgr.Logger.Info(fmt.Sprintf("no sign session with id %s. This process does not participate", sigID))
		return nil
	}

	if err := stream.Send(msgIn); err != nil {
		return sdkerrors.Wrap(err, "failure to send incoming msg to gRPC server")
	}
	return nil
}

func parseSignStartParams(attributes []sdk.Attribute) (keyID string, sigID string, participants []string, payload []byte) {
	for _, attribute := range attributes {
		switch attribute.Key {
		case tss.AttributeKeyKeyID:
			keyID = attribute.Value
		case tss.AttributeKeySigID:
			sigID = attribute.Value
		case tss.AttributeKeyParticipants:
			codec.Cdc.MustUnmarshalJSON([]byte(attribute.Value), &participants)
		case tss.AttributeKeyPayload:
			payload = []byte(attribute.Value)
		default:
		}
	}

	return keyID, sigID, participants, payload
}

func (mgr *Mgr) startSign(keyID string, sigID string, participants []string, payload []byte) (tss.Stream, context.CancelFunc, error) {
	if _, ok := mgr.getSignStream(sigID); ok {
		return nil, nil, fmt.Errorf("sign protocol for ID %s already in progress", sigID)
	}

	grpcCtx, cancel := context.WithTimeout(context.Background(), mgr.Timeout)
	stream, err := mgr.client.Sign(grpcCtx)
	if err != nil {
		cancel()
		return nil, nil, sdkerrors.Wrap(err, "failed tofnd gRPC call Sign")
	}

	signInit := &tofnd.MessageIn_SignInit{
		SignInit: &tofnd.SignInit{
			NewSigUid:     sigID,
			KeyUid:        keyID,
			PartyUids:     participants,
			MessageToSign: payload,
		},
	}

	if err := stream.Send(&tofnd.MessageIn{Data: signInit}); err != nil {
		cancel()
		return nil, nil, err
	}

	return stream, cancel, nil
}

func (mgr *Mgr) handleIntermediateSignMsgs(sigID string, intermediate <-chan *tofnd.TrafficOut) error {
	for msg := range intermediate {
		mgr.Logger.Debug(fmt.Sprintf("outgoing sign msg: sig [%.20s] from me [%.20s] to [%.20s] broadcast [%t]\n",
			sigID, mgr.principalAddr, msg.ToPartyUid, msg.IsBroadcast))
		// sender is set by broadcaster
		tssMsg := &tss.MsgSignTraffic{Sender: mgr.sender, SessionID: sigID, Payload: msg}
		if err := mgr.broadcaster.Broadcast(tssMsg); err != nil {
			return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing sign msg")
		}
	}
	return nil
}

func (mgr *Mgr) handleSignResult(sigID string, result <-chan []byte) error {
	// Delete the reference to the signing stream with sigID because entering this function means the tss protocol has completed
	defer func() {
		mgr.sign.Lock()
		defer mgr.sign.Unlock()
		delete(mgr.signStreams, sigID)
	}()

	bz := <-result
	mgr.Logger.Info(fmt.Sprintf("handler goroutine: received sig from server! [%.20s]", bz))

	poll := voting.NewPollMeta(tss.ModuleName, tss.EventTypeSign, sigID)
	vote := tss.MsgVoteSig{Sender: mgr.sender, PollMeta: poll, SigBytes: bz}
	return mgr.broadcaster.Broadcast(vote)
}

func (mgr *Mgr) getSignStream(sigID string) (tss.Stream, bool) {
	mgr.sign.RLock()
	defer mgr.sign.RUnlock()

	stream, ok := mgr.signStreams[sigID]
	return stream, ok
}

func (mgr *Mgr) setSignStream(sigID string, stream tss.Stream) {
	mgr.sign.Lock()
	defer mgr.sign.Unlock()

	mgr.signStreams[sigID] = stream
}
