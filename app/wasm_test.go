package app_test

import (
	"encoding/json"
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/app"
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
						Data:      nil,
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
	).Run(t)
}
