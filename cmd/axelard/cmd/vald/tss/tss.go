package tss

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	rewardtypes "github.com/axelarnetwork/axelar-core/x/reward/types"
	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	broadcasterTypes "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/parse"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss/rpc"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
	tmEvents "github.com/axelarnetwork/tm-events/events"
)

// Session defines a tss session which is either signing or keygen
type Session struct {
	ID        string
	TimeoutAt int64
	timeout   chan struct{}
}

// Timeout signals a session has timed out
func (s *Session) Timeout() {
	close(s.timeout)
}

// WaitForTimeout waits until the session has timed out
func (s *Session) WaitForTimeout() {
	<-s.timeout
}

// TimeoutQueue is a queue of sessions order by timeoutAt
type TimeoutQueue struct {
	lock  sync.RWMutex
	queue []*Session
}

// Enqueue adds a new session with ID and timeoutAt into the queue
func (q *TimeoutQueue) Enqueue(ID string, timeoutAt int64) *Session {
	q.lock.Lock()
	defer q.lock.Unlock()

	session := Session{ID: ID, TimeoutAt: timeoutAt, timeout: make(chan struct{})}
	q.queue = append(q.queue, &session)

	return &session
}

// Dequeue pops the first session in queue
func (q *TimeoutQueue) Dequeue() *Session {
	q.lock.Lock()
	defer q.lock.Unlock()

	if len(q.queue) == 0 {
		return nil
	}

	result := q.queue[0]
	q.queue = q.queue[1:]

	return result
}

// Top returns the first session in queue
func (q *TimeoutQueue) Top() *Session {
	q.lock.RLock()
	defer q.lock.RUnlock()

	if len(q.queue) == 0 {
		return nil
	}

	return q.queue[0]
}

// NewTimeoutQueue is the constructor for TimeoutQueue
func NewTimeoutQueue() *TimeoutQueue {
	return &TimeoutQueue{
		lock:  sync.RWMutex{},
		queue: []*Session{},
	}
}

// Stream is the abstracted communication stream with tofnd
type Stream interface {
	Send(in *tofnd.MessageIn) error
	Recv() (*tofnd.MessageOut, error)
	CloseSend() error
}

// LockableStream is a thread-safe Stream
type LockableStream struct {
	sendLock sync.Mutex
	recvLock sync.Mutex
	stream   Stream
}

// NewLockableStream return a new thread-safe stream instance
func NewLockableStream(stream Stream) *LockableStream {
	return &LockableStream{
		sendLock: sync.Mutex{},
		recvLock: sync.Mutex{},
		stream:   stream,
	}
}

// Send implements the Stream interface
func (l *LockableStream) Send(in *tofnd.MessageIn) error {
	l.sendLock.Lock()
	defer l.sendLock.Unlock()

	return l.stream.Send(in)
}

// Recv implements the Stream interface
func (l *LockableStream) Recv() (*tofnd.MessageOut, error) {
	l.recvLock.Lock()
	defer l.recvLock.Unlock()

	return l.stream.Recv()
}

// CloseSend implements the Stream interface
func (l *LockableStream) CloseSend() error {
	l.sendLock.Lock()
	defer l.sendLock.Unlock()

	return l.stream.CloseSend()
}

// Mgr represents an object that manages all communication with the external tss process
type Mgr struct {
	client         rpc.Client
	multiSigClient rpc.MultiSigClient
	cliCtx         sdkClient.Context
	keygen         *sync.RWMutex
	sign           *sync.RWMutex
	keygenStreams  map[string]*LockableStream
	signStreams    map[string]*LockableStream
	timeoutQueue   *TimeoutQueue
	Timeout        time.Duration
	principalAddr  string
	Logger         log.Logger
	broadcaster    broadcasterTypes.Broadcaster
	cdc            *codec.LegacyAmino
}

// Connect connects to tofnd gRPC Server
func Connect(host string, port string, timeout time.Duration, logger log.Logger) (*grpc.ClientConn, error) {
	tofndServerAddress := host + ":" + port
	logger.Info(fmt.Sprintf("initiate connection to tofnd gRPC server: %s", tofndServerAddress))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return grpc.DialContext(ctx, tofndServerAddress, grpc.WithInsecure(), grpc.WithBlock())
}

