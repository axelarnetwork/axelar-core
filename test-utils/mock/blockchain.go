package mock

import (
	"fmt"
	"log"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type BlockChain struct {
	blockSize    int
	in           chan sdk.Msg
	nodes        []Node
	blockTimeOut time.Duration
}

type block struct {
	msgs   []sdk.Msg
	height int64
}

func newBlock(size int, height int64) block {
	return block{msgs: make([]sdk.Msg, 0, size), height: height}
}

// NewBlockchain returns a mocked blockchain with default parameters.
// Use the With* functions to specify different parameters.
// By default, the blockchain does not time out,
// so a block will only be disseminated once the specified block size is reached.
func NewBlockchain() BlockChain {
	return BlockChain{
		blockSize:    1,
		blockTimeOut: 0,
		in:           make(chan sdk.Msg, 1000),
		nodes:        make([]Node, 0),
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

// Input returns the message input of the blockchain. Any message received on this channel will be put into a block.
func (bc BlockChain) Input() chan<- sdk.Msg {
	return bc.in
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
	go disseminateBlocks(bc)
}

func disseminateBlocks(bc BlockChain) {
	for b := range cutBlocks(bc.in, bc.blockSize, bc.blockTimeOut) {
		for _, n := range bc.nodes {
			n.in <- b
		}
	}
}

func cutBlocks(msgs <-chan sdk.Msg, blockSize int, timeOut time.Duration) <-chan block {
	blocks := make(chan block, 1)
	go func() {
		// close block channel when message channel is closed
		defer close(blocks)
		nextBlock := newBlock(blockSize, 1)

	loop:
		for {
			timedOut := reset(timeOut)

			select {
			case msg, ok := <-msgs:
				// channel is closed, send what you have and then stop
				if !ok {
					blocks <- nextBlock
					break loop
				}

				nextBlock.msgs = append(nextBlock.msgs, msg)
				if len(nextBlock.msgs) == blockSize {
					blocks <- nextBlock
					nextBlock = newBlock(blockSize, nextBlock.height+1)
				}
			// timeout happened before receiving a message, cut the block here and start a new one
			case <-timedOut:
				blocks <- nextBlock
				nextBlock = newBlock(blockSize, nextBlock.height+1)
			}
		}
	}()
	return blocks
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
	handlers    map[string]sdk.Handler
	endBlockers []func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate
	Ctx         sdk.Context
	moniker     string
}

// NewNode creates a new node that can be added to the blockchain.
// The moniker is used to differentiate nodes for logging purposes.
// The context will be passed on to the registered handlers.
func NewNode(moniker string, ctx sdk.Context) Node {
	return Node{
		moniker:     moniker,
		Ctx:         ctx,
		in:          make(chan block, 1),
		handlers:    make(map[string]sdk.Handler, 0),
		endBlockers: make([]func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate, 0),
	}
}

// WithHandler returns a node with a handler for the specified module.
func (n Node) WithHandler(moduleName string, handler sdk.Handler) Node {
	n.handlers[moduleName] = handler
	return n
}

// WithEndBlockers returns a node with the specified EndBlocker functions.
// They are executed in the order they are provided.
func (n Node) WithEndBlockers(endBlockers ...func(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate) Node {
	n.endBlockers = append(n.endBlockers, endBlockers...)
	return n
}

func (n Node) start() {
	for b := range n.in {
		log.Printf("node %s begins block %v", n.moniker, b.height)
		/*
			While Cosmos also has BeginBlockers, so far we implement none.
			Extend the Node struct analogously to the EndBlockers
			and add any logic that deals with the begin of a block here when necessary
		*/

		// handle messages
		for _, msg := range b.msgs {
			if h, ok := n.handlers[msg.Route()]; ok {
				if err := msg.ValidateBasic(); err != nil {
					log.Printf("node %s returned an error when validating message %s", n.moniker, msg.Type())
				}
				if _, err := h(n.Ctx, msg); err != nil {
					log.Printf("node %s returned an error from handler for route %s", n.moniker, msg.Route())
				}
			} else {
				panic(fmt.Sprintf("no handler for route %s defined", msg.Route()))
			}
		}

		log.Printf("node %s ends block %v", n.moniker, b.height)
		// end block
		for _, endBlocker := range n.endBlockers {
			endBlocker(n.Ctx, abci.RequestEndBlock{Height: b.height})
		}
	}
}
