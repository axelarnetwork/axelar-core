package broadcast

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
)

// Broadcaster submits transactions to a tendermint node
type Broadcaster struct {
	logger    log.Logger
	queue     chan func()
	ctx       sdkClient.Context
	txFactory tx.Factory
}

// NewBroadcaster returns a broadcaster to submit transactions to the blockchain.
// Only one instance of a broadcaster should be run for a given account, otherwise risk conflicting sequence numbers for submitted transactions.
func NewBroadcaster(ctx sdkClient.Context, txf tx.Factory, logger log.Logger) *Broadcaster {
	broadcaster := &Broadcaster{
		ctx:       ctx,
		logger:    logger,
		queue:     make(chan func(), 1000),
		txFactory: txf,
	}

	// sequential function queue
	go func() {
		// this is expected to run for the full life time of the process, so there is no need to be able to escape the loop
		for executeCommand := range broadcaster.queue {
			executeCommand()
		}
	}()

	return broadcaster
}

// Reset resets the account number and sequence number of the sender. This is executed in serialized order with other incoming broadcast commands
func (b *Broadcaster) Reset() {
	b.queue <- func() {
		b.txFactory = b.txFactory.
			WithAccountNumber(0).
			WithSequence(0)
	}
}

// Broadcast sends the passed messages to the network. This function in thread-safe.
func (b *Broadcaster) Broadcast(msgs ...sdk.Msg) error {
	errChan := make(chan error, 1)
	// push the "intent to run broadcast" into the queue so it can be executed sequentially,
	// even if the public Broadcast function is called concurrently
	b.queue <- func() { errChan <- b.broadcast(msgs) }
	// block until the broadcast call has actually been run
	return <-errChan
}

func (b *Broadcaster) broadcast(msgs []sdk.Msg) error {
	if len(msgs) == 0 {
		return fmt.Errorf("call broadcast with at least one message")
	}

	// By convention the first signer of a tx pays the fees
	if len(msgs[0].GetSigners()) == 0 {
		return fmt.Errorf("messages must have at least one signer")
	}

	txf, err := tx.PrepareFactory(b.ctx, b.txFactory)
	if err != nil {
		return err
	}

	if txf.SimulateAndExecute() || b.ctx.Simulate {
		_, adjusted, err := tx.CalculateGas(b.ctx.QueryWithData, txf, msgs...)
		if err != nil {
			return err
		}

		txf = txf.WithGas(adjusted)
	}

	txBuilder, err := tx.BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return err
	}

	err = tx.Sign(txf, b.ctx.GetFromName(), txBuilder, true)
	if err != nil {
		return err
	}

	txBytes, err := b.ctx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := b.ctx.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	if res.Code != abci.CodeTypeOK {
		return fmt.Errorf(res.RawLog)
	}
	// broadcast has been successful, so increment sequence number
	b.txFactory = b.txFactory.WithSequence(txf.Sequence() + 1)

	return nil
}

// BackOffBroadcaster is a broadcast wrapper that adds retries with backoff
type BackOffBroadcaster struct {
	broadcaster *Broadcaster
	timeout     time.Duration
	maxRetries  int
	backOff     func(retryCount int, minTimeout time.Duration) time.Duration
	queue       chan func()
}

// WithBackoff wraps a broadcaster so that failed queue are retried with the given back-off strategy
func WithBackoff(b *Broadcaster, strategy BackOff, minTimeout time.Duration, maxRetries int) *BackOffBroadcaster {
	backOff := &BackOffBroadcaster{
		broadcaster: b,
		timeout:     minTimeout,
		maxRetries:  maxRetries,
		backOff:     strategy,
		queue:       make(chan func(), 1000),
	}

	// call broadcast functions sequentially
	go func() {
		// this is expected to run for the full life time of the process, so there is no need to be able to escape the loop
		for broadcast := range backOff.queue {
			broadcast()
		}
	}()

	return backOff
}

// Broadcast submits messages synchronously and retries with exponential backoff.
// This function is thread-safe but might block for a long time depending on the exponential backoff parameters.
func (b *BackOffBroadcaster) Broadcast(msgs ...sdk.Msg) error {
	errChan := make(chan error, 1)
	// push the "intent to run broadcast" into a channel so it can be executed sequentially,
	// even if the public Broadcast function is called concurrently
	b.queue <- func() { errChan <- b.broadcastWithBackoff(msgs) }
	// block until the broadcast call has actually been run
	return <-errChan
}

func (b *BackOffBroadcaster) broadcastWithBackoff(msgs []sdk.Msg) (err error) {
	for i := 0; i <= b.maxRetries; i++ {
		err = b.broadcaster.Broadcast(msgs...)
		if err == nil {
			return nil
		}

		// in case the issue is sequence number, we reset it here
		b.broadcaster.Reset()

		// exponential backoff
		if i < b.maxRetries {
			timeout := b.backOff(i, b.timeout)
			b.broadcaster.logger.Error(sdkerrors.Wrapf(err, "backing off (retry in %v )", timeout).Error())
			time.Sleep(timeout)
		}
	}

	return sdkerrors.Wrap(err, fmt.Sprintf("aborting broadcast after %d retries", b.maxRetries))
}

// BackOff computes the next back-off duration
type BackOff func(retryCount int, minTimeout time.Duration) time.Duration

var (
	// Exponential computes an exponential back-off
	Exponential = func(i int, minTimeout time.Duration) time.Duration {
		jitter := rand.Float64()
		strategy := math.Pow(2, float64(i))
		// casting a float <1 to time.Duration ==0, so need normalize by multiplying with float64(time.Second) first
		backoff := math.Max(strategy*jitter, 1) * minTimeout.Seconds() * float64(time.Second)

		return time.Duration(backoff)
	}

	// Linear computes a linear back-off
	Linear = func(i int, minTimeout time.Duration) time.Duration {
		jitter := rand.Float64()
		strategy := float64(i)

		// casting a float <1 to time.Duration ==0, so need normalize by multiplying with float64(time.Second) first
		backoff := math.Max(strategy*jitter, 1) * minTimeout.Seconds() * float64(time.Second)
		return time.Duration(backoff)
	}
)
