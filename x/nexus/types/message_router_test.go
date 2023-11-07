package types_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	. "github.com/axelarnetwork/utils/test"
)

func TestAddRoute(t *testing.T) {
	var (
		router types.MessageRouter
		module string
	)

	givenRouter := Given("a message router", func() {
		router = types.NewMessageRouter()
	})

	givenRouter.
		When("it is sealed", func() {
			router.Seal()
			module = "module"
		}).
		Then("it panics when adding a route", func(t *testing.T) {
			assert.PanicsWithValue(t, "cannot add route (router sealed)", func() {
				router.AddRoute(module, nil)
			})
		}).
		Run(t)

	givenRouter.
		When("module is empty", func() {
			module = ""
		}).
		Then("it panics when adding a route", func(t *testing.T) {
			assert.PanicsWithValue(t, "module name cannot be an empty string", func() {
				router.AddRoute(module, nil)
			})
		}).
		Run(t)

	givenRouter.
		When("module route is added already", func() {
			module = "module"
			router.AddRoute(module, func(_ sdk.Context, _ exported.RoutingContext, _ exported.GeneralMessage) error {
				return nil
			})
		}).
		Then("it panics when adding a route again", func(t *testing.T) {
			assert.PanicsWithValue(t, fmt.Sprintf("route for module %s has already been registered", module), func() {
				router.AddRoute(module, func(_ sdk.Context, _ exported.RoutingContext, _ exported.GeneralMessage) error {
					return nil
				})
			})
		}).
		Run(t)
}

func TestRoute(t *testing.T) {
	var (
		ctx        sdk.Context
		routingCtx exported.RoutingContext
		msg        exported.GeneralMessage
		router     types.MessageRouter
		module     string
		routeCount uint
		route      exported.MessageRoute
	)

	givenRouter := Given("a message router", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		router = types.NewMessageRouter()
	})

	givenRouter.
		When("it is not sealed", func() {}).
		Then("it panics when routing a message", func(t *testing.T) {
			assert.PanicsWithValue(t, "cannot route message (router not sealed)", func() {
				router.Route(ctx, exported.RoutingContext{}, exported.GeneralMessage{})
			})
		}).
		Run(t)

	whenIsSealed := When("it is sealed", func() {
		router.Seal()
	})

	givenRouter.
		When2(whenIsSealed).
		When("module is not found", func() {
			msg = exported.GeneralMessage{Recipient: exported.CrossChainAddress{Chain: exported.Chain{Module: "unknown"}}}
		}).
		Then("it should return error", func(t *testing.T) {
			assert.ErrorContains(t, router.Route(ctx, routingCtx, msg), "no route found")
		}).
		Run(t)

	givenRouter.
		When("route is added", func() {
			module = "module"
			routeCount = 0
			route = func(_ sdk.Context, _ exported.RoutingContext, msg exported.GeneralMessage) error {
				routeCount++
				return nil
			}

			router.AddRoute(module, route)
		}).
		When2(whenIsSealed).
		Branch(
			When("payload is provided but does not match the payload hash", func() {
				routingCtx = exported.RoutingContext{Payload: []byte("payload")}
				msg = exported.GeneralMessage{PayloadHash: rand.Bytes(common.HashLength), Recipient: exported.CrossChainAddress{Chain: exported.Chain{Module: module}}}
			}).
				Then("it should return error", func(t *testing.T) {
					assert.ErrorContains(t, router.Route(ctx, routingCtx, msg), "payload hash does not match")
				}),

			When("payload is provided and matches the payload hash", func() {
				payload := rand.Bytes(100)
				routingCtx = exported.RoutingContext{Payload: payload}
				msg = exported.GeneralMessage{PayloadHash: crypto.Keccak256Hash(payload).Bytes(), Recipient: exported.CrossChainAddress{Chain: exported.Chain{Module: module}}}
			}).
				Then("it should succeed", func(t *testing.T) {
					assert.NoError(t, router.Route(ctx, routingCtx, msg), "payload hash does not match")
					assert.Equal(t, uint(1), routeCount)
				}),

			When("payload is not provided", func() {
				routingCtx = exported.RoutingContext{Payload: nil}
				msg = exported.GeneralMessage{Recipient: exported.CrossChainAddress{Chain: exported.Chain{Module: module}}}
			}).
				Then("it should succeed", func(t *testing.T) {
					assert.NoError(t, router.Route(ctx, routingCtx, msg), "payload hash does not match")
					assert.Equal(t, uint(1), routeCount)
				}),
		).
		Run(t)
}
