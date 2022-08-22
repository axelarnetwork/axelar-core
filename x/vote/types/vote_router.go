package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/vote/exported"
)

// VoteRouter implements a Vote router based on module name.
type VoteRouter interface {
	AddHandler(module string, handler exported.VoteHandler) VoteRouter
	HasHandler(module string) bool
	GetHandler(module string) exported.VoteHandler
	Seal()
}

var _ VoteRouter = (*router)(nil)

type router struct {
	routes map[string]exported.VoteHandler
	sealed bool
}

// NewRouter creates a new Router interface instance
func NewRouter() VoteRouter {
	return &router{
		routes: make(map[string]exported.VoteHandler),
	}
}

// Seal prevents additional route handlers from  being added to the router.
func (r *router) Seal() {
	r.sealed = true
}

// AddHandler registers a vote handler for a given path and returns the handler.
// Panics if the router is sealed, module is an empty string, or if the module has been registered already.
func (r *router) AddHandler(module string, handler exported.VoteHandler) VoteRouter {
	if r.sealed {
		panic("cannot add handler (router sealed)")
	}

	if module == "" {
		panic("module name cannot be an empty string")
	}

	if r.HasHandler(module) {
		panic(fmt.Sprintf("handler for module %s has already been registered", module))
	}

	r.routes[module] = handler
	return r
}

// HasHandler returns true if the router has an handler registered for the given module
func (r *router) HasHandler(module string) bool {
	return r.routes[module] != nil
}

// GetHandler returns a Handler for a given module.
func (r *router) GetHandler(module string) exported.VoteHandler {
	if !r.HasHandler(module) {
		panic(fmt.Sprintf("handler for module \"%s\" not registered", module))
	}

	return handlerWrapper{r.routes[module]}
}

type handlerWrapper struct {
	handler exported.VoteHandler
}

func (w handlerWrapper) IsFalsyResult(result codec.ProtoMarshaler) bool {
	return w.handler.IsFalsyResult(result)
}

func (w handlerWrapper) HandleResult(ctx sdk.Context, result codec.ProtoMarshaler) error {
	return cache(w.handler.HandleResult)(ctx, result)
}

func (w handlerWrapper) HandleFailedPoll(ctx sdk.Context, poll exported.Poll) error {
	return cache(w.handler.HandleFailedPoll)(ctx, poll)
}

func (w handlerWrapper) HandleCompletedPoll(ctx sdk.Context, poll exported.Poll) error {
	return cache(w.handler.HandleCompletedPoll)(ctx, poll)
}

func (w handlerWrapper) HandleExpiredPoll(ctx sdk.Context, poll exported.Poll) error {
	return cache(w.handler.HandleExpiredPoll)(ctx, poll)
}

func cache[T any](f func(ctx sdk.Context, x T) error) func(ctx sdk.Context, x T) error {
	return func(ctx sdk.Context, x T) error {
		cachedCtx, writeCache := ctx.CacheContext()
		if err := f(cachedCtx, x); err != nil {
			return err
		}
		writeCache()
		ctx.EventManager().EmitEvents(cachedCtx.EventManager().Events())
		return nil
	}
}
