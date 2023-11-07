package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	exported "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// MessageRouter implements a message router based on the message's destination
// chain's module name
type MessageRouter interface {
	AddRoute(module string, route exported.MessageRoute) MessageRouter
	Route(ctx sdk.Context, routingCtx exported.RoutingContext, msg exported.GeneralMessage) error
	Seal()
}

var _ MessageRouter = (*messageRouter)(nil)

type messageRouter struct {
	routes map[string]exported.MessageRoute
	sealed bool
}

// NewMessageRouter creates a new MessageRouter interface instance
func NewMessageRouter() MessageRouter {
	return &messageRouter{
		routes: make(map[string]exported.MessageRoute),
		sealed: false,
	}
}

func (r *messageRouter) AddRoute(module string, route exported.MessageRoute) MessageRouter {
	if r.sealed {
		panic("cannot add route (router sealed)")
	}

	if module == "" {
		panic("module name cannot be an empty string")
	}

	if _, ok := r.routes[module]; ok {
		panic(fmt.Sprintf("route for module %s has already been registered", module))
	}

	r.routes[module] = route

	return r
}

func (r messageRouter) Route(ctx sdk.Context, routingCtx exported.RoutingContext, msg exported.GeneralMessage) error {
	if !r.sealed {
		panic("cannot route message (router not sealed)")
	}

	route, ok := r.routes[msg.Recipient.Chain.Module]
	if !ok {
		return fmt.Errorf("no route found for module %s", msg.Recipient.Chain.Module)
	}

	if routingCtx.Payload != nil && !msg.Match(routingCtx.Payload) {
		return fmt.Errorf("payload hash does not match")
	}

	return route(ctx, routingCtx, msg)
}

func (r *messageRouter) Seal() {
	r.sealed = true
}
