package tss

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/parse"
	"github.com/axelarnetwork/axelar-core/utils"
	tssexported "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
	tmEvents "github.com/axelarnetwork/tm-events/events"
)

// ProcessKeygenStart starts the communication with the keygen protocol
func (mgr *Mgr) ProcessKeygenStart(e tmEvents.Event) error {
	keyType, keyID, threshold, participants, participantShareCounts, timeout, err := parseKeygenStartParams(mgr.cdc, e.Attributes)
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
		return mgr.thresholdKeygenStart(e, keyID, timeout, threshold, myIndex, participants, participantShareCounts)
	case tssexported.Multisig.SimpleString():
		return mgr.multiSigKeygenStart(keyID, participantShareCounts[myIndex])
	default:
		return fmt.Errorf(fmt.Sprintf("unknown keytype %s", keyType))
	}
}

func (mgr *Mgr) thresholdKeygenStart(e tmEvents.Event, keyID string, timeout int64, threshold uint32, myIndex int, participants []string, participantShareCounts []uint32) error {
	done := false
	session := mgr.timeoutQueue.Enqueue(keyID, e.Height+timeout)

	stream, cancel, err := mgr.startKeygen(keyID, threshold, uint32(myIndex), participants, participantShareCounts)
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

func (mgr *Mgr) multiSigKeygenStart(keyID string, shares uint32) error {
	grpcCtx, cancel := context.WithTimeout(context.Background(), mgr.Timeout)
	defer cancel()

	var sigKeyPairs []tssexported.SigKeyPair
	pubKeys := make([][]byte, shares)

	for i := uint32(0); i < shares; i++ {
		keyUID := fmt.Sprintf("%s_%d", keyID, i)
		keygenRequest := &tofnd.KeygenRequest{
			KeyUid:   keyUID,
			PartyUid: mgr.principalAddr,
		}

		res, err := mgr.multiSigClient.Keygen(grpcCtx, keygenRequest)
		if err != nil {
			return sdkerrors.Wrapf(err, "failed to generate multisig key")
		}

		switch res.GetKeygenResponse().(type) {
		case *tofnd.KeygenResponse_PubKey:
			//  proof validator owns the pub key
			d := sha256.Sum256([]byte(mgr.principalAddr))
			sig, err := mgr.multiSigSign(keyUID, d[:], res.GetPubKey())
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to sign")
			}
			sigKeyPairs = append(sigKeyPairs, tssexported.SigKeyPair{PubKey: res.GetPubKey(), Signature: sig})
			pubKeys[i] = res.GetPubKey()
		case *tofnd.KeygenResponse_Error:
			return sdkerrors.Wrap(err, res.GetError())
		default:
			return sdkerrors.Wrap(err, "unknown multisig keygen response")
		}
	}

	// TODO: Evict keys older than X rotations (they can be retrieved again if needed)
	mgr.setKey(keyID, pubKeys)

	mgr.Logger.Info(fmt.Sprintf("operator %s sending multisig keys for key %s", mgr.principalAddr, keyID))
	tssMsg := tss.NewSubmitMultiSigPubKeysRequest(mgr.cliCtx.FromAddress, tssexported.KeyID(keyID), sigKeyPairs)

	if _, err := mgr.broadcaster.Broadcast(context.TODO(), tssMsg); err != nil {
		return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing multisig keys msg")
	}

	return nil
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

func parseKeygenStartParams(cdc *codec.LegacyAmino, attributes map[string]string) (
	keyType, keyID string, threshold uint32, participants []string, participantShareCounts []uint32, timeout int64, err error) {

	parsers := []*parse.AttributeParser{
		{Key: tss.AttributeKeyKeyType, Map: parse.IdentityMap},
		{Key: tss.AttributeKeyKeyID, Map: parse.IdentityMap},
		{Key: tss.AttributeKeyThreshold, Map: func(s string) (interface{}, error) {
			t, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return 0, err
			}
			return uint32(t), nil
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
		return "", "", 0, nil, nil, 0, err
	}

	return results[0].(string), results[1].(string), results[2].(uint32), results[3].([]string), results[4].([]uint32), results[5].(int64), nil
}

func (mgr *Mgr) startKeygen(keyID string, threshold uint32, myIndex uint32, participants []string, participantShareCounts []uint32) (Stream, context.CancelFunc, error) {
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
		tssMsg := &tss.ProcessKeygenTrafficRequest{Sender: mgr.cliCtx.FromAddress, SessionID: keyID, Payload: *msg}
		if _, err := mgr.broadcaster.Broadcast(context.TODO(), tssMsg); err != nil {
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
	vote := &tss.VotePubKeyRequest{Sender: mgr.cliCtx.FromAddress, PollKey: pollKey, Result: *result}
	_, err := mgr.broadcaster.Broadcast(context.TODO(), vote)
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
