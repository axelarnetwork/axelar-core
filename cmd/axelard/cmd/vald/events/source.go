package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/axelarnetwork/tm-events/pkg/tendermint/events"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/pubsub/query"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tm "github.com/tendermint/tendermint/types"

	"github.com/axelarnetwork/axelar-core/utils"
)

//go:generate moq -pkg mock -out ./mock/source.go . BlockSource BlockClient BlockResultClient BlockNotifier

type dialOptions struct {
	timeout   time.Duration
	retries   int
	keepAlive time.Duration
}

// DialOption for Tendermint connections
type DialOption struct {
	apply func(options dialOptions) dialOptions
}

// Timeout sets the time after which the call to Tendermint is cancelled
func Timeout(timeout time.Duration) DialOption {
	return DialOption{
		apply: func(options dialOptions) dialOptions {
			options.timeout = timeout
			return options
		},
	}
}

// Retries sets the number of times a Tendermint call is retried
func Retries(retries int) DialOption {
	return DialOption{
		apply: func(options dialOptions) dialOptions {
			options.retries = retries
			return options
		},
	}
}

// KeepAlive sets the time after which contact to Tendermint is reestablished if no there is no communication
func KeepAlive(interval time.Duration) DialOption {
	return DialOption{
		apply: func(options dialOptions) dialOptions {
			options.keepAlive = interval
			return options
		},
	}
}

// BlockNotifier notifies the caller of new blocks
type BlockNotifier interface {
	BlockHeights(ctx context.Context) (<-chan int64, <-chan error)
}

type eventblockNotifier struct {
	logger            log.Logger
	client            SubscriptionClient
	query             string
	timeout           time.Duration
	retries           int
	keepAliveInterval time.Duration
}

func newEventBlockNotifier(client SubscriptionClient, logger log.Logger, options ...DialOption) *eventblockNotifier {
	var opts dialOptions
	for _, option := range options {
		opts = option.apply(opts)
	}
	return &eventblockNotifier{
		client: client,
		logger: logger,
		query:  query.MustParse(fmt.Sprintf("%s='%s'", tm.EventTypeKey, tm.EventNewBlockHeader)).String(),

		timeout:           opts.timeout,
		retries:           opts.retries,
		keepAliveInterval: opts.keepAlive,
	}
}

func (b *eventblockNotifier) BlockHeights(ctx context.Context) (<-chan int64, <-chan error) {
	blocks := make(chan int64)
	errChan := make(chan error, 1)

	var keepAlive context.Context
	var keepAliveCancel context.CancelFunc = func() {}
	go func() {
		defer func() { keepAliveCancel() }() // the cancel function gets reassigned, so call it indirectly
		defer close(blocks)
		defer b.logger.Info("stopped listening to events")
		defer b.tryUnsubscribe(context.Background(), b.timeout)

		eventChan, err := b.subscribe(ctx, b.retries, b.timeout)
		if err != nil {
			errChan <- err
			return
		}

		for {
			keepAlive, keepAliveCancel = ctxWithTimeout(context.Background(), b.keepAliveInterval)
			var blockHeaderEvent tm.EventDataNewBlockHeader
			select {
			case <-keepAlive.Done():
				b.logger.Debug(fmt.Sprintf("no block event in %.2f seconds", b.keepAliveInterval.Seconds()))
				b.tryUnsubscribe(ctx, b.timeout)
				eventChan, err = b.subscribe(ctx, b.retries, b.timeout)
				if err != nil {
					errChan <- err
					return
				}

			case event := <-eventChan:
				var ok bool
				blockHeaderEvent, ok = event.Data.(tm.EventDataNewBlockHeader)
				if !ok {
					errChan <- fmt.Errorf("event data is of type %T, expected %T", event.Data, tm.EventDataNewBlockHeader{})
					return
				}
			case <-ctx.Done():
				return
			}

			select {
			case blocks <- blockHeaderEvent.Header.Height:
				break
			case <-ctx.Done():
				return
			}
		}
	}()

	return blocks, errChan
}

func (b *eventblockNotifier) subscribe(ctx context.Context, retries int, timeout time.Duration) (<-chan coretypes.ResultEvent, error) {
	backoff := utils.LinearBackOff(timeout)
	for i := 0; i <= retries; i++ {
		ctx, cancel := ctxWithTimeout(ctx, timeout)
		eventChan, err := b.client.Subscribe(ctx, "", b.query, events.WebsocketQueueSize)
		cancel()
		if err == nil {
			b.logger.Debug(fmt.Sprintf("subscribed to query \"%s\"", b.query))
			return eventChan, nil
		}
		b.logger.Debug(sdkerrors.Wrapf(err, "failed to subscribe to query \"%s\", attempt %d", b.query, i+1).Error())
		time.Sleep(backoff(i))
	}
	return nil, fmt.Errorf("aborting Tendermint block header subscription after %d attemts ", retries+1)
}

