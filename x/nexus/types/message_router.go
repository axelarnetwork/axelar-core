package types

import (
	"fmt"

	exported "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MessageRouter implements a message router based on the message's destination
// chain's  module name
type MessageRouter interface {
	AddRoute(module string, router exported.MessageRouter) MessageRouter
	Route(ctx sdk.Context, routingCtx exported.RoutingContext, msg exported.GeneralMessage) error
	Seal()
}

var _ MessageRouter = (*messageRouter)(nil)

type messageRouter struct {
	routes map[string]exported.MessageRouter
	sealed bool
}

// NewMessageRouter creates a new MessageRouter interface instance
func NewMessageRouter() MessageRouter {
	return &messageRouter{
		routes: make(map[string]exported.MessageRouter),
		sealed: false,
	}
}

func (r *messageRouter) AddRoute(module string, router exported.MessageRouter) MessageRouter {
	if r.sealed {
		panic("cannot add handler (router sealed)")
	}

	if module == "" {
		panic("module name cannot be an empty string")
	}

	if _, ok := r.routes[module]; ok {
		panic(fmt.Sprintf("router for module %s has already been registered", module))
	}

	r.routes[module] = router

	return r
}

func (r messageRouter) Route(ctx sdk.Context, routingCtx exported.RoutingContext, msg exported.GeneralMessage) error {
	if !r.sealed {
		panic("cannot route message (router not sealed)")
	}

	router, ok := r.routes[msg.Recipient.Chain.Module]
	if !ok {
		return fmt.Errorf("no router found for module %s", msg.Recipient.Chain.Module)
	}

	if routingCtx.Payload != nil && !msg.Match(routingCtx.Payload) {
		return fmt.Errorf("payload hash does not match")
	}

	return router(ctx, routingCtx, msg)
}

func (r *messageRouter) Seal() {
	r.sealed = true
}
