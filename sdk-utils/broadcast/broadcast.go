package broadcast

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/crypto/tmhash"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	errors2 "github.com/axelarnetwork/axelar-core/utils/errors"
	auxiliarytypes "github.com/axelarnetwork/axelar-core/x/auxiliary/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/utils"
	"github.com/axelarnetwork/utils/log"
	"github.com/axelarnetwork/utils/slices"
)

//go:generate moq -pkg mock -out mock/broadcast.go . Broadcaster

// PrepareTx returns a marshalled tx that can be broadcast to the blockchain
func PrepareTx(ctx sdkClient.Context, txf tx.Factory, msgs ...sdk.Msg) ([]byte, error) {
	if len(msgs) == 0 {
		return nil, errors.New("call broadcast with at least one message")
	}

	// By convention the first signer of a tx pays the fees
	signers, _, err := ctx.Codec.GetMsgV1Signers(msgs[0])
	if err != nil {
		return nil, errorsmod.Wrapf(err, "failed to get signers from message %T", msgs[0])
	}
	if len(signers) == 0 {
		return nil, fmt.Errorf("messages must have at least one signer")
	}

	if txf.SimulateAndExecute() || ctx.Simulate {
		_, adjusted, err := tx.CalculateGas(ctx, txf, msgs...)
		if isSequenceMismatch(err) {
			return nil, sdkerrors.ErrWrongSequence
		}
		if err != nil {

			return nil, err
		}

		txf = txf.WithGas(adjusted)
	}

	txBuilder, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	txBuilder.SetFeeGranter(ctx.GetFeeGranterAddress())
	err = tx.Sign(ctx.CmdContext, txf, ctx.GetFromName(), txBuilder, true)
	if err != nil {
		return nil, err
	}

	txBytes, err := ctx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}
	return txBytes, nil
}

func isSequenceMismatch(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), sdkerrors.ErrWrongSequence.Error())
}

// Broadcast sends the given tx to the blockchain and blocks until it is added to a block (or timeout).
func Broadcast(ctx sdkClient.Context, subscriber TxEventSubscriber, txBytes []byte, options ...BroadcasterOption) (*sdk.TxResponse, error) {
	params := broadcastParams{
		Timeout: config.DefaultRPCConfig().TimeoutBroadcastTxCommit,
	}
	for _, option := range options {
		params = option(params)
	}

	// Subscribe to the inclusion event before broadcasting so the event cannot
	// be missed. Confirmation is done purely via this subscription, so it works
	// with the tx indexer disabled (tx_index="null"). A failure here is loud on
	// purpose: we never silently fall back to the indexer.
	txHash := fmt.Sprintf("%X", tmhash.Sum(txBytes))
	subscription, err := subscribeTxEvent(subscriber, txHash, params.Timeout)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to subscribe to tx inclusion event")
	}
	defer subscription.close()

	res, err := ctx.BroadcastTx(txBytes)

	switch {
	case err != nil:
		return nil, err
	case res.Code != abci.CodeTypeOK:
		return nil, errorsmod.ABCIError(res.Codespace, res.Code, res.RawLog)
	}

	res, err = waitForBlockInclusion(ctx, res.TxHash, subscription)

	switch {
	case err != nil:
		return nil, err
	case res.Code != abci.CodeTypeOK:
		return nil, errorsmod.ABCIError(res.Codespace, res.Code, res.RawLog)
	}

	return res, nil
}

// TxEventSubscriber subscribes to and unsubscribes from CometBFT events. It is
// implemented by any started CometBFT RPC client (e.g. a started *rpchttp.HTTP)
// and by vald's tendermint.RobustClient. A *started* client is required: event
// subscriptions need a live websocket, which the default client context's
// client does not have. Unsubscribe is per-query (we never call UnsubscribeAll)
// so the subscriber can be safely shared with other subscriptions on the same
// client (e.g. vald's event bus) without tearing them down.
type TxEventSubscriber interface {
	Subscribe(ctx context.Context, subscriber, query string, outCapacity ...int) (<-chan coretypes.ResultEvent, error)
	Unsubscribe(ctx context.Context, subscriber, query string) error
}

type txSubscription struct {
	eventCh <-chan coretypes.ResultEvent
	ctx     context.Context
	cancel  context.CancelFunc
	cleanup func()
}

// close cancels the subscription context and unsubscribes from the event.
func (s *txSubscription) close() {
	s.cancel()
	s.cleanup()
}

