package broadcaster

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types"
	"github.com/axelarnetwork/axelar-core/utils"
)

// Broadcaster submits transactions to a tendermint node
type Broadcaster struct {
	logger         log.Logger
	retryPipeline  types.Pipeline
	backlog        backlog
	txFactory      tx.Factory
	clientCtx      sdkClient.Context
	batchThreshold int
	batchSizeLimit int
}

// NewBroadcaster returns a broadcaster to submit transactions to the blockchain.
// Only one instance of a broadcaster should be run for a given account, otherwise risk conflicting sequence numbers for submitted transactions.
func NewBroadcaster(txf tx.Factory, clientCtx sdkClient.Context, pipeline types.Pipeline, batchThreshold, batchSizeLimit int, logger log.Logger) *Broadcaster {
	b := &Broadcaster{
		logger:         logger.With("process", "broadcast"),
		retryPipeline:  pipeline,
		txFactory:      txf,
		clientCtx:      clientCtx,
		batchThreshold: batchThreshold,
		batchSizeLimit: batchSizeLimit,
		backlog:        backlog{tail: make(chan broadcastTask, 10000)},
	}

	go b.processBacklog()
	return b
}

func (b *Broadcaster) processBacklog() {
	for {
		// do not batch if there is no backlog pressure to minimize the risk of broadcast errors (and subsequent retries)
		if b.backlog.Len() < b.batchThreshold {
			task := b.backlog.Pop()

			b.logger.Debug("low traffic; no batch merging", "batch_size", len(task.Msgs))
			b.broadcastWithRetry(task)
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
		b.broadcastWithRetry(batch...)
	}
}

// Broadcast queues up the given messages for broadcast to the network. This function in thread-safe, blocks until it gets a response.
func (b *Broadcaster) Broadcast(ctx context.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	if len(msgs) == 0 {
		return nil, fmt.Errorf("no messages to broadcast")
	}

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

func (b *Broadcaster) broadcastWithRetry(tasks ...broadcastTask) {
	var response *sdk.TxResponse

	var msgs []sdk.Msg
	for _, task := range tasks {
		if task.Ctx.Err() != nil {
			continue
		}
		msgs = append(msgs, task.Msgs...)
	}

	err := b.retryPipeline.Push(
		func() error {
			logger := b.logger.With("batch_size", len(msgs))
			logger.Debug("starting to broadcast message batch")
			txf, err := prepareFactory(b.clientCtx, b.txFactory)
			if err != nil {
				return err
			}

			response, err = Broadcast(b.clientCtx, txf, msgs)
			if err != nil {
				// reset account and sequence number in case they were the issue
				b.txFactory = b.txFactory.
					WithAccountNumber(0).
					WithSequence(0)
				return err
			}

			logger.Debug("received tx response",
				"hash", response.TxHash,
				"op_code", response.Code,
				"raw_log", response.RawLog,
			)

			// broadcast has been successful, so increment sequence number
			b.txFactory = txf.WithSequence(txf.Sequence() + 1)

			return nil
		},
		func(err error) bool {
			logger := b.logger.With("batch_size", len(msgs))
			i, ok := tryParseErrorMsgIndex(err)
			if ok && len(msgs) > 1 {
				logger.Debug(fmt.Sprintf("excluding message at index %d due to error", i))
				msgs = append(msgs[:i], msgs[i+1:]...)
				return true
			}

			if !utils.IsABCIError(err) {
				return true
			}

			if sdkerrors.ErrWrongSequence.Is(err) || sdkerrors.ErrOutOfGas.Is(err) {
				return true
			}

			return false
		})

	for _, task := range tasks {
		task.Callback <- broadcastResult{
			Response: response,
			Err:      err,
		}
	}
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

// Broadcast bundles the given messages into a single transaction and submits it to the blockchain.
// If there are more than one message, all messages must have the single same signer
func Broadcast(ctx sdkClient.Context, txf tx.Factory, msgs []sdk.Msg) (*sdk.TxResponse, error) {
	if len(msgs) == 0 {
		return nil, fmt.Errorf("call broadcast with at least one message")
	}

	// By convention the first signer of a tx pays the fees
	if len(msgs[0].GetSigners()) == 0 {
		return nil, fmt.Errorf("messages must have at least one signer")
	}

	if txf.SimulateAndExecute() || ctx.Simulate {
		_, adjusted, err := tx.CalculateGas(ctx, txf, msgs...)
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

	// broadcast to a Tendermint node
	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	if res.Code != abci.CodeTypeOK {
		return nil, sdkerrors.ABCIError(res.Codespace, res.Code, res.RawLog)
	}

	return res, nil
}

// RetryPipeline manages serialized execution of functions with retry on error
type RetryPipeline struct {
	c          chan func()
	backOff    utils.BackOff
	maxRetries int
	logger     log.Logger
}

// Push adds the given function to the serialized execution retryPipeline
func (p RetryPipeline) Push(f func() error, retryOnError func(error) bool) error {
	e := make(chan error, 1)
	p.c <- func() { e <- p.retry(f, retryOnError) }
	return <-e
}

func (p RetryPipeline) retry(f func() error, retryOnError func(error) bool) error {
	var err error
	for i := 0; i <= p.maxRetries; i++ {
		err = f()
		if err == nil {
			if i > 0 {
				p.logger.Info("successful execution after backoff")
			}
			return nil
		}

		if !retryOnError(err) {
			p.logger.Error(fmt.Sprintf("tx response with error: %s", err))
			return nil
		}

		if i < p.maxRetries {
			timeout := p.backOff(i)
			p.logger.Info(sdkerrors.Wrapf(err, "backing off (retry in %v )", timeout).Error())
			time.Sleep(timeout)
		}
	}
	return sdkerrors.Wrap(err, fmt.Sprintf("aborting after %d retries", p.maxRetries))
}

// Close closes the retryPipeline
func (p RetryPipeline) Close() {
	close(p.c)
}

// NewPipelineWithRetry returns a retryPipeline with the given configuration
func NewPipelineWithRetry(cap int, maxRetries int, backOffStrategy utils.BackOff, logger log.Logger) *RetryPipeline {
	p := &RetryPipeline{
		c:          make(chan func(), cap),
		backOff:    backOffStrategy,
		maxRetries: maxRetries,
		logger:     logger.With("process", "retry pipeline"),
	}

	go func() {
		for f := range p.c {
			f()
		}
	}()

	return p
}

type backlog struct {
	tail chan broadcastTask
	head broadcastTask
}

func (bl *backlog) Pop() broadcastTask {
	bl.loadHead()

	next := bl.head
	bl.head = broadcastTask{}
	return next
}

func (bl *backlog) loadHead() {
	if len(bl.head.Msgs) == 0 {
		bl.head = <-bl.tail
	}
}

func (bl *backlog) Push(task broadcastTask) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		if len(task.Msgs) == 0 {
			return
		}
		bl.tail <- task
	}()

	return done
}

func (bl *backlog) Peek() broadcastTask {
	bl.loadHead()

	return bl.head
}

func (bl *backlog) Len() int {
	// do not block in this function because it might be used to inform other calls like Peek()
	if len(bl.head.Msgs) == 0 {
		// head is not currently loaded
		return len(bl.tail)
	}

	return 1 + len(bl.tail)
}

type broadcastResult struct {
	Response *sdk.TxResponse
	Err      error
}

type broadcastTask struct {
	Ctx      context.Context
	Msgs     []sdk.Msg
	Callback chan<- broadcastResult
}
