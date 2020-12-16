package keeper

import (
	"context"
	"fmt"
	"io"
	"time"

	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"

	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

type Keeper struct {
	broadcaster   types.Broadcaster
	client        tssd.GG18Client
	keygenStreams map[string]types.Stream
	signStreams   map[string]types.Stream
	paramSpace    params.Subspace
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
}

// NewKeeper constructs a tss keeper
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, client types.TSSDClient, paramSpace params.Subspace, broadcaster types.Broadcaster) Keeper {
	return Keeper{
		broadcaster:   broadcaster,
		client:        client,
		cdc:           cdc,
		keygenStreams: map[string]types.Stream{},
		signStreams:   map[string]types.Stream{},
		paramSpace:    paramSpace.WithKeyTable(types.KeyTable()),
		storeKey:      storeKey,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// need to create a new context for every new protocol start
func (k Keeper) newGrpcContext() (context.Context, context.CancelFunc) {
	// TODO: make timeout a config parameter?
	return context.WithTimeout(context.Background(), 2*time.Hour)
}

// SetParams sets the tss module's parameters
func (k Keeper) SetParams(ctx sdk.Context, set types.Params) {
	k.paramSpace.SetParamSet(ctx, &set)
}

// SetParams gets the tss module's parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return
}

func (k Keeper) prepareTrafficIn(ctx sdk.Context, sender sdk.AccAddress, sessionID string, payload *tssd.TrafficOut) (*tssd.MessageIn, error) {
	// deterministic error
	senderAddress := k.broadcaster.GetPrincipal(ctx, sender)
	if senderAddress.Empty() {
		err := fmt.Errorf("invalid message: sender [%s] is not a validator", sender)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}
	k.Logger(ctx).Debug(fmt.Sprintf("session [%s] from [%s] to [%s] broadcast? [%t]", sessionID, senderAddress.String(), payload.ToPartyUid, payload.IsBroadcast))

	// non-deterministic errors must not change behaviour, therefore log error and return nil instead
	myAddress := k.broadcaster.GetLocalPrincipal(ctx)
	if myAddress.Empty() {
		k.Logger(ctx).Info(fmt.Sprintf("ignore message: my validator address is empty so I must not be a validator"))
		return nil, nil
	}
	toAddress, err := sdk.ValAddressFromBech32(payload.ToPartyUid)
	if err != nil {
		k.Logger(ctx).Error(sdkerrors.Wrap(err, fmt.Sprintf("failed to parse [%s] into a validator address",
			payload.ToPartyUid)).Error())
		return nil, nil
	}
	if toAddress.String() != payload.ToPartyUid {
		k.Logger(ctx).Error("address parse discrepancy: given [%s] got [%s]", payload.ToPartyUid, toAddress.String())
	}
	if !payload.IsBroadcast && !myAddress.Equals(toAddress) {
		k.Logger(ctx).Info(fmt.Sprintf("I should ignore: msg to [%s] not directed to me [%s]", toAddress, myAddress))
		return nil, nil
	}
	if payload.IsBroadcast && myAddress.Equals(senderAddress) {
		k.Logger(ctx).Info(fmt.Sprintf("I should ignore: broadcast msg from [%s] came from me [%s]", senderAddress, myAddress))
		return nil, nil
	}
	k.Logger(ctx).Info(fmt.Sprintf("I should NOT ignore: msg from [%s] to [%s] broadcast [%t] me [%s]", senderAddress, toAddress, payload.IsBroadcast, myAddress))

	msgIn := &tssd.MessageIn{
		Data: &tssd.MessageIn_Traffic{
			Traffic: &tssd.TrafficIn{
				Payload:      payload.Payload,
				IsBroadcast:  payload.IsBroadcast,
				FromPartyUid: senderAddress.String(),
			},
		},
	}

	k.Logger(ctx).Debug(fmt.Sprintf(
		"forward incoming msg to tssd: session [%s] from [%s] to [%s] broadcast [%t] me [%s]",
		sessionID,
		senderAddress.String(),
		toAddress.String(),
		payload.IsBroadcast,
		myAddress.String(),
	))
	return msgIn, nil
}

func (k Keeper) handleStream(ctx sdk.Context, s types.Stream) (broadcast <-chan *tssd.TrafficOut, result <-chan []byte) {
	broadcastChan := make(chan *tssd.TrafficOut)
	resChan := make(chan []byte)

	// server handler https://grpc.io/docs/languages/go/basics/#bidirectional-streaming-rpc-1
	go func() {
		defer close(broadcastChan)
		defer close(resChan)
		defer func() {
			// close the stream on error or protocol completion
			if err := s.CloseSend(); err != nil {
				k.Logger(ctx).Error(sdkerrors.Wrap(err, "handler goroutine: failure to CloseSend stream").Error())
			}
		}()

		for {
			msgOneof, err := s.Recv() // blocking
			if err == io.EOF {        // output stream closed by server
				k.Logger(ctx).Debug("handler goroutine: gRPC stream closed by server")
				return
			}
			if err != nil {
				newErr := sdkerrors.Wrap(err, "handler goroutine: failure to receive msg from gRPC server stream")
				k.Logger(ctx).Error(newErr.Error())
				return
			}

			switch msg := msgOneof.GetData().(type) {
			case *tssd.MessageOut_Traffic:
				broadcastChan <- msg.Traffic
			case *tssd.MessageOut_KeygenResult:
				resChan <- msg.KeygenResult
				return
			case *tssd.MessageOut_SignResult:
				resChan <- msg.SignResult
				return
			default:
				newErr := sdkerrors.Wrap(types.ErrTss, "handler goroutine: server stream should send only msg type")
				k.Logger(ctx).Error(newErr.Error())
				return
			}
		}
	}()
	return broadcastChan, resChan
}

// addrToUids returns an error if myAddr is not part of the validator slice
func addrToUids(validators []snapshot.Validator, myAddress sdk.ValAddress) (partyIDs []string, myIndex int32, err error) {
	// populate a []tss.Party with all validator addresses
	partyUids := make([]string, 0, len(validators))
	alreadySeen, myIndex := false, 0
	for i, v := range validators {
		partyUids = append(partyUids, v.Address.String())
		if v.Address.Equals(myAddress) {
			if alreadySeen {
				return nil, 0, fmt.Errorf("cosmos bug: my validator address appears multiple times in the validator list: [%s]", myAddress)
			}
			alreadySeen, myIndex = true, int32(i)
		}
	}

	if !alreadySeen {
		return nil, 0, fmt.Errorf("broadcaster module bug: my validator address is not in the validator list: [%s]", myAddress)
	}

	return partyUids, myIndex, nil
}

func (k Keeper) checkProxies(ctx sdk.Context, validators []snapshot.Validator) error {
	for _, v := range validators {
		if k.broadcaster.GetProxy(ctx, v.Address) == nil {
			return fmt.Errorf("validator %s has not registered a proxy", v.Address)
		}
	}
	return nil
}
