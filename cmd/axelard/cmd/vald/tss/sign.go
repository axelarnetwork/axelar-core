package tss

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// ProcessSignAck broadcasts an acknowledgment for a signature
func (mgr *Mgr) ProcessSignAck(blockHeight int64, attributes []sdk.Attribute) error {
	keyID, sigID, height, err := parseSignAckParams(mgr.cdc, attributes)
	grpcCtx, cancel := context.WithTimeout(context.Background(), mgr.Timeout)
	defer cancel()

	request := &tofnd.KeyPresenceRequest{
		KeyUid: keyID,
	}

	response, err := mgr.client.KeyPresence(grpcCtx, request)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to invoke KeyPresence grpc for key ID '%s'", keyID)
	}

	switch response.Response {
	case tofnd.RESPONSE_UNSPECIFIED:
		fallthrough
	case tofnd.RESPONSE_FAIL:
		return sdkerrors.Wrap(err, "tofnd not set up correctly")
	case tofnd.RESPONSE_ABSENT:
		return sdkerrors.Wrap(err, "key ID '%s' not present at tofnd")
	case tofnd.RESPONSE_PRESENT:
		mgr.Logger.Info(fmt.Sprintf("sending keygen ack for key ID '%s' and sig ID '%s'", keyID, sigID))
		tssMsg := tss.NewAckRequest(mgr.sender, sigID, exported.AckType_Sign, height)
		if err := mgr.broadcaster.Broadcast(tssMsg); err != nil {
			return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing ack msg")
		}
	default:
		return sdkerrors.Wrap(err, "unknown tofnd response")
	}

	return nil
}

// ProcessSignStart starts the communication with the sign protocol
func (mgr *Mgr) ProcessSignStart(blockHeight int64, attributes []sdk.Attribute) error {
	keyID, sigID, participants, payload, timeout, err := parseSignStartParams(mgr.cdc, attributes)
	if err != nil {
		return err
	}

	if utils.IndexOf(participants, mgr.principalAddr) == -1 {
		// do not participate
		return nil
	}

	done := false
	session := mgr.timeoutQueue.Enqueue(sigID, blockHeight+timeout)

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

		if done {
			return
		}

		errChan <- mgr.abortSign(sigID)
		mgr.Logger.Info(fmt.Sprintf("aborted sign protocol %s due to timeout", sigID))
	}()
	go func() {
		err := mgr.handleSignResult(sigID, result)
		done = true

		errChan <- err
	}()

	return <-errChan
}

// ProcessSignMsg forwards blockchain messages to the sign protocol
func (mgr *Mgr) ProcessSignMsg(attributes []sdk.Attribute) error {
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

func parseSignAckParams(cdc *codec.LegacyAmino, attributes []sdk.Attribute) (keyID string, sigID string, height int64, err error) {
	var keyIDFound, sigIDFound, heightFound bool
	for _, attribute := range attributes {
		switch attribute.Key {
		case tss.AttributeKeyKeyID:
			keyID = attribute.Value
			keyIDFound = true
		case tss.AttributeKeySigID:
			sigID = attribute.Value
			sigIDFound = true

		case tss.AttributeKeyHeight:
			height, err = strconv.ParseInt(attribute.Value, 10, 64)
			if err != nil {
				return "", "", -1, err
			}
			heightFound = true
		default:
		}
	}

	if !keyIDFound || !sigIDFound || !heightFound {
		return "", "", -1, fmt.Errorf("insufficient event attributes")
	}

	return keyID, sigID, height, nil
}

func parseSignStartParams(cdc *codec.LegacyAmino, attributes []sdk.Attribute) (keyID string, sigID string, participants []string, payload []byte, timeout int64, err error) {
	var keyIDFound, sigIDFound, participantsFound, payloadFound, timeoutFound bool
	for _, attribute := range attributes {
		switch attribute.Key {
		case tss.AttributeKeyKeyID:
			keyID = attribute.Value
			keyIDFound = true
		case tss.AttributeKeySigID:
			sigID = attribute.Value
			sigIDFound = true
		case tss.AttributeKeyParticipants:
			cdc.MustUnmarshalJSON([]byte(attribute.Value), &participants)
			participantsFound = true
		case tss.AttributeKeyPayload:
			payload = []byte(attribute.Value)
			payloadFound = true
		case tss.AttributeKeyTimeout:
			timeout, err = strconv.ParseInt(attribute.Value, 10, 64)
			if err != nil {
				panic(err)
			}
			timeoutFound = true
		default:
		}
	}

	if !keyIDFound || !sigIDFound || !participantsFound || !payloadFound || !timeoutFound {
		return "", "", nil, nil, 0, fmt.Errorf("insufficient event attributes")
	}

	return keyID, sigID, participants, payload, timeout, nil
}

func (mgr *Mgr) startSign(keyID string, sigID string, participants []string, payload []byte) (Stream, context.CancelFunc, error) {
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

	r, ok := <-resultChan
	if !ok {
		return fmt.Errorf("failed to receive sign result, channel was closed by the server")
	}

	result := r.(*tofnd.MessageOut_SignResult)
	if result.GetCriminals() != nil {
		// criminals have to be sorted in ascending order
		sort.Stable(result.GetCriminals())
	}

	mgr.Logger.Debug(fmt.Sprintf("handler goroutine: received sign result for %s [%+v]", sigID, result))

	key := voting.NewPollKey(tss.ModuleName, sigID)
	vote := &tss.VoteSigRequest{Sender: mgr.sender, PollKey: key, Result: result}
	return mgr.broadcaster.Broadcast(vote)
}

func (mgr *Mgr) getSignStream(sigID string) (Stream, bool) {
	mgr.sign.RLock()
	defer mgr.sign.RUnlock()

	stream, ok := mgr.signStreams[sigID]
	return stream, ok
}

func (mgr *Mgr) setSignStream(sigID string, stream Stream) {
	mgr.sign.Lock()
	defer mgr.sign.Unlock()

	mgr.signStreams[sigID] = NewLockableStream(stream)
}
