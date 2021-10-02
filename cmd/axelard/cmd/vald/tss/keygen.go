package tss

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	tmEvents "github.com/axelarnetwork/tm-events/events"
	"github.com/btcsuite/btcd/btcec"
	sdkFlags "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/parse"
	"github.com/axelarnetwork/axelar-core/utils"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// ProcessKeygenAck broadcasts an acknowledgment for a keygen
func (mgr *Mgr) ProcessKeygenAck(e tmEvents.Event) error {
	keyID, height, err := parseKeygenAckParams(e.Attributes)
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
	case tofnd.RESPONSE_PRESENT:
		return sdkerrors.Wrap(err, "key ID '%s' already present at tofnd")
	case tofnd.RESPONSE_ABSENT:
		mgr.Logger.Info(fmt.Sprintf("sending keygen ack for key ID '%s'", keyID))
		tssMsg := tss.NewAckRequest(mgr.cliCtx.FromAddress, keyID, exported.AckType_Keygen, height)
		refundableMsg := axelarnet.NewRefundMsgRequest(mgr.cliCtx.FromAddress, tssMsg)
		if _, err := mgr.broadcaster.Broadcast(mgr.cliCtx.WithBroadcastMode(sdkFlags.BroadcastSync), refundableMsg); err != nil {
			return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing ack msg")
		}
	default:
		return sdkerrors.Wrap(err, "unknown tofnd response")
	}

	return nil
}

// ProcessKeygenStart starts the communication with the keygen protocol
func (mgr *Mgr) ProcessKeygenStart(e tmEvents.Event) error {
	keyID, threshold, participants, participantShareCounts, timeout, err := parseKeygenStartParams(mgr.cdc, e.Attributes)
	if err != nil {
		return err
	}

	myIndex := utils.IndexOf(participants, mgr.principalAddr)
	if myIndex == -1 {
		// do not participate
		return nil
	}

	done := false
	session := mgr.timeoutQueue.Enqueue(keyID, e.Height+timeout)

	stream, cancel, err := mgr.startKeygen(keyID, threshold, int32(myIndex), participants, participantShareCounts)
	if err != nil {
		return err
	}
	mgr.setKeygenStream(keyID, stream)

	// use error channel to coordinate errors during communication with sign protocol
	errChan := make(chan error, 4)
	intermediateMsgs, result, streamErrChan := handleStream(stream, cancel, mgr.Logger)
	go func() {
		err, ok := <-streamErrChan
		if ok {
			errChan <- err
		}
	}()
	go func() {
		err := mgr.handleIntermediateKeygenMsgs(keyID, intermediateMsgs)
		if err != nil {
			errChan <- err
		}
	}()
	go func() {
		session.WaitForTimeout()

		if done {
			return
		}

		errChan <- mgr.abortKeygen(keyID)
		mgr.Logger.Info(fmt.Sprintf("aborted keygen protocol %s due to timeout", keyID))
	}()
	go func() {
		err := mgr.handleKeygenResult(keyID, result)
		done = true

		errChan <- err
	}()

	return <-errChan
}

// ProcessKeygenMsg forwards blockchain messages to the keygen protocol
func (mgr *Mgr) ProcessKeygenMsg(e tmEvents.Event) error {
	keyID, from, payload := parseMsgParams(mgr.cdc, e.Attributes)
	msgIn := prepareTrafficIn(mgr.principalAddr, from, keyID, payload, mgr.Logger)

	stream, ok := mgr.getKeygenStream(keyID)
	if !ok {
		mgr.Logger.Info(fmt.Sprintf("no keygen session with id %s. This process does not participate", keyID))
		return nil
	}

	if err := stream.Send(msgIn); err != nil {
		return sdkerrors.Wrap(err, "failure to send incoming msg to gRPC server")
	}
	return nil
}

func parseKeygenAckParams(attributes map[string]string) (keyID string, height int64, err error) {
	parsers := []*parse.AttributeParser{
		{Key: tss.AttributeKeyKeyID, Map: parse.IdentityMap},
		{Key: tss.AttributeKeyHeight, Map: func(s string) (interface{}, error) { return strconv.ParseInt(s, 10, 64) }},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", 0, err
	}

	return results[0].(string), results[1].(int64), nil
}

func parseKeygenStartParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	keyID string, threshold int32, participants []string, participantShareCounts []uint32, timeout int64, err error) {

	parsers := []*parse.AttributeParser{
		{Key: tss.AttributeKeyKeyID, Map: parse.IdentityMap},
		{Key: tss.AttributeKeyThreshold, Map: func(s string) (interface{}, error) {
			t, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return 0, err
			}
			return int32(t), nil
		}},
		{Key: tss.AttributeKeyParticipants, Map: func(s string) (interface{}, error) {
			cdc.MustUnmarshalJSON([]byte(s), &participants)
			return participants, nil
		}},
		{Key: tss.AttributeKeyParticipantShareCounts, Map: func(s string) (interface{}, error) {
			cdc.MustUnmarshalJSON([]byte(s), &participantShareCounts)
			return participantShareCounts, nil
		}},
		{Key: tss.AttributeKeyTimeout, Map: func(s string) (interface{}, error) {
			return strconv.ParseInt(s, 10, 64)
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		return "", 0, nil, nil, 0, err
	}

	return results[0].(string), results[1].(int32), results[2].([]string), results[3].([]uint32), results[4].(int64), nil
}

func (mgr *Mgr) startKeygen(keyID string, threshold int32, myIndex int32, participants []string, participantShareCounts []uint32) (Stream, context.CancelFunc, error) {
	if _, ok := mgr.getKeygenStream(keyID); ok {
		return nil, nil, fmt.Errorf("keygen protocol for ID %s already in progress", keyID)
	}

	grpcCtx, cancel := context.WithTimeout(context.Background(), mgr.Timeout)
	stream, err := mgr.client.Keygen(grpcCtx)
	if err != nil {
		cancel()
		return nil, nil, sdkerrors.Wrap(err, "failed tofnd gRPC call Keygen")
	}

	keygenInit := &tofnd.MessageIn_KeygenInit{
		KeygenInit: &tofnd.KeygenInit{
			NewKeyUid:        keyID,
			Threshold:        threshold,
			PartyUids:        participants,
			PartyShareCounts: participantShareCounts,
			MyPartyIndex:     myIndex,
		},
	}

	if err := stream.Send(&tofnd.MessageIn{Data: keygenInit}); err != nil {
		cancel()
		return nil, nil, err
	}

	return stream, cancel, nil
}

func (mgr *Mgr) handleIntermediateKeygenMsgs(keyID string, intermediate <-chan *tofnd.TrafficOut) error {
	for msg := range intermediate {
		mgr.Logger.Debug(fmt.Sprintf("outgoing keygen msg: key [%.20s] from me [%.20s] to [%.20s] broadcast [%t]\n",
			keyID, mgr.principalAddr, msg.ToPartyUid, msg.IsBroadcast))
		// sender is set by broadcaster
		tssMsg := &tss.ProcessKeygenTrafficRequest{Sender: mgr.cliCtx.FromAddress, SessionID: keyID, Payload: msg}
		refundableMsg := axelarnet.NewRefundMsgRequest(mgr.cliCtx.FromAddress, tssMsg)
		if _, err := mgr.broadcaster.Broadcast(mgr.cliCtx.WithBroadcastMode(sdkFlags.BroadcastSync), refundableMsg); err != nil {
			return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing keygen msg")
		}
	}
	return nil
}

func (mgr *Mgr) handleKeygenResult(keyID string, resultChan <-chan interface{}) error {
	// Delete the reference to the keygen stream with keyID because entering this function means the tss protocol has completed
	defer func() {
		mgr.keygen.Lock()
		defer mgr.keygen.Unlock()
		delete(mgr.keygenStreams, keyID)
	}()

	r, ok := <-resultChan
	if !ok {
		return fmt.Errorf("failed to receive keygen result, channel was closed by the server")
	}

	// get result. Result will be implicity checked by Validate() during Braodcast(), so no checks are needed here
	result, ok := r.(*tofnd.MessageOut_KeygenResult)
	if !ok {
		return fmt.Errorf("failed to receive keygen result, received unexpected type %T", r)
	}

	mgr.Logger.Debug(fmt.Sprintf("handler goroutine: received keygen result for %s [%+v]", keyID, result))

	switch res := result.GetKeygenResultData().(type) {
	case *tofnd.MessageOut_KeygenResult_Criminals:
		// prepare criminals for Validate()
		// criminals have to be sorted in ascending order
		sort.Stable(res.Criminals)
	case *tofnd.MessageOut_KeygenResult_Data:
		if res.Data.GetPubKey() == nil {
			return fmt.Errorf("public key missing from the result")
		}
		if res.Data.GetGroupRecoverInfo() == nil {
			return fmt.Errorf("group recovery data missing from the result")
		}
		if res.Data.GetPrivateRecoverInfo() == nil {
			return fmt.Errorf("private recovery data missing from the result")
		}

		btcecPK, err := btcec.ParsePubKey(res.Data.GetPubKey(), btcec.S256())
		if err != nil {
			return sdkerrors.Wrap(err, "failure to deserialize pubkey")
		}

		mgr.Logger.Info(fmt.Sprintf("handler goroutine: received pubkey from server! [%v]", btcecPK.ToECDSA()))
	default:
		return fmt.Errorf("invalid data type")
	}

	pollKey := voting.NewPollKey(tss.ModuleName, keyID)
	vote := &tss.VotePubKeyRequest{Sender: mgr.cliCtx.FromAddress, PollKey: pollKey, Result: result}
	refundableMsg := axelarnet.NewRefundMsgRequest(mgr.cliCtx.FromAddress, vote)

	_, err := mgr.broadcaster.Broadcast(mgr.cliCtx.WithBroadcastMode(sdkFlags.BroadcastBlock), refundableMsg)
	return err
}

func (mgr *Mgr) getKeygenStream(keyID string) (Stream, bool) {
	mgr.keygen.RLock()
	defer mgr.keygen.RUnlock()

	stream, ok := mgr.keygenStreams[keyID]
	return stream, ok
}

func (mgr *Mgr) setKeygenStream(keyID string, stream Stream) {
	mgr.keygen.Lock()
	defer mgr.keygen.Unlock()

	mgr.keygenStreams[keyID] = NewLockableStream(stream)
}
