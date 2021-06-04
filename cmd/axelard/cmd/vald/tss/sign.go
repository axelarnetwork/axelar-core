package tss

import (
	"context"
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// ProcessSignStart starts the communication with the sign protocol
func (mgr *Mgr) ProcessSignStart(blockHeight int64, attributes []sdk.Attribute) error {
	keyID, sigID, participants, payload := parseSignStartParams(mgr.cdc, attributes)
	_, ok := indexOf(participants, mgr.principalAddr)
	if !ok {
		// do not participate
		return nil
	}

	session := mgr.timeoutQueue.Enqueue(sigID, blockHeight+mgr.sessionTimeout)

	stream, cancel, err := mgr.startSign(keyID, sigID, participants, payload)
	if err != nil {
		return err
	}
	mgr.setSignStream(sigID, stream)

	// use error channel to coordinate errors during communication with keygen protocol
	errChan := make(chan error, 4)
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
		session.WaitForTimeout()

		found, err := mgr.abortSign(sigID)
		if !found {
			return
		}

		mgr.Logger.Info(fmt.Sprintf("aborted sign protocol %s due to timeout", sigID))
		errChan <- err
	}()
	go func() {
		errChan <- mgr.handleSignResult(sigID, result)
	}()

	return <-errChan
}

// ProcessSignMsg forwards blockchain messages to the sign protocol
func (mgr *Mgr) ProcessSignMsg(_ int64, attributes []sdk.Attribute) error {
	sigID, from, payload := parseMsgParams(mgr.cdc, attributes)
	msgIn := prepareTrafficIn(mgr.principalAddr, from, sigID, payload, mgr.Logger)
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

func parseSignStartParams(cdc *codec.LegacyAmino, attributes []sdk.Attribute) (keyID string, sigID string, participants []string, payload []byte) {
	for _, attribute := range attributes {
		switch attribute.Key {
		case tss.AttributeKeyKeyID:
			keyID = attribute.Value
		case tss.AttributeKeySigID:
			sigID = attribute.Value
		case tss.AttributeKeyParticipants:
			cdc.MustUnmarshalJSON([]byte(attribute.Value), &participants)
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
		tssMsg := &tss.ProcessSignTrafficRequest{Sender: mgr.sender, SessionID: sigID, Payload: msg}
		if err := mgr.broadcaster.Broadcast(tssMsg); err != nil {
			return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing sign msg")
		}
	}
	return nil
}

func (mgr *Mgr) handleSignResult(sigID string, resultChan <-chan interface{}) error {
	// Delete the reference to the signing stream with sigID because entering this function means the tss protocol has completed
	defer func() {
		mgr.sign.Lock()
		defer mgr.sign.Unlock()
		delete(mgr.signStreams, sigID)
	}()

	result := (<-resultChan).(*tofnd.MessageOut_SignResult)
	if result.GetCriminals() != nil {
		// criminals have to be sorted in ascending order
		sort.Stable(result.GetCriminals())
	}

	poll := voting.NewPollMeta(tss.ModuleName, sigID)
	vote := &tss.VoteSigRequest{Sender: mgr.sender, PollMeta: poll, Result: result}
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