func (b *eventblockNotifier) tryUnsubscribe(ctx context.Context, timeout time.Duration) {
	ctx, cancel := ctxWithTimeout(ctx, timeout)
	defer cancel()

	// this unsubscribe is a best-effort action, we still try to continue as usual if it fails, so errors are only logged
	err := b.client.Unsubscribe(ctx, "", b.query)
	if err != nil {
		b.logger.Info(sdkerrors.Wrapf(err, "could not unsubscribe from query \"%s\"", b.query).Error())
		return
	}
	b.logger.Debug(fmt.Sprintf("unsubscribed from query \"%s\"", b.query))
}

type queryBlockNotifier struct {
	client            BlockHeightClient
	retries           int
	timeout           time.Duration
	keepAliveInterval time.Duration
	logger            log.Logger
}

func newQueryBlockNotifier(client BlockHeightClient, logger log.Logger, options ...DialOption) *queryBlockNotifier {
	var opts dialOptions
	for _, option := range options {
		opts = option.apply(opts)
	}

	return &queryBlockNotifier{
		client:            client,
		retries:           opts.retries,
		timeout:           opts.timeout,
		keepAliveInterval: opts.keepAlive,
		logger:            logger,
	}
}

func (q queryBlockNotifier) BlockHeights(ctx context.Context) (<-chan int64, <-chan error) {
	blocks := make(chan int64)
	errChan := make(chan error, 1)

	go func() {
		defer close(blocks)
		defer q.logger.Info("stopped block query")

		keepAlive, keepAliveCancel := ctxWithTimeout(context.Background(), q.keepAliveInterval)
		defer func() { keepAliveCancel() }() // the cancel function might get reassigned, so call it indirectly

		for {
			var latest int64
			var err error
			select {
			case <-keepAlive.Done():
				latest, err = q.latestFromSyncStatus(ctx)
				if err != nil {
					errChan <- err
					return
				}
				keepAlive, keepAliveCancel = ctxWithTimeout(context.Background(), q.keepAliveInterval)
			case <-ctx.Done():
				return
			}

			select {
			case blocks <- latest:
				break
			case <-ctx.Done():
				return
			}
		}
	}()
	return blocks, errChan
}

func (q *queryBlockNotifier) latestFromSyncStatus(ctx context.Context) (int64, error) {
	backoff := utils.LinearBackOff(q.timeout)
	for i := 0; i <= q.retries; i++ {
		ctx, cancel := ctxWithTimeout(ctx, q.timeout)
		latestBlockHeight, err := q.client.LatestBlockHeight(ctx)
		cancel()
		if err == nil {
			return latestBlockHeight, nil
		}
		q.logger.Info(sdkerrors.Wrapf(err, "failed to retrieve node status, attempt %d", i+1).Error())
		time.Sleep(backoff(i))
	}
	return 0, fmt.Errorf("aborting sync status retrieval after %d attemts ", q.retries+1)
}

// BlockHeightClient can query the latest block height
type BlockHeightClient interface {
	LatestBlockHeight(ctx context.Context) (int64, error)
}

// SubscriptionClient subscribes to and unsubscribes from Tendermint events
type SubscriptionClient interface {
	Subscribe(ctx context.Context, subscriber, query string, outCapacity ...int) (out <-chan coretypes.ResultEvent, err error)
	Unsubscribe(ctx context.Context, subscriber, query string) error
}

// BlockClient is both BlockHeightClient and SubscriptionClient
type BlockClient interface {
	BlockHeightClient
	SubscriptionClient
}

// Notifier can notify a consumer about new blocks
type Notifier struct {
	logger         log.Logger
	queryNotifier  *queryBlockNotifier
	eventsNotifier *eventblockNotifier
	running        context.Context
	shutdown       context.CancelFunc
	start          int64
}

// NewBlockNotifier returns a new BlockNotifier instance
func NewBlockNotifier(client BlockClient, logger log.Logger, options ...DialOption) *Notifier {
	var opts dialOptions
	for _, option := range options {
		opts = option.apply(opts)
	}
	ctx, cancel := ctxWithTimeout(context.Background(), opts.timeout)
	start, err := client.LatestBlockHeight(ctx)
	cancel()
	if err != nil {
		start = 1
	}

	return &Notifier{
		start:          start,
		logger:         logger,
		eventsNotifier: newEventBlockNotifier(client, logger, options...),
		queryNotifier:  newQueryBlockNotifier(client, logger, options...),
	}
}

// StartingAt sets the start block from which on to receive notifications
func (b Notifier) StartingAt(block int64) Notifier {
	if block > 0 {
		b.start = block
	}
	return b
}

