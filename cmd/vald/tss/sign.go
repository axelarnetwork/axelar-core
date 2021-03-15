package tss

import (
	"context"
	"fmt"

	"github.com/axelarnetwork/c2d2/pkg/pubsub"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	voting "github.com/axelarnetwork/axelar-core/x/vote/exported"
)

func (mgr *TSSMgr) ProcessSign(subscriber pubsub.Subscriber, errChan chan<- error) {
	for {
		select {
		case event := <-subscriber.Events():
			switch e := event.(type) {
			case types.Event:
				// all events of the transaction are returned, so need to filter for sign
				if e.Type != tss.EventTypeSign {
					continue
				}
				switch e.Action {
				case tss.AttributeValueStart:
					keyID, sigID, participants, payload := parseSignStartParams(e.Attributes)
					_, ok := mgr.findMyIndex(participants)
					if !ok {
						// do not participate
						continue
					}
					stream, cancel, err := mgr.startSign(keyID, sigID, participants, payload)
					if err != nil {
						errChan <- err
						continue
					}
					mgr.signStreams[sigID] = stream

					intermediateMsgs, result := handleStream(stream, cancel, errChan, mgr.Logger)
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
						}
					}()
				case tss.AttributeValueMsg:
					sigID, from, payload := parseMsgParams(e.Attributes)
					err := mgr.processSignMsg(sigID, from, payload)
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

func (mgr *TSSMgr) startSign(keyID string, sigID string, participants []string, payload []byte) (tss.Stream, context.CancelFunc, error) {
	if _, ok := mgr.signStreams[sigID]; ok {
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

func (mgr *TSSMgr) handleIntermediateSignMsgs(sigID string, intermediate <-chan *tofnd.TrafficOut) error {
	for msg := range intermediate {
		mgr.Logger.Debug(fmt.Sprintf("outgoing sign msg: sig [%.20s] from me [%.20s] to [%.20s] broadcast [%t]\n",
			sigID, mgr.myAddress, msg.ToPartyUid, msg.IsBroadcast))
		// sender is set by broadcaster
		tssMsg := &tss.MsgSignTraffic{Sender: mgr.sender, SessionID: sigID, Payload: msg}
		if err := mgr.broadcaster.Broadcast(tssMsg); err != nil {
			return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing sign msg")
		}
	}
	return nil
}

func (mgr *TSSMgr) handleSignResult(sigID string, result <-chan []byte) error {
	// Delete the reference to the signing stream with sigID because entering this function means the tss protocol has completed
	defer delete(mgr.signStreams, sigID)

	bz := <-result
	mgr.Logger.Info(fmt.Sprintf("handler goroutine: received sig from server! [%.20s]", bz))

	poll := voting.PollMeta{Module: tss.ModuleName, Type: tss.EventTypeSign, ID: sigID}
	vote := &tss.MsgVoteSig{Sender: mgr.sender, PollMeta: poll, SigBytes: bz}
	return mgr.broadcaster.Broadcast(vote)
}

func (mgr *TSSMgr) processSignMsg(sigID string, from string, payload *tofnd.TrafficOut) error {
	msgIn, err := prepareTrafficIn(mgr.myAddress, from, sigID, payload, mgr.Logger)
	if err != nil {
		return err
	}
	// this message is not meant for this tofnd instance
	if msgIn == nil {
		return nil
	}

	stream, ok := mgr.signStreams[sigID]
	if !ok {
		mgr.Logger.Info(fmt.Sprintf("no sign session with id %s. This process does not participate", sigID))
		return nil
	}

	if err := stream.Send(msgIn); err != nil {
		return sdkerrors.Wrap(err, "failure to send incoming msg to gRPC server")
	}
	return nil
}
