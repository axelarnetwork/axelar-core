package broadcaster

import (
	"fmt"
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
	logger    log.Logger
	pipeline  types.Pipeline
	txFactory tx.Factory
}

// NewBroadcaster returns a broadcaster to submit transactions to the blockchain.
// Only one instance of a broadcaster should be run for a given account, otherwise risk conflicting sequence numbers for submitted transactions.
func NewBroadcaster(txf tx.Factory, pipeline types.Pipeline, logger log.Logger) *Broadcaster {
	return &Broadcaster{
		logger:    logger,
		pipeline:  pipeline,
		txFactory: txf,
	}
}

// Broadcast sends the passed messages to the network. This function in thread-safe.
func (b *Broadcaster) Broadcast(ctx sdkClient.Context, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	var response *sdk.TxResponse
	// serialize concurrent calls to broadcast
	err := b.pipeline.Push(func() error {

		txf, err := prepareFactory(ctx, b.txFactory)
		if err != nil {
			return err
		}

		response, err = Broadcast(ctx, txf, msgs)
		if err != nil {
			// reset account and sequence number in case they were the issue
			b.txFactory = b.txFactory.
				WithAccountNumber(0).
				WithSequence(0)
			return err
		}

		b.logger.Debug(fmt.Sprintf("tx response with hash [%s] and opcode [%d]: %s",
			response.TxHash, response.Code, response.RawLog))

		// broadcast has been successful, so increment sequence number
		b.txFactory = txf.WithSequence(txf.Sequence() + 1)

		return nil
	}, func(err error) bool {
		if !utils.IsABCIError(err) {
			return true
		}

		if sdkerrors.ErrWrongSequence.Is(err) || sdkerrors.ErrOutOfGas.Is(err) {
			return true
		}

		return false
	})
	return response, err
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

// Push adds the given function to the serialized execution pipeline
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
				p.logger.Info("successful broadcast after backoff")
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

// Close closes the pipeline
func (p RetryPipeline) Close() {
	close(p.c)
}

// NewPipelineWithRetry returns a pipeline with the given configuration
func NewPipelineWithRetry(cap int, maxRetries int, backOffStrategy utils.BackOff, logger log.Logger) *RetryPipeline {
	p := &RetryPipeline{
		c:          make(chan func(), cap),
		backOff:    backOffStrategy,
		maxRetries: maxRetries,
		logger:     logger,
	}

	go func() {
		for f := range p.c {
			f()
		}
	}()

	return p
}
