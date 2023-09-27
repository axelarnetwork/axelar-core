package keeper_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestMessenger_DispatchMsg(t *testing.T) {
	var (
		ctx       sdk.Context
		messenger keeper.Messenger
		nexus     *mock.NexusMock
		msg       wasmvmtypes.CosmosMsg
	)

	contractAddr := rand.AccAddr()

	givenMessenger := Given("a messenger", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		nexus = &mock.NexusMock{LoggerFunc: func(_ sdk.Context) log.Logger { return log.TestingLogger() }}
		messenger = keeper.NewMessenger(nexus)
	})

	givenMessenger.
		When("the msg is encoded incorrectly", func() {
			msg = wasmvmtypes.CosmosMsg{
				Custom: []byte("{}"),
			}
		}).
		Then("should return error", func(t *testing.T) {
			_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

			assert.Error(t, err)
			assert.True(t, errors.Is(err, wasmtypes.ErrUnknownMsg))
		}).
		Run(t)

	givenMessenger.
		When("the msg is encoded correctly", func() {
			msg = wasmvmtypes.CosmosMsg{
				Custom: []byte("[]"),
			}
		}).
		Branch(
			When("the connection router is not set", func() {
				nexus.GetParamsFunc = func(_ sdk.Context) types.Params {
					return types.DefaultParams()
				}
			}).
				Then("should return error", func(t *testing.T) {
					_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

					assert.ErrorContains(t, err, "connection router is not set")
					assert.False(t, errors.Is(err, wasmtypes.ErrUnknownMsg))
				}),

			When("the connection router is set but given contract address does not match", func() {
				nexus.GetParamsFunc = func(_ sdk.Context) types.Params {
					params := types.DefaultParams()
					params.ConnectionRouter = rand.AccAddr()

					return params
				}
			}).
				Then("should return error", func(t *testing.T) {
					_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

					assert.ErrorContains(t, err, "is not the connection router")
					assert.False(t, errors.Is(err, wasmtypes.ErrUnknownMsg))
				}),
		).
		Run(t)

	givenMessenger.
		When("the msg is encoded correctly and the connection router is set correctly", func() {
			msg = wasmvmtypes.CosmosMsg{
				Custom: []byte(`[
					{"sender_chain":"SomeChain","sender_address":"SomeAddress","recipient_chain":"Ethereum","recipient_address":"0xDAFEA492D9c6733ae3d56b7Ed1ADB60692c98Bc5","payload_hash":"pOcUbpJ7WCC/TIx2sJA/qm0gZGSvDvXgK9QagbH4E2w=","source_tx_id":"0ITBsic95Pt5EqsbMyKO04iW/74srqpDnPMzthkCM6w=","source_tx_index":0},
					{"sender_chain":"SomeChain","sender_address":"SomeAddress","recipient_chain":"Axelarnet","recipient_address":"axelarvaloper1zh9wrak6ke4n6fclj5e8yk397czv430ygs5jz7","payload_hash":"XZx9n7ycI4EWhVo411N4PVWPconX0CPuNfVvKDLMSOQ=","source_tx_id":"jvJHwR7yyDhI53dnhELdJj5ZUDO/FJovyCjamgOQ5Xk=","source_tx_index":100}
				]`),
			}

			nexus.GetParamsFunc = func(_ sdk.Context) types.Params {
				params := types.DefaultParams()
				params.ConnectionRouter = contractAddr

				return params
			}
		}).
		Branch(
			When("the destination chain is not registered", func() {
				nexus.GetChainFunc = func(_ sdk.Context, chain exported.ChainName) (exported.Chain, bool) {
					return exported.Chain{}, false
				}
			}).
				Then("should not do anything", func(t *testing.T) {
					_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

					assert.NoError(t, err)
					assert.Len(t, slices.Filter(ctx.EventManager().Events(), func(ev sdk.Event) bool {
						return ev.Type == "axelar.nexus.v1beta1.ConnectionRouterMessageReceived" && slices.Any(ev.Attributes, func(attr abci.EventAttribute) bool {
							return bytes.Equal(attr.Key, []byte("routed")) && bytes.Equal(attr.Value, []byte("false"))
						})
					}), 2)
				}),

			When("the destination chain is registered", func() {
				nexus.GetChainFunc = func(_ sdk.Context, chain exported.ChainName) (exported.Chain, bool) {
					switch chain {
					case evm.Ethereum.Name:
						return evm.Ethereum, true
					case axelarnet.Axelarnet.Name:
						return axelarnet.Axelarnet, true
					default:
						return exported.Chain{}, false
					}
				}
			}).
				When("the msg fails to be set", func() {
					nexus.GenerateMessageIDFunc = func(_ sdk.Context) (string, []byte, uint64) {
						return "1", []byte("1"), 1
					}
					nexus.SetNewMessageFromWasmFunc = func(_ sdk.Context, _ exported.GeneralMessage) error {
						return fmt.Errorf("set msg error")
					}
				}).
				Then("should do nothing", func(t *testing.T) {
					_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

					assert.NoError(t, err)
					assert.Len(t, slices.Filter(ctx.EventManager().Events(), func(ev sdk.Event) bool {
						return ev.Type == "axelar.nexus.v1beta1.ConnectionRouterMessageReceived" && slices.Any(ev.Attributes, func(attr abci.EventAttribute) bool {
							return bytes.Equal(attr.Key, []byte("routed")) && bytes.Equal(attr.Value, []byte("false"))
						})
					}), 2)
				}),

			When("the destination chain is registered", func() {
				nexus.GetChainFunc = func(_ sdk.Context, chain exported.ChainName) (exported.Chain, bool) {
					switch chain {
					case evm.Ethereum.Name:
						return evm.Ethereum, true
					case axelarnet.Axelarnet.Name:
						return axelarnet.Axelarnet, true
					default:
						return exported.Chain{}, false
					}
				}
			}).
				When("the msg succeeds to be set", func() {
					nexus.GenerateMessageIDFunc = func(_ sdk.Context) (string, []byte, uint64) {
						return "1", []byte("1"), 1
					}
					nexus.SetNewMessageFromWasmFunc = func(_ sdk.Context, msg exported.GeneralMessage) error {
						return nil
					}
				}).
				Then("should return error", func(t *testing.T) {
					_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

					assert.NoError(t, err)
					assert.False(t, errors.Is(err, wasmtypes.ErrUnknownMsg))

					assert.Len(t, nexus.SetNewMessageFromWasmCalls(), 2)
					assert.Equal(t, nexus.SetNewMessageFromWasmCalls()[0].Msg.Recipient.Chain, evm.Ethereum)
					assert.Equal(t, nexus.SetNewMessageFromWasmCalls()[0].Msg.Status, exported.Processing)
					assert.Nil(t, nexus.SetNewMessageFromWasmCalls()[0].Msg.Asset)
					assert.Equal(t, nexus.SetNewMessageFromWasmCalls()[1].Msg.Recipient.Chain, axelarnet.Axelarnet)
					assert.Equal(t, nexus.SetNewMessageFromWasmCalls()[1].Msg.Status, exported.Approved)
					assert.Nil(t, nexus.SetNewMessageFromWasmCalls()[1].Msg.Asset)
					assert.Len(t, slices.Filter(ctx.EventManager().Events(), func(ev sdk.Event) bool {
						return ev.Type == "axelar.nexus.v1beta1.ConnectionRouterMessageReceived" && slices.Any(ev.Attributes, func(attr abci.EventAttribute) bool {
							return bytes.Equal(attr.Key, []byte("routed")) && bytes.Equal(attr.Value, []byte("true"))
						})
					}), 2)
				}),
		).
		Run(t)
}