// NewMgr returns a new tss manager instance
func NewMgr(client rpc.Client, multiSigClient rpc.MultiSigClient, cliCtx sdkClient.Context, timeout time.Duration, principalAddr string, broadcaster broadcasterTypes.Broadcaster, logger log.Logger, cdc *codec.LegacyAmino) *Mgr {
	return &Mgr{
		client:         client,
		multiSigClient: multiSigClient,
		cliCtx:         cliCtx,
		keygen:         &sync.RWMutex{},
		sign:           &sync.RWMutex{},
		keygenStreams:  make(map[string]*LockableStream),
		signStreams:    make(map[string]*LockableStream),
		timeoutQueue:   NewTimeoutQueue(),
		Timeout:        timeout,
		principalAddr:  principalAddr,
		Logger:         logger.With("listener", "tss"),
		broadcaster:    broadcaster,
		cdc:            cdc,
	}
}

// Recover instructs tofnd to recover the node's shares given the recovery info provided
func (mgr *Mgr) Recover(recoverJSON []byte) error {
	var requests []tofnd.RecoverRequest
	err := mgr.cdc.UnmarshalJSON(recoverJSON, &requests)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to unmarshal recovery requests")
	}

	for _, request := range requests {
		uid := request.KeygenInit.PartyUids[int(request.KeygenInit.MyPartyIndex)]

		if mgr.principalAddr != uid {
			return fmt.Errorf("party UID mismatch (expected %s, got %s)", mgr.principalAddr, uid)
		}

		grpcCtx, cancel := context.WithTimeout(context.Background(), mgr.Timeout)
		defer cancel()

		response, err := mgr.client.Recover(grpcCtx, &request)
		if err != nil {
			return sdkerrors.Wrap(err,
				fmt.Sprintf("failed tofnd gRPC call Recover for key ID %s", request.KeygenInit.NewKeyUid))
		}

		if response.GetResponse() == tofnd.RecoverResponse_RESPONSE_FAIL {
			return fmt.Errorf("failed to recover tofnd shares for validator %s and key ID %s", uid, request.KeygenInit.NewKeyUid)
		}
		mgr.Logger.Info(
			fmt.Sprintf("successfully recovered tofnd shares for validator %s and key ID %s", uid, request.KeygenInit.NewKeyUid))
	}

	return nil
}

