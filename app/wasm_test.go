package app_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/mock"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexusmock "github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestAnteHandlerMessenger_DispatchMsg(t *testing.T) {
	var (
		messenger            wasmkeeper.Messenger
		antehandlerCalled    bool
		messagehandlerCalled bool
		err                  error
	)

	Given("an ante handler messenger", func() {
		antehandlerCalled = false
		messagehandlerCalled = false
		encoder := wasm.MessageEncoders{
			Bank:         func(_ sdk.AccAddress, _ *wasmvmtypes.BankMsg) ([]sdk.Msg, error) { return nil, nil },
			Custom:       func(_ sdk.AccAddress, _ json.RawMessage) ([]sdk.Msg, error) { return nil, nil },
			Distribution: func(_ sdk.AccAddress, _ *wasmvmtypes.DistributionMsg) ([]sdk.Msg, error) { return nil, nil },
			IBC: func(_ sdk.Context, _ sdk.AccAddress, _ string, _ *wasmvmtypes.IBCMsg) ([]sdk.Msg, error) {
				return nil, nil
			},
			Staking:  func(_ sdk.AccAddress, _ *wasmvmtypes.StakingMsg) ([]sdk.Msg, error) { return nil, nil },
			Stargate: func(_ sdk.AccAddress, _ *wasmvmtypes.StargateMsg) ([]sdk.Msg, error) { return nil, nil },
			Wasm:     func(_ sdk.AccAddress, _ *wasmvmtypes.WasmMsg) ([]sdk.Msg, error) { return nil, nil },
			Gov:      func(_ sdk.AccAddress, _ *wasmvmtypes.GovMsg) ([]sdk.Msg, error) { return nil, nil },
		}

		anteHandler := func(ctx sdk.Context, msgs []sdk.Msg, simulate bool) (sdk.Context, error) {
			antehandlerCalled = true
			return ctx, nil
		}
		messageHandler := wasmkeeper.MessageHandlerFunc(
			func(_ sdk.Context, _ sdk.AccAddress, _ string, _ wasmvmtypes.CosmosMsg) ([]sdk.Event, [][]byte, error) {
				messagehandlerCalled = true
				return nil, nil, nil
			})
		messenger = app.WithAnteHandlers(encoder, anteHandler, messageHandler)
	}).Branch(
		When("it dispatches an empty message", func() {
			_, _, err = messenger.DispatchMsg(sdk.Context{}, nil, "", wasmvmtypes.CosmosMsg{})
		}).
			Then("it should return an error", func(t *testing.T) {
				assert.Error(t, err)
			}),

		When("it dispatches multiple messages of different types", func() {
			_, _, err = messenger.DispatchMsg(sdk.Context{}, nil, "", wasmvmtypes.CosmosMsg{
				Bank: &wasmvmtypes.BankMsg{
					Burn: &wasmvmtypes.BurnMsg{
						Amount: wasmvmtypes.Coins{{Denom: "foo", Amount: "1"}},
					},
				},
				Staking: &wasmvmtypes.StakingMsg{
					Delegate: &wasmvmtypes.DelegateMsg{
						Validator: "validator",
						Amount:    wasmvmtypes.Coin{Denom: "foo", Amount: "1"},
					},
				},
			})
		}).
			Then("it should return an error", func(t *testing.T) {
				assert.Error(t, err)
			}),

		When("it dispatches multiple messages of the same type", func() {
			_, _, err = messenger.DispatchMsg(sdk.Context{}, nil, "", wasmvmtypes.CosmosMsg{
				Bank: &wasmvmtypes.BankMsg{
					Burn: &wasmvmtypes.BurnMsg{
						Amount: wasmvmtypes.Coins{{Denom: "foo", Amount: "1"}},
					},
					Send: &wasmvmtypes.SendMsg{
						ToAddress: "recipient",
						Amount:    wasmvmtypes.Coins{{Denom: "foo", Amount: "1"}},
					},
				},
			})
		}).
			Then("it should return an error", func(t *testing.T) {
				assert.Error(t, err)
			}),

		When("it dispatches a single message that is neither burn nor ibc send", func() {
			_, _, err = messenger.DispatchMsg(sdk.Context{}, nil, "", wasmvmtypes.CosmosMsg{
				Bank: &wasmvmtypes.BankMsg{
					Send: &wasmvmtypes.SendMsg{
						ToAddress: "recipient",
						Amount:    wasmvmtypes.Coins{{Denom: "foo", Amount: "1"}},
					},
				},
			})
		}).
			Then("antehandlers should get triggered", func(t *testing.T) {
				assert.True(t, antehandlerCalled)
			}).
			Then("messagehandlers should get triggered", func(t *testing.T) {
				assert.True(t, messagehandlerCalled)
			}).
			Then("the message should get dispatched without error", func(t *testing.T) {
				assert.NoError(t, err)
			}),

		When("it dispatches a single burn message", func() {
			_, _, err = messenger.DispatchMsg(sdk.Context{}, nil, "", wasmvmtypes.CosmosMsg{
				Bank: &wasmvmtypes.BankMsg{
					Burn: &wasmvmtypes.BurnMsg{
						Amount: wasmvmtypes.Coins{{Denom: "foo", Amount: "1"}},
					},
				},
			})
		}).
			Then("antehandlers should get skipped", func(t *testing.T) {
				assert.False(t, antehandlerCalled)
			}).
			Then("messagehandlers should get triggered", func(t *testing.T) {
				assert.True(t, messagehandlerCalled)
			}).
			Then("the message should get dispatched without error", func(t *testing.T) {
				assert.NoError(t, err)
			}),

		When("it dispatches a single ibc send message", func() {
			_, _, err = messenger.DispatchMsg(sdk.Context{}, nil, "", wasmvmtypes.CosmosMsg{
				IBC: &wasmvmtypes.IBCMsg{
					SendPacket: &wasmvmtypes.SendPacketMsg{
						ChannelID: "channel",
						Data:      []byte("data"),
						Timeout:   wasmvmtypes.IBCTimeout{},
					},
				},
			})
		}).
			Then("antehandlers should get skipped", func(t *testing.T) {
				assert.False(t, antehandlerCalled)
			}).
			Then("messagehandlers should get triggered", func(t *testing.T) {
				assert.True(t, messagehandlerCalled)
			}).
			Then("the message should get dispatched without error", func(t *testing.T) {
				assert.NoError(t, err)
			}),

		When("it dispatches a single custom message", func() {
			_, _, err = messenger.DispatchMsg(sdk.Context{}, nil, "", wasmvmtypes.CosmosMsg{
				Custom: json.RawMessage(`{"foo":"bar", "baz":1}`),
			})
		}).
			Then("antehandlers should get triggered", func(t *testing.T) {
				assert.True(t, antehandlerCalled)
			}).
			Then("messagehandlers should get triggered", func(t *testing.T) {
				assert.True(t, messagehandlerCalled)
			}).
			Then("the message should get dispatched without error", func(t *testing.T) {
				assert.NoError(t, err)
			}),

		When("it dispatches a single stargate message", func() {
			_, _, err = messenger.DispatchMsg(sdk.Context{}, nil, "",
				wasmvmtypes.CosmosMsg{Stargate: &wasmvmtypes.StargateMsg{
					TypeURL: "type",
					Value:   []byte("value"),
				}},
			)
		}).
			Then("antehandlers should get triggered", func(t *testing.T) {
				assert.True(t, antehandlerCalled)
			}).
			Then("messagehandlers should get triggered", func(t *testing.T) {
				assert.True(t, messagehandlerCalled)
			}).
			Then("the message should get dispatched without error", func(t *testing.T) {
				assert.NoError(t, err)
			}),
	).Run(t)
}