// BlockHeights returns a channel with the block heights from the beginning of the chain to all newly discovered blocks.
// Optionally, starts at the given start block.
func (b Notifier) BlockHeights(ctx context.Context) (<-chan int64, <-chan error) {
	errChan := make(chan error, 1)
	b.running, b.shutdown = context.WithCancel(ctx)

	blocksFromQuery, queryErrs := b.queryNotifier.BlockHeights(b.running)
	blocksFromEvents, eventErrs := b.eventsNotifier.BlockHeights(b.running)

	go b.handleErrors(eventErrs, queryErrs, errChan)

	b.logger.Info(fmt.Sprintf("syncing blocks starting with block %d", b.start))

	blocks := make(chan int64)
	go b.pipeLatestBlock(blocksFromQuery, blocksFromEvents, blocks)

	return blocks, errChan
}

func (b Notifier) handleErrors(eventErrs <-chan error, queryErrs <-chan error, errChan chan error) {
	defer b.shutdown()

	for {
		// the query notifier is more reliable but slower, so we still continue on as long as it's available
		select {
		case err := <-queryErrs:
			errChan <- err
			return
		case err := <-eventErrs:
			b.logger.Error(sdkerrors.Wrapf(err, "cannot receive new blocks from events, falling back on querying actively for blocks").Error())
		case <-b.running.Done():
			return
		}
	}
}

func (b Notifier) pipeLatestBlock(fromQuery <-chan int64, fromEvents <-chan int64, blockHeights chan int64) {
	defer close(blockHeights)
	defer b.logger.Info("stopped block sync")

	pendingBlockHeight := b.start
	latestBlock := int64(0)
	for {
		for {
			select {
			case block := <-fromQuery:
				if latestBlock > block || pendingBlockHeight > block {
					continue
				}
				latestBlock = block
			case block := <-fromEvents:
				if latestBlock > block || pendingBlockHeight > block {
					continue
				}
				latestBlock = block
			case <-b.running.Done():
				return
			}
			break
		}

		b.logger.Debug(fmt.Sprintf("found latest block %d", latestBlock))

		for pendingBlockHeight <= latestBlock {
			select {
			case blockHeights <- pendingBlockHeight:
				pendingBlockHeight++
			case <-b.running.Done():
				return
			}
		}
	}
}

// BlockSource returns all block results sequentially
type BlockSource interface {
	BlockResults(ctx context.Context) (<-chan *coretypes.ResultBlockResults, <-chan error)
}

type blockSource struct {
	once     sync.Once
	notifier BlockNotifier
	timeout  time.Duration
	client   BlockResultClient
	shutdown context.CancelFunc
	running  context.Context
}

// BlockResultClient can query for the block results of a specific block
type BlockResultClient interface {
	BlockResults(ctx context.Context, height *int64) (*coretypes.ResultBlockResults, error)
}

// NewBlockSource returns a new BlockSource instance
func NewBlockSource(client BlockResultClient, notifier BlockNotifier, queryTimeout ...time.Duration) BlockSource {
	b := &blockSource{
		client:   client,
		notifier: notifier,
	}
	for _, timeout := range queryTimeout {
		b.timeout = timeout
	}

	return b
}

// BlockResults returns a channel of block results. Blocks are pushed into the channel sequentially as they are discovered
func (b *blockSource) BlockResults(ctx context.Context) (<-chan *coretypes.ResultBlockResults, <-chan error) {
	// either the notifier or the block source could push an error at the same time, so we need to make sure we don't block
	errChan := make(chan error, 2)
	// notifier is an external dependency, so it's lifetime should be managed by the calling process, i.e. the given context, as well
	blockHeights, notifyErrs := b.notifier.BlockHeights(ctx)

	b.running, b.shutdown = context.WithCancel(ctx)
	go func() {
		defer b.shutdown()

		select {
		case err := <-notifyErrs:
			errChan <- err
			return
		case <-ctx.Done():
			return
		}
	}()

	blockResults := make(chan *coretypes.ResultBlockResults)
	go b.streamBlockResults(blockHeights, blockResults, errChan)

	return blockResults, errChan
}

func (b *blockSource) streamBlockResults(blockHeights <-chan int64, blocks chan *coretypes.ResultBlockResults, errChan chan error) {
	defer close(blocks)
	defer b.shutdown()

	for {
		select {
		case height, ok := <-blockHeights:
			if !ok {
				errChan <- fmt.Errorf("cannot detect new blocks anymore")
				return
			}
			ctx, cancel := ctxWithTimeout(b.running, b.timeout)
			result, err := b.client.BlockResults(ctx, &height)
			cancel()
			if err != nil {
				errChan <- err
				return
			}

			blocks <- result
		case <-b.running.Done():
			return
		}
	}
}

func ctxWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}