// ProcessHeartBeatEvent broadcasts the heartbeat
func (mgr *Mgr) ProcessHeartBeatEvent(e tmEvents.Event) error {
	grpcCtx, cancel := context.WithTimeout(context.Background(), mgr.Timeout)
	defer cancel()

	// tofnd health check using a dummy ID
	// TODO: we should have a specific GRPC to do this diagnostic
	request := &tofnd.KeyPresenceRequest{
		KeyUid: "dummyID",
	}

	response, err := mgr.client.KeyPresence(grpcCtx, request)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to invoke KeyPresence grpc")
	}

	switch response.Response {
	case tofnd.RESPONSE_UNSPECIFIED, tofnd.RESPONSE_FAIL:
		return sdkerrors.Wrap(err, "tofnd not set up correctly")
	case tofnd.RESPONSE_PRESENT, tofnd.RESPONSE_ABSENT:
		// tofnd is working properly
	default:
		return sdkerrors.Wrap(err, "unknown tofnd response")
	}

	// check for keys presence according to the IDs included in the event
	keyInfos := parseHeartBeatParams(mgr.cdc, e.Attributes)
	var present []exported.KeyID

	for _, keyInfo := range keyInfos {

		grpcCtx, cancel = context.WithTimeout(context.Background(), mgr.Timeout)
		defer cancel()

		switch keyInfo.KeyType {
		case exported.Threshold:
			request = &tofnd.KeyPresenceRequest{
				KeyUid: string(keyInfo.KeyID),
			}
			response, err = mgr.client.KeyPresence(grpcCtx, request)
		case exported.Multisig:
			request = &tofnd.KeyPresenceRequest{
				KeyUid: fmt.Sprintf("%s_%d", string(keyInfo.KeyID), 0),
			}
			response, err = mgr.multiSigClient.KeyPresence(grpcCtx, request)
		default:
			return sdkerrors.Wrapf(err, fmt.Sprintf("unrecognize key type %s", keyInfo.KeyType.SimpleString()))
		}

		if err != nil {
			return sdkerrors.Wrapf(err, "failed to invoke KeyPresence grpc")
		}

		switch response.Response {
		case tofnd.RESPONSE_UNSPECIFIED, tofnd.RESPONSE_FAIL:
			return sdkerrors.Wrap(err, "tofnd not set up correctly")
		case tofnd.RESPONSE_ABSENT:
			// key is absent
		case tofnd.RESPONSE_PRESENT:
			present = append(present, keyInfo.KeyID)
		default:
			return sdkerrors.Wrap(err, "unknown tofnd response")
		}
	}

	tssMsg := tss.NewHeartBeatRequest(mgr.cliCtx.FromAddress, present)
	refundableMsg := rewardtypes.NewRefundMsgRequest(mgr.cliCtx.FromAddress, tssMsg)

	mgr.Logger.Info(fmt.Sprintf("operator %s sending heartbeat acknowledging keys: %s", mgr.principalAddr, present))
	txRes, err := mgr.broadcaster.Broadcast(context.TODO(), refundableMsg)
	if err != nil {
		return sdkerrors.Wrap(err, "handler goroutine: failure to broadcast outgoing heartbeat msg")
	}

	heartbeats, err := mgr.extractHeartBeatResponses(txRes)
	if err != nil {
		return err
	}

	for _, heartbeat := range heartbeats {
		allGood := true
		if !heartbeat.KeygenIllegibility.Is(snapshot.None) {
			mgr.Logger.Error(fmt.Sprintf("operator %s unable to participate in keygen due to: %s",
				mgr.principalAddr, heartbeat.KeygenIllegibility.String()))
			allGood = false
		}
		if !heartbeat.SigningIllegibility.Is(snapshot.None) {
			mgr.Logger.Error(fmt.Sprintf("operator %s unable to participate in signing due to: %s",
				mgr.principalAddr, heartbeat.SigningIllegibility.String()))
			allGood = false
		}

		if allGood {
			mgr.Logger.Info(fmt.Sprintf("no keygen/signing issues reported for operator %s", mgr.principalAddr))
		}
	}
	return nil
}

func (mgr *Mgr) extractHeartBeatResponses(txRes *sdk.TxResponse) ([]tss.HeartBeatResponse, error) {
	responses, err := mgr.extractRefundMsgResponses(txRes)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "handler goroutine: failure to retrieve refund reply")
	}

	var heartbeatRes []tss.HeartBeatResponse
	for _, res := range responses {
		var heartBeat tss.HeartBeatResponse
		if err := mgr.cliCtx.Codec.Unmarshal(res.Data, &heartBeat); err != nil {
			continue
		}
		heartbeatRes = append(heartbeatRes, heartBeat)
	}

	if len(heartbeatRes) == 0 {
		return nil, fmt.Errorf("handler goroutine: failure to retrieve heartbeat reply")
	}
	return heartbeatRes, nil
}

func (mgr *Mgr) abortSign(sigID string) (err error) {
	stream, ok := mgr.getSignStream(sigID)
	if !ok {
		return nil
	}

	return abort(stream)
}

func (mgr *Mgr) abortKeygen(keyID string) (err error) {
	stream, ok := mgr.getKeygenStream(keyID)
	if !ok {
		return nil
	}

	return abort(stream)
}

func (mgr *Mgr) extractRefundMsgResponses(res *sdk.TxResponse) ([]rewardtypes.RefundMsgResponse, error) {
	var txMsg sdk.TxMsgData
	var refundReply []rewardtypes.RefundMsgResponse

	bz, err := hex.DecodeString(res.Data)
	if err != nil {
		return nil, err
	}

	err = mgr.cdc.Unmarshal(bz, &txMsg)
	if err != nil {
		return nil, err
	}

	for _, msg := range txMsg.Data {
		var refund rewardtypes.RefundMsgResponse
		err = refund.Unmarshal(msg.Data)
		if err != nil {
			mgr.Logger.Debug(fmt.Sprintf("not a refund response: %v", err))
			continue
		}
		refundReply = append(refundReply, refund)
	}

	return refundReply, nil
}