func TestMsgTypeBlacklistMessenger_DispatchMsg(t *testing.T) {
	var (
		messenger wasmkeeper.Messenger
		err       error
	)

	Given("a message handler with blacklisted message types", func() {
		messenger = app.NewMsgTypeBlacklistMessenger()
	}).Branch(
		When("it dispatches a stargate message", func() {
			_, _, err = messenger.DispatchMsg(sdk.Context{}, nil, "",
				wasmvmtypes.CosmosMsg{Stargate: &wasmvmtypes.StargateMsg{
					TypeURL: "type",
					Value:   []byte("value"),
				}},
			)
		}).
			Then("it should return an error that is not 'unknown msg'", func(t *testing.T) {
				assert.Error(t, err)
				assert.NotEqual(t, err, wasmtypes.ErrUnknownMsg)
			}),

		When("it dispatches a stargate message", func() {
			_, _, err = messenger.DispatchMsg(sdk.Context{}, nil, "",
				wasmvmtypes.CosmosMsg{IBC: &wasmvmtypes.IBCMsg{SendPacket: &wasmvmtypes.SendPacketMsg{
					ChannelID: "channel",
					Data:      []byte("data"),
					Timeout:   wasmvmtypes.IBCTimeout{},
				}}},
			)
		}).
			Then("it should return an error that is not 'unknown msg'", func(t *testing.T) {
				assert.Error(t, err)
				assert.NotEqual(t, err, wasmtypes.ErrUnknownMsg)
			}),
		When("it dispatches a custom message", func() {
			_, _, err = messenger.DispatchMsg(sdk.Context{}, nil, "",
				wasmvmtypes.CosmosMsg{Custom: json.RawMessage(`{"foo":"bar", "baz":1}`)},
			)
		}).
			Then("it should return an 'unknown msg' error", func(t *testing.T) {
				assert.Equal(t, err, wasmtypes.ErrUnknownMsg)
			}),
	).Run(t)
}

