package keeper

import (
	"context"
	"fmt"
	"io"
	"math"
	"time"

	tssd "github.com/axelarnetwork/tssd/pb"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

const (
	rotationPrefix         = "rotationCount_"
	keygenStartHeight      = "blockHeight_"
	pkPrefix               = "pk_"
	snapshotForKeyIDPrefix = "sfkid_"
	sigPrefix              = "sig_"
	keyIDForSigPrefix      = "kidfs_"
)

type Keeper struct {
	broadcaster   types.Broadcaster
	snapshotter   types.Snapshotter
	client        tssd.GG18Client
	keygenStreams map[string]types.Stream
	signStreams   map[string]types.Stream
	params        params.Subspace
	storeKey      sdk.StoreKey
	cdc           *codec.Codec
	voter         types.Voter
}

// NewKeeper constructs a tss keeper
func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, client types.TSSDClient,
	paramSpace params.Subspace, v types.Voter, broadcaster types.Broadcaster, snapshotter types.Snapshotter) Keeper {
	return Keeper{
		broadcaster:   broadcaster,
		snapshotter:   snapshotter,
		client:        client,
		cdc:           cdc,
		keygenStreams: map[string]types.Stream{},
		signStreams:   map[string]types.Stream{},
		params:        paramSpace.WithKeyTable(types.KeyTable()),
		storeKey:      storeKey,
		voter:         v,
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
	k.params.SetParamSet(ctx, &set)
}

// GetParams gets the tss module's parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.params.GetParamSet(ctx, &params)
	return
}

// getLockingPeriod returns the period of blocks that keygen is locked after a new snapshot has been created
func (k Keeper) getLockingPeriod(ctx sdk.Context) int64 {
	var period int64
	k.params.Get(ctx, types.KeyLockingPeriod, &period)
	return period
}

func (k Keeper) prepareTrafficIn(ctx sdk.Context, sender sdk.AccAddress, sessionID string, payload *tssd.TrafficOut) (*tssd.MessageIn, error) {
	// deterministic error
	senderAddress := k.broadcaster.GetPrincipal(ctx, sender)
	if senderAddress.Empty() {
		err := fmt.Errorf("invalid message: sender [%s] is not a validator", sender)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}
	k.Logger(ctx).Debug(fmt.Sprintf("session [%.20s] from [%.20s] to [%.20s] broadcast? [%t]", sessionID, senderAddress.String(), payload.ToPartyUid, payload.IsBroadcast))

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
		return nil, nil
	}
	if payload.IsBroadcast && myAddress.Equals(senderAddress) {
		return nil, nil
	}

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
		"incoming msg to tssd: session [%.20s] from [%.20s] to [%.20s] broadcast [%t] me [%.20s]",
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
		partyUids = append(partyUids, v.GetOperator().String())
		if v.GetOperator().Equals(myAddress) {
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
		if k.broadcaster.GetProxy(ctx, v.GetOperator()) == nil {
			return fmt.Errorf("validator %s has not registered a proxy", v.GetOperator())
		}
	}
	return nil
}

// ComputeCorruptionThreshold returns corruption threshold to be used by tss
func (k Keeper) ComputeCorruptionThreshold(ctx sdk.Context, totalvalidators int) int {
	var threshold utils.Threshold
	k.params.Get(ctx, types.KeyCorruptionThreshold, &threshold)
	// threshold = totalValidators * corruption threshold - 1
	return int(math.Ceil(float64(totalvalidators)*float64(threshold.Numerator)/
		float64(threshold.Denominator))) - 1
}

// GetMinKeygenThreshold returns minimum threshold of stake that must be met to execute keygen
func (k Keeper) GetMinKeygenThreshold(ctx sdk.Context) utils.Threshold {
	var threshold utils.Threshold
	k.params.Get(ctx, types.KeyMinKeygenThreshold, &threshold)
	return threshold
}
