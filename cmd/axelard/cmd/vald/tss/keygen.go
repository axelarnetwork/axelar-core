package tss

import (
	"context"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/third_party/proto/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// ProcessKeygenStart starts the communication with the keygen protocol
func (mgr *Mgr) ProcessKeygenStart(attributes []sdk.Attribute) error {
	keyID, threshold, participants := parseKeygenStartParams(mgr.cdc, attributes)
	myIndex, ok := indexOf(participants, mgr.principalAddr)
	if !ok {
		// do not participate
		return nil
	}

	stream, cancel, err := mgr.startKeygen(keyID, threshold, myIndex, participants)
	if err != nil {
		return err
	}
	mgr.setKeygenStream(keyID, stream)

	// use error channel to coordinate errors during communication with sign protocol
	errChan := make(chan error, 3)
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
		err := mgr.handleKeygenResult(keyID, result)
		if err != nil {
			errChan <- err
		} else {
			// this is the last part of the sign, so if there are no errors here return nil
			errChan <- nil
		}
	}()
	return <-errChan
}

// ProcessKeygenMsg forwards blockchain messages to the keygen protocol
func (mgr *Mgr) ProcessKeygenMsg(attributes []sdk.Attribute) error {
	keyID, from, payload := parseMsgParams(mgr.cdc, attributes)
	msgIn, err := prepareTrafficIn(mgr.principalAddr, from, keyID, payload, mgr.Logger)
	if err != nil {
		return err
	}
	// this message is not meant for this tofnd instance
	if msgIn == nil {
		return nil
	}

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

func parseKeygenStartParams(cdc *codec.LegacyAmino, attributes []sdk.Attribute) (keyID string, threshold int32, participants []string) {
	for _, attribute := range attributes {
		switch attribute.Key {
		case tss.AttributeKeyKeyID:
			keyID = attribute.Value
		case tss.AttributeKeyThreshold:
			t, err := strconv.Atoi(attribute.Value)
			if err != nil {
				panic(err)
			}
			threshold = int32(t)
		case tss.AttributeKeyParticipants:
			cdc.MustUnmarshalJSON([]byte(attribute.Value), &participants)
		default:
		}
	}

	return keyID, threshold, participants
}

func (mgr *Mgr) startKeygen(keyID string, threshold int32, myIndex int32, participants []string) (tss.Stream, context.CancelFunc, error) {
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
			NewKeyUid:    keyID,
			Threshold:    threshold,
			PartyUids:    participants,
			MyPartyIndex: myIndex,
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
		tssMsg := &tss.MsgKeygenTraffic{Sender: mgr.sender, SessionID: keyID, Payload: msg}
		if err := mgr.broadcaster.Broadcast(tssMsg); err != nil {
			return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing keygen msg")
		}
	}
	return nil
}

func (mgr *Mgr) handleKeygenResult(keyID string, result <-chan []byte) error {
	// Delete the reference to the keygen stream with keyID because entering this function means the tss protocol has completed
	defer func() {
		mgr.keygen.Lock()
		defer mgr.keygen.Unlock()
		delete(mgr.keygenStreams, keyID)
	}()

	bz := <-result
	btcecPK, err := btcec.ParsePubKey(bz, btcec.S256())
	if err != nil {
		return sdkerrors.Wrap(err, "handler goroutine: failure to deserialize pubkey")
	}
	pubkey := btcecPK.ToECDSA()

	mgr.Logger.Info(fmt.Sprintf("handler goroutine: received pubkey from server! [%v]", pubkey))

	poll := voting.NewPollMeta(tss.ModuleName, keyID)
	vote := &tss.MsgVotePubKey{Sender: mgr.sender, PollMeta: poll, PubKeyBytes: bz}
	return mgr.broadcaster.Broadcast(vote)
}

func (mgr *Mgr) getKeygenStream(keyID string) (tss.Stream, bool) {
	mgr.keygen.RLock()
	defer mgr.keygen.RUnlock()

	stream, ok := mgr.keygenStreams[keyID]
	return stream, ok
}

func (mgr *Mgr) setKeygenStream(keyID string, stream tss.Stream) {
	mgr.keygen.Lock()
	defer mgr.keygen.Unlock()

	mgr.keygenStreams[keyID] = stream
}