func TestNewWasmAppModuleBasicOverride(t *testing.T) {
	wasmModule := app.NewWasmAppModuleBasicOverride(wasm.AppModuleBasic{})
	cdc := app.MakeEncodingConfig().Codec

	genesis := wasmModule.DefaultGenesis(cdc)
	assert.NotEqual(t, genesis, wasm.AppModuleBasic{}.DefaultGenesis(cdc))

	var state wasm.GenesisState
	assert.NoError(t, cdc.UnmarshalJSON(genesis, &state))

	assert.Equal(t, state.Params.InstantiateDefaultPermission, wasmtypes.AccessTypeNobody)
	assert.True(t, state.Params.CodeUploadAccess.Equals(wasmtypes.AllowNobody))
}

func TestICSMiddleWare(t *testing.T) {
	keys := app.CreateStoreKeys()

	testCases := []struct {
		wasm  string
		hooks string
	}{
		{"false", "false"},
		{"true", "false"},
		{"true", "true"}}

	for _, testCase := range testCases {
		t.Run("wasm_enabled:"+testCase.wasm+"-hooks_enabled:"+testCase.hooks, func(t *testing.T) {
			app.WasmEnabled, app.IBCWasmHooksEnabled = testCase.wasm, testCase.hooks

			axelarApp := app.NewAxelarApp(
				log.TestingLogger(),
				dbm.NewMemDB(),
				nil,
				true,
				nil,
				"",
				"",
				0,
				app.MakeEncodingConfig(),
				simapp.EmptyAppOptions{},
				[]wasm.Option{},
			)

			// this is the focus of the test, we need to ensure that the hooks and wrapper are correctly set up for each valid wasm/hooks flag combination
			wasmHooks := app.InitWasmHooks(keys)
			ics4Wrapper := app.InitICS4Wrapper(axelarApp.Keepers, wasmHooks)

			ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
			packet := &mock.PacketIMock{
				ValidateBasicFunc:    func() error { return nil },
				GetSourcePortFunc:    func() string { return "source port" },
				GetSourceChannelFunc: func() string { return "source channel" },
				GetDestPortFunc:      func() string { return "destination port" },
				GetDestChannelFunc:   func() string { return "destination channel" },
			}

			// these must not panic and return an error unrelated to the wasm hook
			assert.ErrorContains(t, ics4Wrapper.SendPacket(ctx, nil, packet), "channel: channel not found")
			assert.ErrorContains(t, ics4Wrapper.WriteAcknowledgement(ctx, nil, packet, nil), "channel: channel not found")
			_, ok := ics4Wrapper.GetAppVersion(ctx, "port", "channel")
			assert.False(t, ok)
		})
	}
}

