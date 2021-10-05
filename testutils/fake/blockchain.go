package fake

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/legacy/legacytx"
	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// Result contains either the result of a successful message execution or the error that occurred
type Result struct {
	*sdk.Result
	Error error
}

// BlockChain is a fake that emulates the behaviour of a full blockchain network (consensus and message dissemination)
// for testing
type BlockChain struct {
	blockSize int
	in        chan struct {
		sdk.Msg
		out chan<- *Result
	}
	nodes         []*Node
	blockTimeOut  time.Duration
	currentHeight int64
	blockListener func(block)
}

type block struct {
	msgs []struct {
		sdk.Msg
		out chan<- *Result
	}
	header tmproto.Header
}

func newBlock(size int, header tmproto.Header) block {
	return block{msgs: make([]struct {
		sdk.Msg
		out chan<- *Result
	}, 0, size), header: header}
}

// NewBlockchain returns a faked blockchain with default parameters.
// Use the With* functions to specify different parameters.
// By default, the blockchain does not time out,
// so a block will only be disseminated once the specified block size is reached.
func NewBlockchain() *BlockChain {
	return &BlockChain{
		blockSize:    1,
		blockTimeOut: 0,
		in: make(chan struct {
			sdk.Msg
			out chan<- *Result
		}, 1000),
		nodes:         make([]*Node, 0),
		currentHeight: 0,
		blockListener: func(block) {},
	}
}

// WithBlockSize returns a blockchain with blocks of at most the specified size.
func (bc BlockChain) WithBlockSize(size int) *BlockChain {
	if size < 1 {
		panic("block size must be at least 1")
	}
	newChain := deepCopy(bc)
	newChain.blockSize = size
	return newChain
}

// WithBlockTimeOut returns a blockchain with a timeout. The timeout resets whenever a message is received.
// When the timer runs out it disseminates the next block regardless of its size.
func (bc BlockChain) WithBlockTimeOut(timeOut time.Duration) *BlockChain {
	newChain := deepCopy(bc)
	newChain.blockTimeOut = timeOut
	return newChain
}

// Submit sends a message to the blockchain. It returns a channel with the result.
func (bc *BlockChain) Submit(msg sdk.Msg) <-chan *Result {
	// all nodes will push their output into this channel
	out := make(chan *Result, len(bc.nodes))
	bc.in <- struct {
		sdk.Msg
		out chan<- *Result
	}{msg, out}

	// check that all nodes return the same result
	result := make(chan *Result, 1)
	go func() {
		var r *Result
		for i := 0; i < cap(out); i++ {
			temp := <-out
			if r == nil {
				r = temp
			} else if temp == nil || !equals(*r, *temp) {
				panic(fmt.Sprintf("expected %v, got %v", r, temp))
			}
		}
		if r == nil {
			panic("no result")
		}
		result <- r
	}()
	return result
}

func equals(this Result, other Result) bool {
	if this.Error != nil && other.Error != nil && this.Error.Error() == other.Error.Error() {
		return true
	}
	if this.Result != nil && other.Result != nil && this.Log == other.Log && bytes.Equal(this.Data, other.Data) {
		return true
	}

	return false
}

// AddNodes adds a node to the blockchain. This node will receive blocks from the blockchain.
func (bc *BlockChain) AddNodes(nodes ...*Node) {
	bc.nodes = append(bc.nodes, nodes...)
}

// Start starts the block dissemination. Only call once all parameters and nodes are fully set up.
func (bc *BlockChain) Start() {
	for _, n := range bc.nodes {
		go n.start()
	}
	go bc.disseminateBlocks()
}

// CurrentHeight returns the current block height.
func (bc BlockChain) CurrentHeight() int64 {
	return bc.currentHeight
}

func (bc *BlockChain) disseminateBlocks() {
	for b := range bc.cutBlocks() {
		bc.blockListener(b)
		for _, n := range bc.nodes {
			n.in <- b
		}
	}
}

