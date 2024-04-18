package broadcast

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/config"

	errors2 "github.com/axelarnetwork/axelar-core/utils/errors"
	auxiliarytypes "github.com/axelarnetwork/axelar-core/x/auxiliary/types"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/utils"
	"github.com/axelarnetwork/utils/log"
)

//go:generate moq -pkg mock -out mock/broadcast.go . Broadcaster

// PrepareTx returns a marshalled tx that can be broadcast to the blockchain
func PrepareTx(ctx sdkClient.Context, txf tx.Factory, msgs ...sdk.Msg) ([]byte, error) {
	if len(msgs) == 0 {
		return nil, fmt.Errorf("call broadcast with at least one message")
	}

	// By convention the first signer of a tx pays the fees
	if len(msgs[0].GetSigners()) == 0 {
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

	txBuilder, err := tx.BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return nil, err
	}

	txBuilder.SetFeeGranter(ctx.GetFeeGranterAddress())
	err = tx.Sign(txf, ctx.GetFromName(), txBuilder, true)
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
func Broadcast(ctx sdkClient.Context, txBytes []byte, options ...BroadcasterOption) (*sdk.TxResponse, error) {
	res, err := ctx.BroadcastTx(txBytes)
	switch {
	case err != nil:
		return nil, err
	case res.Code != abci.CodeTypeOK:
		return nil, sdkerrors.ABCIError(res.Codespace, res.Code, res.RawLog)
	case ctx.BroadcastMode != flags.BroadcastBlock:
		params := broadcastParams{
			Timeout:         config.DefaultRPCConfig().TimeoutBroadcastTxCommit,
			PollingInterval: 2 * time.Second,
		}
		for _, option := range options {
			params = option(params)
		}

		res, err = waitForBlockInclusion(ctx, res.TxHash, params)
	}

	switch {
	case err != nil:
		return nil, err
	case res.Code != abci.CodeTypeOK:
		return nil, sdkerrors.ABCIError(res.Codespace, res.Code, res.RawLog)
	}

	return res, nil
}

func waitForBlockInclusion(clientCtx sdkClient.Context, txHash string, options broadcastParams) (*sdk.TxResponse, error) {
	timeout, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	ticker := time.NewTicker(options.PollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			res, err := authtx.QueryTx(clientCtx, txHash)
			if err != nil {
				// query failed or tx is not found yet
				continue
			}

			return res, nil
		case <-timeout.Done():
			// try one last time to find the tx
			res, err := authtx.QueryTx(clientCtx, txHash)
			if err != nil {
				return nil, errors.New("timed out waiting for tx to be included in a block")
			}
			return res, err
		}
	}
}

// Broadcaster broadcasts msgs to the blockchain
type Broadcaster interface {
	Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error)
}

type statefulBroadcaster struct {
	clientCtx sdkClient.Context
	txf       tx.Factory
	options   []BroadcasterOption
}

// WithStateManager tracks sequence numbers, so it can be used to broadcast consecutive txs
func WithStateManager(clientCtx sdkClient.Context, txf tx.Factory, options ...BroadcasterOption) Broadcaster {
	return &statefulBroadcaster{
		clientCtx: clientCtx,
		txf:       txf,
		options:   options,
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
		return nil, sdkerrors.Wrap(err, "tx preparation failed")
	}
	response, err := Broadcast(b.clientCtx, bz, b.options...)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "broadcast failed")
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
	PollingInterval time.Duration
	Timeout         time.Duration
}

// WithResponseTimeout sets the time to wait for a tx response
func WithResponseTimeout(timeout time.Duration) BroadcasterOption {
	return func(params broadcastParams) broadcastParams {
		params.Timeout = timeout
		return params
	}
}

// WithPollingInterval modifies how often the broadcaster checks the blockchain for tx responses
func WithPollingInterval(interval time.Duration) BroadcasterOption {
	return func(params broadcastParams) broadcastParams {
		params.PollingInterval = interval
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

			if !errors2.Is[*sdkerrors.Error](err) {
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
	broadcaster    Broadcaster
	backlog        backlog
	batchThreshold int
	batchSizeLimit int
}

// Batched returns a broadcaster that batches msgs together if there is high traffic to increase throughput
func Batched(broadcaster Broadcaster, batchThreshold, batchSizeLimit int) Broadcaster {
	b := &batchedBroadcaster{
		broadcaster:    broadcaster,
		backlog:        backlog{tail: make(chan broadcastTask, 10000)},
		batchThreshold: batchThreshold,
		batchSizeLimit: batchSizeLimit,
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

		response, err := b.broadcaster.Broadcast(ctx, auxiliarytypes.NewBatchRequest(msgs[0].GetSigners()[0], msgs))

		for _, callback := range callbacks {
			callback <- broadcastResult{
				Response: response,
				Err:      err,
			}
		}

	}
}

type refundableBroadcaster struct {
	broadcaster Broadcaster
}

// Broadcast wraps all given msgs into RefundMsgRequest msgs before broadcasting them
func (b *refundableBroadcaster) Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	var refundables []sdk.Msg
	for _, msg := range msgs {
		if len(msg.GetSigners()) > 0 {
			refundables = append(refundables, types.NewRefundMsgRequest(msg.GetSigners()[0], msg))
		}
	}
	return b.broadcaster.Broadcast(ctx, refundables...)
}

// WithRefund wraps a broadcaster into a refundableBroadcaster
func WithRefund(b Broadcaster) Broadcaster {
	return &refundableBroadcaster{broadcaster: b}
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
	if errors2.Is[*sdkerrors.Error](err) {
		log.FromCtx(ctx).Info(fmt.Sprintf("tx response with error: %s", err))
		return nil, nil
	}
	return res, err
}