func abort(stream Stream) error {
	msg := &tofnd.MessageIn{
		Data: &tofnd.MessageIn_Abort{
			Abort: true,
		},
	}

	if err := stream.Send(msg); err != nil {
		return sdkerrors.Wrap(err, "failure to send abort msg to gRPC server")
	}

	return nil
}

func handleStream(stream Stream, cancel context.CancelFunc, logger log.Logger) (broadcast <-chan *tofnd.TrafficOut, result <-chan interface{}, err <-chan error) {
	broadcastChan := make(chan *tofnd.TrafficOut)
	// TODO: MessageOut_KeygenResult and MessageOut_SignResult should be merged into one type of message
	resChan := make(chan interface{})
	errChan := make(chan error, 2)

	// server handler https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc-1
	go func() {
		defer cancel()
		defer close(errChan)
		defer close(broadcastChan)
		defer close(resChan)
		defer func() {
			// close the stream on error or protocol completion
			if err := stream.CloseSend(); err != nil {
				errChan <- sdkerrors.Wrap(err, "handler goroutine: failure to CloseSend stream")
			}
		}()

		for {
			msgOneof, err := stream.Recv() // blocking
			if err == io.EOF {             // output stream closed by server
				logger.Debug("handler goroutine: gRPC stream closed by server")
				return
			}
			if err != nil {
				errChan <- sdkerrors.Wrap(err, "handler goroutine: failure to receive msg from gRPC server stream")
				return
			}

			switch msg := msgOneof.GetData().(type) {
			case *tofnd.MessageOut_Traffic:
				broadcastChan <- msg.Traffic
			case *tofnd.MessageOut_KeygenResult_:
				resChan <- msg.KeygenResult
				return
			case *tofnd.MessageOut_SignResult_:
				resChan <- msg.SignResult
				return
			default:
				errChan <- fmt.Errorf("handler goroutine: server stream should send only msg type")
				return
			}
		}
	}()
	return broadcastChan, resChan, errChan
}

func parseHeartBeatParams(cdc *codec.LegacyAmino, attributes map[string]string) []tss.KeyInfo {
	parsers := []*parse.AttributeParser{
		{Key: tss.AttributeKeyKeyInfos, Map: func(s string) (interface{}, error) {
			var keyInfos []tss.KeyInfo
			cdc.MustUnmarshalJSON([]byte(s), &keyInfos)
			return keyInfos, nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		panic(err)
	}

	return results[0].([]tss.KeyInfo)
}

func parseMsgParams(cdc *codec.LegacyAmino, attributes map[string]string) (sessionID string, from string, payload *tofnd.TrafficOut) {
	parsers := []*parse.AttributeParser{
		{Key: tss.AttributeKeySessionID, Map: parse.IdentityMap},
		{Key: sdk.AttributeKeySender, Map: parse.IdentityMap},
		{Key: tss.AttributeKeyPayload, Map: func(s string) (interface{}, error) {
			cdc.MustUnmarshalJSON([]byte(s), &payload)
			return payload, nil
		}},
	}

	results, err := parse.Parse(attributes, parsers)
	if err != nil {
		panic(err)
	}

	return results[0].(string), results[1].(string), results[2].(*tofnd.TrafficOut)

}

func prepareTrafficIn(principalAddr string, from string, sessionID string, payload *tofnd.TrafficOut, logger log.Logger) *tofnd.MessageIn {
	msgIn := &tofnd.MessageIn{
		Data: &tofnd.MessageIn_Traffic{
			Traffic: &tofnd.TrafficIn{
				Payload:      payload.Payload,
				IsBroadcast:  payload.IsBroadcast,
				FromPartyUid: from,
			},
		},
	}

	logger.Debug(fmt.Sprintf("incoming msg to tofnd: session [%.20s] from [%.20s] to [%.20s] broadcast [%t] me [%.20s]",
		sessionID, from, payload.ToPartyUid, payload.IsBroadcast, principalAddr))

	return msgIn
}