func (bc *BlockChain) cutBlocks() <-chan block {
	blocks := make(chan block, 1)
	go func() {
		// close block channel when message channel is closed
		defer close(blocks)
		nextBlock := newBlock(bc.blockSize, tmproto.Header{Height: bc.CurrentHeight(), Time: time.Now()})
		bc.currentHeight++

		for {
			timeOut, cancel := context.WithTimeout(context.Background(), bc.blockTimeOut)
		timeOutloop:
			for {
				select {
				case msg, ok := <-bc.in:
					// channel is closed, send what you have and then stop
					if !ok {
						blocks <- nextBlock
						cancel()
						return
					}

					nextBlock.msgs = append(nextBlock.msgs, msg)
					if len(nextBlock.msgs) == bc.blockSize {
						blocks <- nextBlock
						nextBlock = newBlock(bc.blockSize, tmproto.Header{Height: bc.CurrentHeight(), Time: time.Now()})
						bc.currentHeight++

					}
				// timeout happened before receiving a message, cut the block here and start a new one
				case <-timeOut.Done():
					blocks <- nextBlock
					nextBlock = newBlock(bc.blockSize, tmproto.Header{Height: bc.CurrentHeight(), Time: time.Now()})
					bc.currentHeight++

					cancel()
					break timeOutloop
				}
			}
		}
	}()
	return blocks
}

// WaitNBlocks waits for n blocks to be disseminated before returning. Do not use without setting a block timeout or the test
// will deadlock.
func (bc *BlockChain) WaitNBlocks(n int64) {
	currHeight := bc.CurrentHeight()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	bc.blockListener = func(b block) {
		if b.header.Height-currHeight >= n {
			bc.blockListener = func(block) {}
			wg.Done()
		}
	}
	wg.Wait()
}

func deepCopy(bc BlockChain) *BlockChain {
	newChain := NewBlockchain()
	newChain.blockSize = bc.blockSize
	newChain.blockTimeOut = bc.blockTimeOut
	newChain.currentHeight = bc.currentHeight
	newChain.nodes = append(newChain.nodes, bc.nodes...)

	return newChain
}

// Node is a fake that emulates the behaviour of a Cosmos node by retrieving blocks from the network,
// unpacking the messages and routing them to the correct modules
type Node struct {
	in             chan block
	router         sdk.Router
	endBlockers    []func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate
	Ctx            sdk.Context
	Moniker        string
	queriers       map[string]sdk.Querier
	events         chan abci.Event
	eventListeners []struct {
		predicate func(event abci.Event) bool
		emitter   chan<- abci.Event
	}
}

// NewNode creates a new node that can be added to the blockchain.
// The moniker is used to differentiate nodes for logging purposes.
// The context will be passed on to the registered handlers.
func NewNode(moniker string, ctx sdk.Context, router sdk.Router, queriers map[string]sdk.Querier) *Node {
	return &Node{
		in:          make(chan block, 1),
		router:      router,
		endBlockers: make([]func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate, 0),
		Ctx:         ctx,
		Moniker:     moniker,
		queriers:    queriers,
		events:      make(chan abci.Event, 10000),
		eventListeners: []struct {
			predicate func(event abci.Event) bool
			emitter   chan<- abci.Event
		}{{predicate: func(event abci.Event) bool { return false }, emitter: nil}}, // default discard listener
	}
}

// WithEndBlockers returns a node with the specified EndBlocker functions.
// They are executed in the order they are provided.
func (n *Node) WithEndBlockers(endBlockers ...func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate) *Node {
	n.endBlockers = append(n.endBlockers, endBlockers...)
	return n
}

// Query allows to query a node. Returns a serialized response
func (n Node) Query(path []string, query abci.RequestQuery) ([]byte, error) {
	return n.queriers[path[0]](n.Ctx, path[1:], query)
}

