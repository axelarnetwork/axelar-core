package app_test

import (
	"encoding/json"
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
	"github.com/axelarnetwork/axelar-core/testutils/fake"
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