// subscribeTxEvent subscribes to the inclusion event for the given tx hash via
// the (required, started) subscriber. Inclusion is confirmed purely through
// this subscription, so it works with the tx indexer disabled.
func subscribeTxEvent(subscriber TxEventSubscriber, txHash string, timeout time.Duration) (*txSubscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	// Unique subscriber name per tx, and Unsubscribe by query (never
	// UnsubscribeAll), so sharing the client with other subscriptions can never
	// tear those down, and a leaked subscription can't collide with the next tx.
	subscriberName := "broadcast-wait-" + txHash
	query := fmt.Sprintf("%s='%s' AND %s='%s'", tmtypes.EventTypeKey, tmtypes.EventTx, tmtypes.TxHashKey, txHash)
	eventCh, err := subscriber.Subscribe(ctx, subscriberName, query)
	if err != nil {
		cancel()
		return nil, err
	}

	return &txSubscription{
		eventCh: eventCh,
		ctx:     ctx,
		cancel:  cancel,
		cleanup: func() {
			// Fresh context for unsubscribe since the original may have timed out.
			_ = subscriber.Unsubscribe(context.Background(), subscriberName, query)
		},
	}, nil
}

func waitForBlockInclusion(clientCtx sdkClient.Context, txHash string, subscription *txSubscription) (*sdk.TxResponse, error) {
	// The subscription is established before the tx is broadcast, so the
	// inclusion event cannot be missed unless the connection drops. On a
	// drop/timeout the caller's retry wrapper re-broadcasts, which is safe
	// thanks to sequence-number tracking.
	select {
	case evt, ok := <-subscription.eventCh:
		if !ok {
			return nil, errors.New("subscription channel closed while waiting for tx to be included in a block")
		}

		txEvent, ok := evt.Data.(tmtypes.EventDataTx)
		if !ok {
			return nil, errors.New("unexpected event data type while waiting for tx inclusion")
		}

		return &sdk.TxResponse{
			TxHash:    txHash,
			Height:    txEvent.Height,
			Codespace: txEvent.Result.Codespace,
			Code:      txEvent.Result.Code,
			Data:      strings.ToUpper(hex.EncodeToString(txEvent.Result.Data)),
			RawLog:    txEvent.Result.Log,
			Info:      txEvent.Result.Info,
			GasWanted: txEvent.Result.GasWanted,
			GasUsed:   txEvent.Result.GasUsed,
			Events:    txEvent.Result.Events,
		}, nil

	case <-subscription.ctx.Done():
		// Missed the inclusion event (timeout or reconnect window). Do a single
		// best-effort indexer lookup, otherwise return a descriptive error.
		res, err := authtx.QueryTx(clientCtx, txHash)
		if err == nil {
			return res, nil
		}

		if strings.Contains(err.Error(), "indexing is disabled") {
			return nil, errors.New("timed out waiting for tx inclusion and transaction indexing is disabled")
		}

		return nil, errors.New("timed out waiting for tx to be included in a block")
	}
}

// Broadcaster broadcasts msgs to the blockchain
type Broadcaster interface {
	Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error)
}

type statefulBroadcaster struct {
	clientCtx  sdkClient.Context
	subscriber TxEventSubscriber
	txf        tx.Factory
	options    []BroadcasterOption
}

// WithStateManager tracks sequence numbers, so it can be used to broadcast consecutive txs.
// The subscriber (a started client, e.g. tendermint.RobustClient) is required: tx inclusion
// is confirmed via its event subscription, which works with the tx indexer disabled.
func WithStateManager(clientCtx sdkClient.Context, subscriber TxEventSubscriber, txf tx.Factory, options ...BroadcasterOption) Broadcaster {
	return &statefulBroadcaster{
		clientCtx:  clientCtx,
		subscriber: subscriber,
		txf:        txf,
		options:    options,
	}
}

