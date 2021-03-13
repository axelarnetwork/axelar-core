package tss

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/cmd/vald/broadcast"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// Mgr represents an object that manages all communication with the external tss process
type Mgr struct {
	client        tofnd.GG20Client
	keygenStreams map[string]tss.Stream
	signStreams   map[string]tss.Stream
	Timeout       time.Duration
	myAddress     string
	Logger        log.Logger
	broadcaster   *broadcast.XBOBroadcaster
	sender        sdk.AccAddress
}

// CreateTOFNDClient creates a client to communicate with the external tofnd process
func CreateTOFNDClient(host string, port string, logger log.Logger) (tofnd.GG20Client, error) {
	tofndServerAddress := host + ":" + port
	logger.Info(fmt.Sprintf("initiate connection to tofnd gRPC server: %s", tofndServerAddress))
	conn, err := grpc.Dial(tofndServerAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	logger.Debug("successful connection to tofnd gRPC server")
	gg20client := tofnd.NewGG20Client(conn)
	return gg20client, nil
}

// NewMgr returns a new tss manager instance
func NewMgr(client tofnd.GG20Client, timeout time.Duration, myAddress string, broadcaster *broadcast.XBOBroadcaster, sender sdk.AccAddress, logger log.Logger) *Mgr {
	return &Mgr{
		client:        client,
		keygenStreams: map[string]tss.Stream{},
		signStreams:   map[string]tss.Stream{},
		Timeout:       timeout,
		myAddress:     myAddress,
		Logger:        logger.With("listener", "tss"),
		broadcaster:   broadcaster,
		sender:        sender,
	}
}

func handleStream(stream tss.Stream, cancel context.CancelFunc, errChan chan<- error, logger log.Logger) (broadcast <-chan *tofnd.TrafficOut, result <-chan []byte) {
	broadcastChan := make(chan *tofnd.TrafficOut)
	resChan := make(chan []byte)

	// server handler https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc-1
	go func() {
		defer cancel()
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
			case *tofnd.MessageOut_KeygenResult:
				resChan <- msg.KeygenResult
				return
			case *tofnd.MessageOut_SignResult:
				resChan <- msg.SignResult
				return
			default:
				errChan <- fmt.Errorf("handler goroutine: server stream should send only msg type")
				return
			}
		}
	}()
	return broadcastChan, resChan
}

func parseMsgParams(attributes []sdk.Attribute) (sessionID string, from string, payload *tofnd.TrafficOut) {
	for _, attribute := range attributes {
		switch attribute.Key {
		case tss.AttributeKeySessionID:
			sessionID = attribute.Value
		case sdk.AttributeKeySender:
			from = attribute.Value
		case tss.AttributeKeyPayload:

			codec.Cdc.MustUnmarshalJSON([]byte(attribute.Value), &payload)
		default:
		}
	}

	return sessionID, from, payload
}

func prepareTrafficIn(myAddress string, from string, sessionID string, payload *tofnd.TrafficOut, logger log.Logger) (*tofnd.MessageIn, error) {
	if myAddress == from {
		return nil, nil
	}

	if !payload.IsBroadcast && myAddress != payload.ToPartyUid {
		return nil, nil
	}

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
		sessionID, from, payload.ToPartyUid, payload.IsBroadcast, myAddress))
	return msgIn, nil
}

func (mgr *Mgr) findMyIndex(participants []string) (int32, bool) {
	var myIndex int32 = -1
	for i, participant := range participants {
		if mgr.myAddress == participant {
			myIndex = int32(i)
			break
		}
	}
	// not participating
	if myIndex == -1 {
		return -1, false
	}

	return myIndex, true
}
