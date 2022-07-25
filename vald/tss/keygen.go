package tss

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/vald/parse"
	tssexported "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	tmEvents "github.com/axelarnetwork/tm-events/events"
)

// ProcessKeygenStart starts the communication with the keygen protocol
func (mgr *Mgr) ProcessKeygenStart(e tmEvents.Event) error {
	keyType, keyID, _, participants, participantShareCounts, _, err := parseKeygenStartParams(mgr.cdc, e.Attributes)
	if err != nil {
		return err
	}

	myIndex := utils.IndexOf(participants, mgr.principalAddr)
	if myIndex == -1 {
		// do not participate
		return nil
	}

	switch keyType {
	case tssexported.Multisig.SimpleString():
		return mgr.multiSigKeygenStart(keyID, participantShareCounts[myIndex])
	default:
		return fmt.Errorf(fmt.Sprintf("unknown keytype %s", keyType))
	}
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