// Broadcast broadcasts the given msgs to the blockchain, keeps track of the sender's sequence number
func (b *statefulBroadcaster) Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	if len(msgs) == 0 {
		return nil, fmt.Errorf("no messages to broadcast")
	}

	log.FromCtx(ctx).Debug("starting to broadcast message batch")

	var err error
	b.txf, err = prepareFactory(b.clientCtx, b.txf)
	if err != nil {
		return nil, err
	}

	bz, err := PrepareTx(b.clientCtx, b.txf, msgs...)
	if sdkerrors.ErrWrongSequence.Is(err) {
		b.txf = b.txf.
			WithAccountNumber(0).
			WithSequence(0)
	}
	if err != nil {
		return nil, errorsmod.Wrap(err, "tx preparation failed")
	}
	response, err := Broadcast(b.clientCtx, b.subscriber, bz, b.options...)
	if err != nil {
		return nil, errorsmod.Wrap(err, "broadcast failed")
	}

	ctx = log.Append(ctx, "hash", response.TxHash).
		Append("op_code", response.Code).
		Append("raw_log", response.RawLog)
	log.FromCtx(ctx).Debug("received tx response")

	// broadcast has been successful, so increment sequence number
	b.txf = b.txf.WithSequence(b.txf.Sequence() + 1)

	return response, nil
}

// prepareFactory ensures the account defined by ctx.GetFromAddress() exists and
// if the account number and/or the account sequence number are zero (not set),
// they will be queried for and set on the provided Factory. A new Factory with
// the updated fields will be returned.
func prepareFactory(clientCtx sdkClient.Context, txf tx.Factory) (tx.Factory, error) {
	from := clientCtx.GetFromAddress()

	if err := txf.AccountRetriever().EnsureExists(clientCtx, from); err != nil {
		return txf, err
	}

	initNum, initSeq := txf.AccountNumber(), txf.Sequence()
	if initNum == 0 || initSeq == 0 {
		num, seq, err := txf.AccountRetriever().GetAccountNumberSequence(clientCtx, from)
		if err != nil {
			return txf, err
		}

		if initNum == 0 {
			txf = txf.WithAccountNumber(num)
		}

		if initSeq == 0 {
			txf = txf.WithSequence(seq)
		}
	}

	return txf, nil
}

// BroadcasterOption modifies broadcaster behaviour
type BroadcasterOption func(broadcaster broadcastParams) broadcastParams

type broadcastParams struct {
	Timeout time.Duration
}

// WithResponseTimeout sets the time to wait for a tx response
func WithResponseTimeout(timeout time.Duration) BroadcasterOption {
	return func(params broadcastParams) broadcastParams {
		params.Timeout = timeout
		return params
	}
}

type pipelinedBroadcaster struct {
	retryPipeline *retryPipeline
	broadcaster   Broadcaster
}

// WithRetry returns a broadcaster that retries the broadcast up to the given number of times if the broadcast fails
func WithRetry(broadcaster Broadcaster, maxRetries int, minSleep time.Duration) Broadcaster {
	b := &pipelinedBroadcaster{
		broadcaster:   broadcaster,
		retryPipeline: newPipelineWithRetry(10000, maxRetries, utils.LinearBackOff(minSleep)),
	}

	return b
}

// Broadcast implements the Broadcaster interface
func (b *pipelinedBroadcaster) Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	var (
		response *sdk.TxResponse
		err      error
	)

	// need to be able to reorder msgs, so clone the msgs slice
	retryMsgs := append(make([]sdk.Msg, 0, len(msgs)), msgs...)
	err = b.retryPipeline.Push(ctx,
		func(ctx context.Context) error {
			response, err = b.broadcaster.Broadcast(ctx, retryMsgs...)
			return err
		},
		func(err error) bool {
			i, ok := tryParseErrorMsgIndex(err)
			if ok && len(retryMsgs) > 1 {
				log.FromCtx(ctx).Debug(fmt.Sprintf("excluding message at index %d due to error", i))
				retryMsgs = append(retryMsgs[:i], retryMsgs[i+1:]...)
				return true
			}

			if !errors2.Is[*errorsmod.Error](err) {
				return true
			}

			if sdkerrors.ErrWrongSequence.Is(err) || sdkerrors.ErrOutOfGas.Is(err) {
				return true
			}

			return false
		})

	return response, err
}

func tryParseErrorMsgIndex(err error) (int, bool) {
	split := strings.SplitAfter(err.Error(), "message index: ")
	if len(split) < 2 {
		return 0, false
	}

	index := strings.Split(split[1], ":")[0]

	i, err := strconv.Atoi(index)
	if err != nil {
		return 0, false
	}
	return i, true
}

type batchedBroadcaster struct {
	cdc            codec.Codec
	broadcaster    Broadcaster
	backlog        backlog
	batchThreshold int
	batchSizeLimit int
}

