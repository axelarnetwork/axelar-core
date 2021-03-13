package tss

import (
	"context"
	"fmt"
	"strconv"

	"github.com/axelarnetwork/c2d2/pkg/pubsub"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/types"
	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// ProcessKeygen manages all communication for keygen protocols between axelar and the external tss process
func (mgr *Mgr) ProcessKeygen(subscriber pubsub.Subscriber, errChan chan<- error) {
	for {
		select {
		case event := <-subscriber.Events():
			switch e := event.(type) {
			case types.Event:
				// all events of the transaction are returned, so need to filter for keygen
				if e.Type != tss.EventTypeKeygen {
					continue
				}
				switch e.Action {
				case tss.AttributeValueStart:
					keyID, threshold, participants := parseKeygenStartParams(e.Attributes)
					myIndex, ok := mgr.findMyIndex(participants)
					if !ok {
						// do not participate
						continue
					}

					stream, cancel, err := mgr.startKeygen(keyID, threshold, myIndex, participants)
					if err != nil {
						errChan <- err
						continue
					}
					mgr.keygenStreams[keyID] = stream

					intermediateMsgs, result := handleStream(stream, cancel, errChan, mgr.Logger)
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
						}
					}()
				case tss.AttributeValueMsg:
					keyID, from, payload := parseMsgParams(e.Attributes)
					err := mgr.forwardKeygenMsg(keyID, from, payload)
					if err != nil {
						errChan <- err
					}
				}
			default:
				panic(fmt.Sprintf("unexpected event type %t", event))
			}
		case <-subscriber.Done():
			break
		}
	}
}

func parseKeygenStartParams(attributes []sdk.Attribute) (keyID string, threshold int32, participants []string) {
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
			codec.Cdc.MustUnmarshalJSON([]byte(attribute.Value), &participants)
		default:
		}
	}

	return keyID, threshold, participants
}

func (mgr *Mgr) startKeygen(keyID string, threshold int32, myIndex int32, participants []string) (tss.Stream, context.CancelFunc, error) {
	if _, ok := mgr.keygenStreams[keyID]; ok {
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
			keyID, mgr.myAddress, msg.ToPartyUid, msg.IsBroadcast))
		// sender is set by broadcaster
		tssMsg := tss.MsgKeygenTraffic{Sender: mgr.sender, SessionID: keyID, Payload: msg}
		if err := <-mgr.broadcaster.Broadcast(tssMsg); err != nil {
			return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing keygen msg")
		}
	}
	return nil
}

func (mgr *Mgr) handleKeygenResult(keyID string, result <-chan []byte) error {
	// Delete the reference to the keygen stream with keyID because entering this function means the tss protocol has completed
	defer delete(mgr.keygenStreams, keyID)

	bz := <-result
	btcecPK, err := btcec.ParsePubKey(bz, btcec.S256())
	if err != nil {
		return sdkerrors.Wrap(err, "handler goroutine: failure to deserialize pubkey")
	}
	pubkey := btcecPK.ToECDSA()

	mgr.Logger.Info(fmt.Sprintf("handler goroutine: received pubkey from server! [%v]", pubkey))

	poll := voting.NewPollMeta(tss.ModuleName, tss.EventTypeKeygen, keyID)
	vote := &tss.MsgVotePubKey{Sender: mgr.sender, PollMeta: poll, PubKeyBytes: bz}
	return <-mgr.broadcaster.Broadcast(vote)
}

func (mgr *Mgr) forwardKeygenMsg(keyID string, from string, payload *tofnd.TrafficOut) error {
	msgIn, err := prepareTrafficIn(mgr.myAddress, from, keyID, payload, mgr.Logger)
	if err != nil {
		return err
	}
	// this message is not meant for this tofnd instance
	if msgIn == nil {
		return nil
	}

	stream, ok := mgr.keygenStreams[keyID]
	if !ok {
		mgr.Logger.Info(fmt.Sprintf("no keygen session with id %s. This process does not participate", keyID))
		return nil
	}

	if err := stream.Send(msgIn); err != nil {
		return sdkerrors.Wrap(err, "failure to send incoming msg to gRPC server")
	}
	return nil
}
