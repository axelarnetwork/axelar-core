package tss

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// ProcessKeygenStart starts the communication with the keygen protocol
func (mgr *Mgr) ProcessKeygenStart(blockHeight int64, attributes []sdk.Attribute) error {
	keyID, threshold, participants, participantShareCounts := parseKeygenStartParams(mgr.cdc, attributes)
	myIndex, ok := indexOf(participants, mgr.principalAddr)
	if !ok {
		// do not participate
		return nil
	}

	done := false
	session := mgr.timeoutQueue.Enqueue(keyID, blockHeight+mgr.sessionTimeout)

	stream, cancel, err := mgr.startKeygen(keyID, threshold, myIndex, participants, participantShareCounts)
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
func (mgr *Mgr) ProcessKeygenMsg(attributes []sdk.Attribute) error {
	keyID, from, payload := parseMsgParams(mgr.cdc, attributes)
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

func parseKeygenStartParams(cdc *codec.LegacyAmino, attributes []sdk.Attribute) (keyID string, threshold int32, participants []string, participantShareCounts []uint32) {
	for _, attribute := range attributes {
		switch attribute.Key {
		case tss.AttributeKeyKeyID:
			keyID = attribute.Value
		case tss.AttributeKeyThreshold:
			t, err := strconv.ParseInt(attribute.Value, 10, 32)
			if err != nil {
				panic(err)
			}
			threshold = int32(t)
		case tss.AttributeKeyParticipants:
			cdc.MustUnmarshalJSON([]byte(attribute.Value), &participants)
		case tss.AttributeKeyParticipantShareCounts:
			cdc.MustUnmarshalJSON([]byte(attribute.Value), &participantShareCounts)
		default:
		}
	}

	return keyID, threshold, participants, participantShareCounts
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
		tssMsg := &tss.ProcessKeygenTrafficRequest{Sender: mgr.sender, SessionID: keyID, Payload: msg}
		if err := mgr.broadcaster.Broadcast(tssMsg); err != nil {
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

	result := (<-resultChan).(*tofnd.MessageOut_KeygenResult)
	if result.GetCriminals() != nil {
		// criminals have to be sorted in ascending order
		sort.Stable(result.GetCriminals())
	}

	mgr.Logger.Debug(fmt.Sprintf("handler goroutine: received keygen result for %s [%+v]", keyID, result))

	pubKeyBytes := result.GetPubkey()
	if pubKeyBytes != nil {
		btcecPK, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
		if err != nil {
			return sdkerrors.Wrap(err, "handler goroutine: failure to deserialize pubkey")
		}

		mgr.Logger.Info(fmt.Sprintf("handler goroutine: received pubkey from server! [%v]", btcecPK.ToECDSA()))
	}

	poll := voting.NewPollMeta(tss.ModuleName, keyID)
	vote := &tss.VotePubKeyRequest{Sender: mgr.sender, PollMeta: poll, Result: result}

	return mgr.broadcaster.Broadcast(vote)
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
