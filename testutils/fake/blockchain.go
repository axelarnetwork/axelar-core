package fake

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type Result struct {
	*sdk.Result
	Error error
}

type BlockChain struct {
	blockSize int
	in        chan struct {
		sdk.Msg
		out chan<- Result
	}
	nodes         []Node
	blockTimeOut  time.Duration
	currentHeight *int64
}

type block struct {
	msgs []struct {
		sdk.Msg
		out chan<- Result
	}
	header abci.Header
}

func newBlock(size int, header abci.Header) block {
	return block{msgs: make([]struct {
		sdk.Msg
		out chan<- Result
	}, 0, size), header: header}
}

// NewBlockchain returns a mocked blockchain with default parameters.
// Use the With* functions to specify different parameters.
// By default, the blockchain does not time out,
// so a block will only be disseminated once the specified block size is reached.
func NewBlockchain() BlockChain {
	return BlockChain{
		blockSize:    1,
		blockTimeOut: 0,
		in: make(chan struct {
			sdk.Msg
			out chan<- Result
		}, 1000),
		nodes:         make([]Node, 0),
		currentHeight: new(int64),
	}
}

// WithBlockSize returns a blockchain with blocks of at most the specified size.
func (bc BlockChain) WithBlockSize(size int) BlockChain {
	if size < 1 {
		panic("block size must be at least 1")
	}
	bc.blockSize = size
	return bc
}

// WithBlockTimeOut returns a blockchain with a timeout. The timeout resets whenever a message is received.
// When the timer runs out it disseminates the next block regardless of its size.
func (bc BlockChain) WithBlockTimeOut(timeOut time.Duration) BlockChain {
	bc.blockTimeOut = timeOut
	return bc
}

// Submit sends a message to the blockchain. It returns a channel with the result.
func (bc BlockChain) Submit(msg sdk.Msg) (result <-chan Result) {
	// all nodes will push their output into this channel
	out := make(chan Result, 1000)
	bc.in <- struct {
		sdk.Msg
		out chan<- Result
	}{msg, out}
	return out
}

// AddNode adds a node to the blockchain. This node will receive blocks from the blockchain.
func (bc *BlockChain) AddNodes(nodes ...Node) {
	bc.nodes = append(bc.nodes, nodes...)
}

// Start starts the block dissemination. Only call once all parameters and nodes are fully set up.
func (bc BlockChain) Start() {
	for _, n := range bc.nodes {
		go n.start()
	}
	go bc.disseminateBlocks()
}

func (bc BlockChain) CurrentHeight() int64 {
	return *bc.currentHeight
}

func (bc BlockChain) disseminateBlocks() {
	for b := range bc.cutBlocks() {
		for _, n := range bc.nodes {
			n.in <- b
		}
	}
}

func (bc BlockChain) cutBlocks() <-chan block {
	blocks := make(chan block, 1)
	go func() {
		// close block channel when message channel is closed
		defer close(blocks)
		nextBlock := newBlock(bc.blockSize, abci.Header{Height: bc.CurrentHeight(), Time: time.Now()})
		atomic.AddInt64(bc.currentHeight, 1)

	loop:
		for {
			timedOut := reset(bc.blockTimeOut)

			select {
			case msg, ok := <-bc.in:
				// channel is closed, send what you have and then stop
				if !ok {
					blocks <- nextBlock
					break loop
				}

				nextBlock.msgs = append(nextBlock.msgs, msg)
				if len(nextBlock.msgs) == bc.blockSize {
					blocks <- nextBlock
					nextBlock = newBlock(bc.blockSize, abci.Header{Height: bc.CurrentHeight(), Time: time.Now()})
					atomic.AddInt64(bc.currentHeight, 1)

				}
			// timeout happened before receiving a message, cut the block here and start a new one
			case <-timedOut:
				blocks <- nextBlock
				nextBlock = newBlock(bc.blockSize, abci.Header{Height: bc.CurrentHeight(), Time: time.Now()})
				atomic.AddInt64(bc.currentHeight, 1)
			}
		}
	}()
	return blocks
}

func (bc BlockChain) WaitNBlocks(n int64) chan struct{} {
	currHeight := bc.CurrentHeight()
	done := make(chan struct{})

	go func() {
		for {
			if bc.CurrentHeight() >= currHeight+n {
				close(done)
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	return done
}

func reset(timeOut time.Duration) chan struct{} {
	var timedOut chan struct{}

	if timeOut > 0 {
		timedOut = make(chan struct{})
		go func() {
			time.Sleep(timeOut)
			close(timedOut)
		}()
	}
	return timedOut
}

type Node struct {
	in          chan block
	router      sdk.Router
	endBlockers []func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate
	Ctx         sdk.Context
	moniker     string
	queriers    map[string]sdk.Querier
}

// NewNode creates a new node that can be added to the blockchain.
// The moniker is used to differentiate nodes for logging purposes.
// The context will be passed on to the registered handlers.
func NewNode(moniker string, ctx sdk.Context, router sdk.Router, queriers map[string]sdk.Querier) Node {
	return Node{
		moniker:     moniker,
		Ctx:         ctx,
		in:          make(chan block, 1),
		router:      router,
		endBlockers: make([]func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate, 0),
		queriers:    queriers,
	}
}

// WithEndBlockers returns a node with the specified EndBlocker functions.
// They are executed in the order they are provided.
func (n Node) WithEndBlockers(endBlockers ...func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate) Node {
	n.endBlockers = append(n.endBlockers, endBlockers...)
	return n
}

func (n Node) Query(path []string) ([]byte, error) {
	return n.queriers[path[0]](n.Ctx, path[1:], abci.RequestQuery{})
}

func (n Node) start() {
	for b := range n.in {
		n.Ctx = n.Ctx.WithBlockHeader(b.header)
		log.Printf("node %s begins block %v", n.moniker, b.header.Height)
		/*
			While Cosmos also has BeginBlockers, so far we implement none.
			Extend the Node struct analogously to the EndBlockers
			and add any logic that deals with the begin of a block here when necessary
		*/

		// handle messages
		for _, msg := range b.msgs {
			if err := msg.ValidateBasic(); err != nil {
				log.Printf("node %s returned an error when validating message %s", n.moniker, msg.Type())

				// another node already returned a result
				if len(msg.out) > 0 {
					continue
				}
				msg.out <- Result{nil, err}

			} else if h := n.router.Route(n.Ctx, msg.Route()); h != nil {
				res, err := h(n.Ctx, msg.Msg)
				if err != nil {
					log.Printf("node %s returned an error from handler for route %s: %s", n.moniker, msg.Route(), err.Error())
				}
				// another node already returned a result
				if len(msg.out) > 0 {
					continue
				}
				msg.out <- Result{res, err}

			} else {
				panic(fmt.Sprintf("no handler for route %s defined", msg.Route()))
			}
		}

		log.Printf("node %s ends block %v", n.moniker, b.header.Height)
		// end block
		for _, endBlocker := range n.endBlockers {
			endBlocker(n.Ctx, abci.RequestEndBlock{Height: b.header.Height})
		}
	}
}

type Router struct {
	handlers map[string]sdk.Handler
}

// NewRouter returns a new Router that deals with handler routing
func NewRouter() sdk.Router {
	return Router{handlers: map[string]sdk.Handler{}}
}

// AddRoute adds a new handler route
func (r Router) AddRoute(moduleName string, h sdk.Handler) sdk.Router {
	r.handlers[moduleName] = h
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
