package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/ante/types"
)

func logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// HandlerDecorator is an ante decorator wrapper for an ante handler
type HandlerDecorator struct {
	handler sdk.AnteHandler
}

// NewAnteHandlerDecorator constructor for HandlerDecorator
func NewAnteHandlerDecorator(handler sdk.AnteHandler) HandlerDecorator {
	return HandlerDecorator{handler}
}

// AnteHandle wraps the next AnteHandler to perform custom pre- and post-processing
func (decorator HandlerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if newCtx, err = decorator.handler(ctx, tx, simulate); err != nil {
		return newCtx, err
	}

	return next(newCtx, tx, simulate)
}

// MessageAnteHandler is analogous to the sdk's AnteHandler, but works on messages instead of the full Tx
type MessageAnteHandler func(ctx sdk.Context, msgs []sdk.Msg, simulate bool) (sdk.Context, error)

// ToAnteHandler converts a MessageAnteHandler to an AnteHandler
func (m MessageAnteHandler) ToAnteHandler() sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		msgs := tx.GetMsgs()
		return m(ctx, msgs, simulate)
	}
}

// MessageAnteDecorator is analogous to the sdk's AnteDecorator, but works on messages instead of the full Tx
type MessageAnteDecorator interface {
	AnteHandle(ctx sdk.Context, msgs []sdk.Msg, simulate bool, next MessageAnteHandler) (sdk.Context, error)
}

// ChainMessageAnteDecorators combines multiple MessageAnteDecorators into a single MessageAnteHandler
func ChainMessageAnteDecorators(chain ...MessageAnteDecorator) MessageAnteHandler {
	if len(chain) == 0 {
		return nil
	}

	// handle non-terminated decorators chain
	if (chain[len(chain)-1] != terminator{}) {
		chain = append(chain, terminator{})
	}

	return func(ctx sdk.Context, msgs []sdk.Msg, simulate bool) (sdk.Context, error) {
		return chain[0].AnteHandle(ctx, msgs, simulate, ChainMessageAnteDecorators(chain[1:]...))
	}
}

type terminator struct{}

// AnteHandle implements MessageAnteDecorator
func (terminator) AnteHandle(ctx sdk.Context, _ []sdk.Msg, _ bool, _ MessageAnteHandler) (sdk.Context, error) {
	return ctx, nil
}