// RegisterEventListener registers a listener for events that satisfy the predicate. Events will be dropped if the event channel fills up
func (n *Node) RegisterEventListener(predicate func(event abci.Event) bool) <-chan abci.Event {
	out := make(chan abci.Event, 100)

	n.eventListeners = append(n.eventListeners, struct {
		predicate func(event abci.Event) bool
		emitter   chan<- abci.Event
	}{predicate: predicate, emitter: out})
	return out
}

func (n *Node) start() {
	defer close(n.events)
	go func() {
		for e := range n.events {
			for _, l := range n.eventListeners {
				if l.predicate(e) {
					if len(l.emitter) >= cap(l.emitter) {
						panic(fmt.Sprintf("node %s event listener ran out of space", n.Moniker))
					}
					l.emitter <- e
				}
			}
		}
	}()

	for b := range n.in {
		n.Ctx = n.Ctx.WithBlockHeader(b.header)
		n.Ctx.Logger().Debug(fmt.Sprintf("begin block %v", b.header.Height))
		/*
			While Cosmos also has BeginBlockers, so far we implement none.
			Extend the Node struct analogously to the EndBlockers
			and add any logic that deals with the begin of a block here when necessary
		*/

		// handle messages
		for _, msg := range b.msgs {
			if err := msg.ValidateBasic(); err != nil {
				n.Ctx.Logger().Error(fmt.Sprintf("error when validating message %s", proto.MessageName(msg)))

				msg.out <- &Result{nil, err}

			} else if legacyMsg, ok := msg.Msg.(legacytx.LegacyMsg); ok {
				msgRoute := legacyMsg.Route()
				handler := n.router.Route(n.Ctx, msgRoute)
				if handler == nil {
					panic(fmt.Sprintf("no handler for route %s defined", msgRoute))
				}

				res, err := handler(n.Ctx, msg.Msg)
				if err != nil {
					n.Ctx.Logger().Error(fmt.Sprintf("error from handler for route %s: %s", msgRoute, err.Error()))
					// to allow failed messages we need to implement a cache for the multistore to revert in case of failure
					// outputing the error message here so that we can have a sense for why it panics in case verbose mode is not active.
					panic(fmt.Sprintf("no failing messages allowed for now: error from handler for route %s: %s\nmessage: %v", msgRoute, err.Error(), msg))
				}
				msgEvents := sdk.Events{
					sdk.NewEvent(sdk.EventTypeMessage, sdk.NewAttribute(sdk.AttributeKeyAction, proto.MessageName(msg))),
				}

				if res != nil {
					msgEvents = msgEvents.AppendEvents(res.GetEvents())
				}

				events := msgEvents.ToABCIEvents()
				for _, event := range events {
					if len(n.events) >= cap(n.events) {
						panic(fmt.Sprintf("node %s event queue ran out of space", n.Moniker))
					}
					n.events <- event
				}
				msg.out <- &Result{res, err}
			} else {
				panic(fmt.Sprintf("can't route message %+v", msg))
			}
		}
		// end block
		for _, endBlocker := range n.endBlockers {
			previousEventCount := len(n.Ctx.EventManager().ABCIEvents())
			endBlocker(n.Ctx, abci.RequestEndBlock{Height: b.header.Height})
			newEvents := n.Ctx.EventManager().ABCIEvents()[previousEventCount:]

			for _, event := range newEvents {
				n.events <- event
			}
		}
	}
}

// Router is a fake that is used by the Node to route messages to the correct module handlers
type Router struct {
	handlers map[string]sdk.Handler
}

// NewRouter returns a new Router that deals with handler routing
func NewRouter() sdk.Router {
	return Router{handlers: map[string]sdk.Handler{}}
}

// AddRoute adds a new handler route
func (r Router) AddRoute(route sdk.Route) sdk.Router {
	r.handlers[route.Path()] = route.Handler()
	return r
}

// Route tries to route the given path to a registered handler. Returns nil when the path is not found.
func (r Router) Route(_ sdk.Context, path string) sdk.Handler {
	h, ok := r.handlers[path]
	if !ok {
		return nil
	}
	return h
}
