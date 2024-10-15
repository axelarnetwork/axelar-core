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
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

type req struct {
	RouteMessages []exported.WasmMessage `json:"route_messages_from_nexus"`
}

func TestNewMessageRoute(t *testing.T) {
	var (
		ctx   sdk.Context
		route exported.MessageRoute
		msg   exported.GeneralMessage

		nexusK   *mock.NexusMock
		accountK *mock.AccountKeeperMock
		wasmK    *mock.WasmKeeperMock
		gateway  sdk.AccAddress
	)

	givenMessageRoute := Given("the message route", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

		nexusK = &mock.NexusMock{}
		nexusK.IsWasmConnectionActivatedFunc = func(_ sdk.Context) bool { return true }
		accountK = &mock.AccountKeeperMock{}
		wasmK = &mock.WasmKeeperMock{}

		route = keeper.NewMessageRoute(nexusK, accountK, wasmK)
	})

	givenMessageRoute.
		When("the gateway is not set", func() {
			nexusK.GetParamsFunc = func(ctx sdk.Context) types.Params { return types.DefaultParams() }
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, route(ctx, exported.RoutingContext{}, msg), "gateway is not set")
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
			nexusK.SetMessageExecutedFunc = func(_ sdk.Context, _ string) error { return nil }
		}).
		Branch(
			When("the message has an asset", func() {
				msg = randMsg(exported.Processing, true)
			}).
				Then("should return error", func(t *testing.T) {
					assert.ErrorContains(t, route(ctx, exported.RoutingContext{}, msg), "asset transfer is not supported")
				}),

			When("the message has no asset", func() {
				msg = randMsg(exported.Processing)
			}).
				Then("should execute the wasm message", func(t *testing.T) {
					moduleAddr := rand.AccAddr()
					accountK.GetModuleAddressFunc = func(_ string) sdk.AccAddress { return moduleAddr }

					wasmK.ExecuteFunc = func(_ sdk.Context, _, _ sdk.AccAddress, _ []byte, _ sdk.Coins) ([]byte, error) {
						return nil, nil
					}

					assert.NoError(t, route(ctx, exported.RoutingContext{}, msg))

					assert.Len(t, wasmK.ExecuteCalls(), 1)
					assert.Equal(t, wasmK.ExecuteCalls()[0].ContractAddress, gateway)
					assert.Equal(t, wasmK.ExecuteCalls()[0].Caller, moduleAddr)
					assert.Empty(t, wasmK.ExecuteCalls()[0].Coins)

					var actual req
					assert.NoError(t, json.Unmarshal(wasmK.ExecuteCalls()[0].Msg, &actual))
					assert.Len(t, actual.RouteMessages, 1)
					assert.Equal(t, exported.FromGeneralMessage(msg), actual.RouteMessages[0])

					assert.Equal(t, len(nexusK.SetMessageExecutedCalls()), 1)
				}),
		).
		Run(t)
}

func TestMessageRoute_WasmConnectionNotActivated(t *testing.T) {
	var (
		ctx      sdk.Context
		route    exported.MessageRoute
		nexusK   *mock.NexusMock
		accountK *mock.AccountKeeperMock
		wasmK    *mock.WasmKeeperMock
	)

	Given("the message route", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

		nexusK = &mock.NexusMock{}

		nexusK.GetParamsFunc = func(ctx sdk.Context) types.Params {
			params := types.DefaultParams()
			params.Gateway = rand.AccAddr()

			return params
		}
		nexusK.SetMessageExecutedFunc = func(_ sdk.Context, _ string) error { return nil }

		accountK = &mock.AccountKeeperMock{}
		accountK.GetModuleAddressFunc = func(_ string) sdk.AccAddress { return rand.AccAddr() }
		wasmK = &mock.WasmKeeperMock{}
		wasmK.ExecuteFunc = func(_ sdk.Context, _, _ sdk.AccAddress, _ []byte, _ sdk.Coins) ([]byte, error) { return nil, nil }

		route = keeper.NewMessageRoute(nexusK, accountK, wasmK)
	}).
		When("the wasm connection is not activated", func() {
			nexusK.IsWasmConnectionActivatedFunc = func(_ sdk.Context) bool { return false }
		}).
		Then("should return error", func(t *testing.T) {
			err := route(ctx, exported.RoutingContext{}, exported.GeneralMessage{
				ID:            "id",
				Sender:        testutils.RandomCrossChainAddress(),
				Recipient:     testutils.RandomCrossChainAddress(),
				PayloadHash:   rand.Bytes(32),
				SourceTxID:    rand.Bytes(32),
				SourceTxIndex: 0,
			})
			assert.ErrorContains(t, err, "wasm connection is not activated")
		}).
		Run(t)
}

func TestReq_MarshalUnmarshalJSON(t *testing.T) {
	bz := []byte("{\"route_messages_from_nexus\":[{\"source_chain\":\"sourcechain\",\"source_address\":\"0xb860\",\"destination_chain\":\"destinationchain\",\"destination_address\":\"0xD419\",\"payload_hash\":[187,155,85,102,194,244,135,104,99,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241],\"source_tx_id\":[47,228],\"source_tx_index\":100},{\"source_chain\":\"sourcechain\",\"source_address\":\"0xc860\",\"destination_chain\":\"destinationchain\",\"destination_address\":\"0xA419\",\"payload_hash\":[203,155,85,102,194,244,135,104,83,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241],\"source_tx_id\":[35,244],\"source_tx_index\":0}]}")

	var actual req
	err := json.Unmarshal(bz, &actual)

	assert.NoError(t, err)
	assert.Equal(t, bz, funcs.Must(json.Marshal(actual)))
}
