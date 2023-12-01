package keeper_test

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
	. "github.com/axelarnetwork/utils/test"
)

func TestNewMessageRoute(t *testing.T) {
	var (
		ctx   sdk.Context
		route nexus.MessageRoute
		msg   nexus.GeneralMessage

		nexusK   *mock.NexusMock
		accountK *mock.AccountKeeperMock
		wasmK    *mock.WasmKeeperMock
		gateway  sdk.AccAddress
	)

	givenMessageRoute := Given("the message route", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

		nexusK = &mock.NexusMock{}
		accountK = &mock.AccountKeeperMock{}
		wasmK = &mock.WasmKeeperMock{}

		route = keeper.NewMessageRoute(nexusK, accountK, wasmK)
	})

	givenMessageRoute.
		When("the gateway is not set", func() {
			nexusK.GetParamsFunc = func(ctx sdk.Context) types.Params { return types.DefaultParams() }
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, route(ctx, nexus.RoutingContext{}, msg), "gateway is not set")
		}).
		Run(t)

	givenMessageRoute.
		When("the gateway is set", func() {
			nexusK.GetParamsFunc = func(ctx sdk.Context) types.Params {
				gateway = rand.AccAddr()

				params := types.DefaultParams()
				params.Gateway = gateway

				return params
			}
		}).
		Branch(
			When("the message has an asset", func() {
				msg = randMsg(nexus.Processing, true)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, route(ctx, nexus.RoutingContext{}, msg), "asset transfer is not supported")
				}),

			When("the message has no asset", func() {
				msg = randMsg(nexus.Processing)
			}).
				Then("should execute the wasm message", func(t *testing.T) {
					moduleAddr := rand.AccAddr()
					accountK.GetModuleAddressFunc = func(_ string) sdk.AccAddress { return moduleAddr }

					wasmK.ExecuteFunc = func(_ sdk.Context, _, _ sdk.AccAddress, _ []byte, _ sdk.Coins) ([]byte, error) {
						return nil, nil
					}

					assert.NoError(t, route(ctx, nexus.RoutingContext{}, msg))

					assert.Len(t, wasmK.ExecuteCalls(), 1)
					assert.Equal(t, wasmK.ExecuteCalls()[0].ContractAddress, gateway)
					assert.Equal(t, wasmK.ExecuteCalls()[0].Caller, moduleAddr)
					assert.Empty(t, wasmK.ExecuteCalls()[0].Coins)

					type req struct {
						RouteMessages []nexus.WasmMessage `json:"route_messages_from_nexus"`
					}

					var actual req
					assert.NoError(t, json.Unmarshal(wasmK.ExecuteCalls()[0].Msg, &actual))
					assert.Len(t, actual.RouteMessages, 1)
					assert.Equal(t, nexus.FromGeneralMessage(msg), actual.RouteMessages[0])
				}),
		).
		Run(t)

}