// Batched returns a broadcaster that batches msgs together if there is high traffic to increase throughput
func Batched(broadcaster Broadcaster, batchThreshold, batchSizeLimit int, cdc codec.Codec) Broadcaster {
	b := &batchedBroadcaster{
		broadcaster:    broadcaster,
		backlog:        backlog{tail: make(chan broadcastTask, 10000)},
		batchThreshold: batchThreshold,
		batchSizeLimit: batchSizeLimit,
		cdc:            cdc,
	}

	go b.processBacklog()
	return b
}

// Broadcast implements the Broadcaster interface
func (b *batchedBroadcaster) Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	ctx = log.Append(ctx, "process", "batched broadcast")

	// serialize concurrent calls to broadcast
	callback := make(chan broadcastResult, 1)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-b.backlog.Push(broadcastTask{ctx, msgs, callback}):
		ctx = log.Append(ctx, "msg_count", len(msgs))
		log.FromCtx(ctx).Debug("queuing up messages")
		break
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-callback:
		return res.Response, res.Err
	}
}

func (b *batchedBroadcaster) processBacklog() {
	for {
		// do not batch if there is no backlog pressure to minimize the risk of broadcast errors (and subsequent retries)
		if b.backlog.Len() < b.batchThreshold {
			task := b.backlog.Pop()

			ctx := log.Append(task.Ctx, "batch_size", len(task.Msgs))
			log.FromCtx(ctx).Debug("low traffic; no batch merging")
			response, err := b.broadcaster.Broadcast(ctx, task.Msgs...)
			task.Callback <- broadcastResult{
				Response: response,
				Err:      err,
			}
			continue
		}

		var (
			ctx       context.Context
			msgs      []sdk.Msg
			callbacks []chan<- broadcastResult
		)

		for {
			// we cannot split a single task, so take at least one task and then fill up the batch
			// until the size limit is reached
			batchWouldBeTooLarge := len(msgs) > 0 && len(msgs)+len(b.backlog.Peek().Msgs) > b.batchSizeLimit
			if batchWouldBeTooLarge {
				break
			}

			task := b.backlog.Pop()

			if task.Ctx.Err() != nil {
				log.FromCtx(task.Ctx).Debug("context expired, discarding msgs")
				continue
			}

			ctx = task.Ctx
			msgs = append(msgs, task.Msgs...)
			callbacks = append(callbacks, task.Callback)

			// if there are no new tasks in the backlog, stop filling up the batch
			if b.backlog.Len() == 0 {
				break
			}
		}

		ctx = log.Append(ctx, "batch_size", len(msgs))
		log.FromCtx(ctx).Debug("high traffic; merging batches")

		signers, _, err := b.cdc.GetMsgV1Signers(msgs[0])
		if err != nil {
			panic(err)
		}

		response, err := b.broadcaster.Broadcast(ctx, auxiliarytypes.NewBatchRequest(signers[0], msgs))

		for _, callback := range callbacks {
			callback <- broadcastResult{
				Response: response,
				Err:      err,
			}
		}

	}
}

type refundableBroadcaster struct {
	cdc         codec.Codec
	broadcaster Broadcaster
}

// Broadcast wraps all given msgs into RefundMsgRequest msgs before broadcasting them
func (b *refundableBroadcaster) Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	var refundables []sdk.Msg
	for _, msg := range msgs {

		msgV1Signers, _, err := b.cdc.GetMsgV1Signers(msgs[0])
		if err != nil {
			return nil, err
		}
		signers := slices.Map(msgV1Signers, func(s []byte) sdk.AccAddress { return s })

		if len(signers) > 0 {
			refundables = append(refundables, types.NewRefundMsgRequest(signers[0], msg))
		}
	}
	return b.broadcaster.Broadcast(ctx, refundables...)
}

// WithRefund wraps a broadcaster into a refundableBroadcaster
func WithRefund(b Broadcaster, cdc codec.Codec) Broadcaster {
	return &refundableBroadcaster{broadcaster: b, cdc: cdc}
}

type suppressorBroadcaster struct {
	b Broadcaster
}

// SuppressExecutionErrs logs errors when msg executions fail and then suppresses them
func SuppressExecutionErrs(broadcaster Broadcaster) Broadcaster {
	return suppressorBroadcaster{
		b: broadcaster,
	}
}

// Broadcast implements the Broadcaster interface
func (s suppressorBroadcaster) Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	res, err := s.b.Broadcast(ctx, msgs...)
	if errors2.Is[*errorsmod.Error](err) {
		log.FromCtx(ctx).Info(fmt.Sprintf("tx response with error: %s", err))
		return nil, nil
	}
	return res, err
}