func TestMaxSizeOverrideForClient(t *testing.T) {
	msg := wasmtypes.MsgStoreCode{
		Sender:                rand.AccAddr().String(),
		WASMByteCode:          rand.Bytes(5000000),
		InstantiatePermission: nil,
	}

	assert.Error(t, msg.ValidateBasic())

	app.MaxWasmSize = "10000000"
	// ensure the override works no matter if it's server or client side
	app.WasmEnabled = "true"
	_, _ = cmd.NewRootCmd()

	// reset the sender, because the encoding has changed after calling the root cmd from prefix 'cosmos' to 'axelar'
	msg.Sender = rand.AccAddr().String()

	assert.Equal(t, 10000000, wasmtypes.MaxWasmSize)

	assert.NoError(t, msg.ValidateBasic())
}

func TestQueryPlugins(t *testing.T) {
	var (
		nexusK *nexusmock.NexusMock
		req    json.RawMessage
		ctx    sdk.Context
	)

	Given("the nexus keeper", func() {
		ctx = sdk.NewContext(nil, tmproto.Header{}, false, log.TestingLogger())
		nexusK = &nexusmock.NexusMock{}
	}).
		Branch(
			When("request is invalid", func() {
				req = []byte("{\"invalid\"}")
			}).
				Then("it should return an error", func(t *testing.T) {
					_, err := app.NewQueryPlugins(nexusK).Custom(ctx, req)

					assert.ErrorContains(t, err, "invalid Custom query request")
				}),

			When("request is unknown", func() {
				req = []byte("{\"unknown\":{}}")
			}).
				Then("it should return an error", func(t *testing.T) {
					_, err := app.NewQueryPlugins(nexusK).Custom(ctx, req)

					assert.ErrorContains(t, err, "unknown Custom query request")
				}),

			When("request is a nexus wasm query but unknown", func() {
				req = []byte("{\"nexus\":{}}")
			}).
				Then("it should return an error", func(t *testing.T) {
					_, err := app.NewQueryPlugins(nexusK).Custom(ctx, req)

					assert.ErrorContains(t, err, "unknown Nexus query request")
				}),

			When("request is a nexus wasm TxID query", func() {
				req = []byte("{\"nexus\":{\"tx_hash_and_nonce\":{}}}")
			}).
				Then("it should return a TxHashAndNonce response", func(t *testing.T) {
					txHash := [32]byte(rand.Bytes(32))
					index := uint64(rand.PosI64())
					nexusK.CurrIDFunc = func(ctx sdk.Context) ([32]byte, uint64) {
						return txHash, index
					}

					actual, err := app.NewQueryPlugins(nexusK).Custom(ctx, req)

					assert.NoError(t, err)
					assert.Equal(t, fmt.Sprintf("{\"tx_hash\":%s,\"nonce\":%d}", funcs.Must(json.Marshal(txHash)), index), string(actual))
				}),
			When("request is a nexus wasm IsChainRegistered query with empty chain name", func() {
				req = []byte("{\"nexus\":{\"is_chain_registered\":{\"chain\": \"\"}}}")
			}).
				Then("should fail validation", func(t *testing.T) {
					nexusK.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
						return nexus.Chain{}, true
					}

					_, err := app.NewQueryPlugins(nexusK).Custom(ctx, req)

					assert.ErrorContains(t, err, "invalid chain name")
				}),
			When("request is a nexus wasm IsChainRegistered query", func() {
				req = []byte("{\"nexus\":{\"is_chain_registered\":{\"chain\": \"chain-0\"}}}")
			}).
				Then("it should return a IsChainRegisteredResponse", func(t *testing.T) {
					nexusK.GetChainFunc = func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
						return nexus.Chain{}, true
					}
					actual, err := app.NewQueryPlugins(nexusK).Custom(ctx, req)

					assert.NoError(t, err)
					assert.Equal(t, "{\"is_registered\":true}", string(actual))
				}),
		).
		Run(t)

}
