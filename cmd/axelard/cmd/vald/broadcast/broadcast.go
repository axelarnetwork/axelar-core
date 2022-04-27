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
	"github.com/tendermint/tendermint/libs/log"

	utils2 "github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/utils"
)

//go:generate moq -pkg mock -out mock/broadcast.go . Broadcaster

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

func Broadcast(ctx sdkClient.Context, txBytes []byte, options ...BroadcastOption) (*sdk.TxResponse, error) {
	res, err := ctx.BroadcastTx(txBytes)
	if err == nil && ctx.BroadcastMode != flags.BroadcastBlock {
		params := broadcastParams{
			Timeout:         config.DefaultRPCConfig().TimeoutBroadcastTxCommit,
			PollingInterval: 2 * time.Second,
		}
		for _, option := range options {
			params = option(params)
		}

		res, err = waitForBlockInclusion(ctx, res.TxHash, params)
	}
	if err != nil {
		return nil, err
	}

	if res.Code != abci.CodeTypeOK {
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
			return nil, errors.New("timed out waiting for tx to be included in a block")
		}
	}
}

type Broadcaster interface {
	Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error)
}

type statefulBroadcaster struct {
	clientCtx sdkClient.Context
	txf       tx.Factory
	options   []BroadcastOption
	logger    log.Logger
}

func WithStateManager(clientCtx sdkClient.Context, txf tx.Factory, logger log.Logger, options ...BroadcastOption) Broadcaster {
	return &statefulBroadcaster{
		clientCtx: clientCtx,
		txf:       txf,
		logger:    logger,
		options:   options,
	}
}

func (b *statefulBroadcaster) Broadcast(_ context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	if len(msgs) == 0 {
		return nil, fmt.Errorf("no messages to broadcast")
	}

	logger := b.logger.With("batch_size", len(msgs))
	logger.Debug("starting to broadcast message batch")

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

	b.logger.Debug("received tx response",
		"hash", response.TxHash,
		"op_code", response.Code,
		"raw_log", response.RawLog,
	)

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

type BroadcastOption func(broadcaster broadcastParams) broadcastParams

type broadcastParams struct {
	PollingInterval time.Duration
	Timeout         time.Duration
}

func WithResponseTimeout(timeout time.Duration) BroadcastOption {
	return func(params broadcastParams) broadcastParams {
		params.Timeout = timeout
		return params
	}
}

func WithPollingInterval(interval time.Duration) BroadcastOption {
	return func(params broadcastParams) broadcastParams {
		params.PollingInterval = interval
		return params
	}
}

type pipelinedBroadcaster struct {
	logger        log.Logger
	retryPipeline *retryPipeline
	txFactory     tx.Factory
	clientCtx     sdkClient.Context
	broadcaster   Broadcaster
}

func WithRetry(broadcaster Broadcaster, maxRetries int, minSleep time.Duration, logger log.Logger) Broadcaster {
	b := &pipelinedBroadcaster{
		broadcaster:   broadcaster,
		retryPipeline: newPipelineWithRetry(10000, maxRetries, utils.LinearBackOff(minSleep), logger),
	}

	return b
}

func (b *pipelinedBroadcaster) Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	var (
		response *sdk.TxResponse
		err      error
	)
	err = b.retryPipeline.Push(
		func() error {
			response, err = b.broadcaster.Broadcast(ctx, msgs...)
			return err
		},
		func(err error) bool {
			logger := b.logger.With("batch_size", len(msgs))
			i, ok := tryParseErrorMsgIndex(err)
			if ok && len(msgs) > 1 {
				logger.Debug(fmt.Sprintf("excluding message at index %d due to error", i))
				msgs = append(msgs[:i], msgs[i+1:]...)
				return true
			}

			if !utils2.IsABCIError(err) {
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
	split := strings.SplitAfter(err.Error(), "failed to execute message; message index: ")
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
	logger         log.Logger
}

func InBatches(broadcaster Broadcaster, batchThreshold, batchSizeLimit int, logger log.Logger) Broadcaster {
	b := &batchedBroadcaster{
		broadcaster:    broadcaster,
		backlog:        backlog{tail: make(chan broadcastTask, 10000)},
		batchThreshold: batchThreshold,
		batchSizeLimit: batchSizeLimit,
		logger:         logger.With("process", "batched broadcast"),
	}

	go b.processBacklog()
	return b
}

func (b *batchedBroadcaster) Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	// serialize concurrent calls to broadcast
	callback := make(chan broadcastResult, 1)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-b.backlog.Push(broadcastTask{ctx, msgs, callback}):
		b.logger.Debug("queuing up messages", "msg_count", len(msgs))
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

			b.logger.Debug("low traffic; no batch merging", "batch_size", len(task.Msgs))
			b.broadcast(task)
			continue
		}

		var batch []broadcastTask
		msgCount := 0
		for {
			// we cannot split a single task, so take at least one task and then fill up the batch
			// until the size limit is reached
			batchWouldBeTooLarge := len(batch) > 0 && msgCount+len(b.backlog.Peek().Msgs) > b.batchSizeLimit
			if batchWouldBeTooLarge {
				break
			}

			task := b.backlog.Pop()

			batch = append(batch, task)
			msgCount += len(task.Msgs)

			// if there are no new tasks in the backlog, stop filling up the batch
			if b.backlog.Len() == 0 {
				break
			}
		}

		b.logger.Debug("high traffic; merging batches", "batch_size", msgCount)
		b.broadcast(batch...)
	}
}

func (b *batchedBroadcaster) broadcast(batch ...broadcastTask) {
	var msgs []sdk.Msg
	for _, task := range batch {
		if task.Ctx.Err() != nil {
			b.logger.Debug("context expired, discarding msgs")
			continue
		}
		msgs = append(msgs, task.Msgs...)
	}

	response, err := b.broadcaster.Broadcast(context.Background(), msgs...)

	for _, task := range batch {
		task.Callback <- broadcastResult{
			Response: response,
			Err:      err,
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
