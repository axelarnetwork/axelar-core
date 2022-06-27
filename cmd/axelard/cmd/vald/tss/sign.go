package tss

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/parse"
	"github.com/axelarnetwork/axelar-core/utils"
	tssexported "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
	tmEvents "github.com/axelarnetwork/tm-events/events"
)

// ProcessSignStart starts the communication with the sign protocol
func (mgr *Mgr) ProcessSignStart(e tmEvents.Event) error {
	keyID, keyType, sigID, participants, participantShareCounts, payload, timeout, err := parseSignStartParams(mgr.cdc, e.Attributes)
	if err != nil {
		return err
	}

	myIndex := utils.IndexOf(participants, mgr.principalAddr)
	if myIndex == -1 {
		// do not participate
		return nil
	}

	switch keyType {
	case tssexported.Threshold.SimpleString():
		return mgr.thresholdSignStart(e, keyID, timeout, sigID, payload, participants)
	case tssexported.Multisig.SimpleString():
		return mgr.multiSigSignStart(keyID, sigID, participantShareCounts[myIndex], payload)
	default:
		return fmt.Errorf(fmt.Sprintf("unknown keytype %s", keyType))
	}

}

// ProcessSignMsg forwards blockchain messages to the sign protocol
func (mgr *Mgr) ProcessSignMsg(e tmEvents.Event) error {
	sigID, from, payload := parseMsgParams(mgr.cdc, e.Attributes)
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

func parseSignStartParams(cdc *codec.LegacyAmino, attributes map[string]string) (keyID string, keyType, sigID string, participants []string, participantShareCounts []uint32, payload []byte, timeout int64, err error) {
	parsers := []*parse.AttributeParser{
		{Key: tss.AttributeKeyKeyID, Map: parse.IdentityMap},
		{Key: tss.AttributeKeyKeyType, Map: parse.IdentityMap},
		{Key: tss.AttributeKeySigID, Map: parse.IdentityMap},
		{Key: tss.AttributeKeyParticipants, Map: func(s string) (interface{}, error) {
			cdc.MustUnmarshalJSON([]byte(s), &participants)
			return participants, nil
		}},
		{Key: tss.AttributeKeyParticipantShareCounts, Map: func(s string) (interface{}, error) {
			cdc.MustUnmarshalJSON([]byte(s), &participantShareCounts)
			return participantShareCounts, nil
		}},
		{Key: tss.AttributeKeyPayload, Map: func(s string) (interface{}, error) {
			return common.Hex2Bytes(s), nil
		}},
		{Key: tss.AttributeKeyTimeout, Map: func(s string) (interface{}, error) {
			return strconv.ParseInt(s, 10, 64)
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", "", "", nil, nil, nil, 0, err
	}

	return results[0].(string), results[1].(string), results[2].(string), results[3].([]string), results[4].([]uint32), results[5].([]byte), results[6].(int64), nil
}

func (mgr *Mgr) thresholdSignStart(e tmEvents.Event, keyID string, timeout int64, sigID string, payload []byte, participants []string) error {
	done := false
	session := mgr.timeoutQueue.Enqueue(sigID, e.Height+timeout)

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
		tssMsg := types.NewProcessSignTrafficRequest(mgr.cliCtx.FromAddress, sigID, *msg)
		if _, err := mgr.broadcaster.Broadcast(context.TODO(), tssMsg); err != nil {
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
	vote := &tss.VoteSigRequest{Sender: mgr.cliCtx.FromAddress, PollKey: key, Result: *result}
	_, err := mgr.broadcaster.Broadcast(context.TODO(), vote)
	return err
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

func (mgr *Mgr) multiSigSignStart(keyID string, sigID string, shares uint32, payload []byte) error {
	var signatures [][]byte
	pubKeys, found := mgr.getKey(keyID)
	if !found {
		return fmt.Errorf("received multisig sign request for sigID %s for an unknown key ID %s", sigID, keyID)
	}

	for i := uint32(0); i < shares; i++ {
		keyUID := fmt.Sprintf("%s_%d", keyID, i)
		signature, err := mgr.multiSigSign(keyUID, payload, pubKeys[i])
		if err != nil {
			return err
		}
		signatures = append(signatures, signature)
	}

	mgr.Logger.Info(fmt.Sprintf("operator %s sending multisig signatures for sig %s", mgr.principalAddr, sigID))
	tssMsg := tss.NewSubmitMultisigSignaturesRequest(mgr.cliCtx.FromAddress, sigID, signatures)

	if _, err := mgr.broadcaster.Broadcast(context.TODO(), tssMsg); err != nil {
		return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing multisig keys msg")
	}

	return nil
}

// multiSigSign send sign request to Tofnd Multisig service
func (mgr *Mgr) multiSigSign(keyUID string, msgToSign []byte, pubKey []byte) ([]byte, error) {
	grpcCtx, cancel := context.WithTimeout(context.Background(), mgr.Timeout)
	defer cancel()

	signRequest := &tofnd.SignRequest{
		KeyUid:    keyUID,
		MsgToSign: msgToSign,
		PartyUid:  mgr.principalAddr,
		PubKey:    pubKey,
	}
	res, err := mgr.multiSigClient.Sign(grpcCtx, signRequest)
	if err != nil {
		return nil, err
	}
	switch res.GetSignResponse().(type) {
	case *tofnd.SignResponse_Signature:
		return res.GetSignature(), nil
	case *tofnd.SignResponse_Error:
		return nil, sdkerrors.Wrap(err, res.GetError())
	default:
		return nil, sdkerrors.Wrap(err, "unknown multisig sign response")
	}
}
