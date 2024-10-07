package keeper_test

import (
	"errors"
	"fmt"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	evm "github.com/axelarnetwork/axelar-core/x/evm/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	exportedmock "github.com/axelarnetwork/axelar-core/x/nexus/exported/mock"
	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
	. "github.com/axelarnetwork/utils/test"
)

func TestMessenger_DispatchMsg(t *testing.T) {
	var (
		ctx       sdk.Context
		messenger keeper.Messenger
		nexus     *mock.NexusMock
		ibc       *mock.IBCKeeperMock
		bank      *mock.BankKeeperMock
		account   *mock.AccountKeeperMock
		msg       wasmvmtypes.CosmosMsg
	)

	contractAddr := rand.AccAddr()

	givenMessenger := Given("a messenger", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		nexus = &mock.NexusMock{
			LoggerFunc:                    func(ctx sdk.Context) log.Logger { return ctx.Logger() },
			IsWasmConnectionActivatedFunc: func(sdk.Context) bool { return true },
		}
		ibc = &mock.IBCKeeperMock{}
		bank = &mock.BankKeeperMock{}
		account = &mock.AccountKeeperMock{}

		messenger = keeper.NewMessenger(nexus, ibc, bank, account)
	})

	givenMessenger.
		When("the msg is encoded incorrectly", func() {
			msg = wasmvmtypes.CosmosMsg{
				Custom: []byte("[]"),
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
				Custom: []byte("{\"source_chain\":\"sourcechain\",\"source_address\":\"0xb860\",\"destination_chain\":\"Axelarnet\",\"destination_address\":\"axelarvaloper1zh9wrak6ke4n6fclj5e8yk397czv430ygs5jz7\",\"payload_hash\":[187,155,85,102,194,244,135,104,99,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241],\"source_tx_id\":[187,155,85,102,194,244,135,104,99,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241],\"source_tx_index\":100, \"id\": \"0x73657e3da2e404f474218fe2789462585d7f6060741bd312c862378cf67ca981-1\"}"),
			}
		}).
		Branch(
			When("the gateway is not set", func() {
				nexus.GetParamsFunc = func(_ sdk.Context) types.Params {
					return types.DefaultParams()
				}
			}).
				Then("should return error", func(t *testing.T) {
					_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

					assert.ErrorContains(t, err, "gateway is not set")
					assert.False(t, errors.Is(err, wasmtypes.ErrUnknownMsg))
				}),

			When("the gateway is set but given contract address does not match", func() {
				nexus.GetParamsFunc = func(_ sdk.Context) types.Params {
					params := types.DefaultParams()
					params.Gateway = rand.AccAddr()

					return params
				}
			}).
				Then("should return error", func(t *testing.T) {
					_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

					assert.ErrorContains(t, err, "is not the gateway")
					assert.False(t, errors.Is(err, wasmtypes.ErrUnknownMsg))
				}),
		).
		Run(t)

	givenMessenger.
		When("the msg is encoded correctly with token and the gateway is set correctly", func() {
			msg = wasmvmtypes.CosmosMsg{
				Custom: []byte("{\"source_chain\":\"sourcechain\",\"source_address\":\"0xb860\",\"destination_chain\":\"Axelarnet\",\"destination_address\":\"axelarvaloper1zh9wrak6ke4n6fclj5e8yk397czv430ygs5jz7\",\"payload_hash\":[187,155,85,102,194,244,135,104,99,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241],\"source_tx_id\":[187,155,85,102,194,244,135,104,99,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241],\"source_tx_index\":100, \"id\": \"0x73657e3da2e404f474218fe2789462585d7f6060741bd312c862378cf67ca981-1\",\"token\":{\"denom\":\"test\",\"amount\":\"100\"}}"),
			}

			nexus.GetParamsFunc = func(_ sdk.Context) types.Params {
				params := types.DefaultParams()
				params.Gateway = contractAddr

				return params
			}
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
			nexus.GetMessageFunc = func(_ sdk.Context, _ string) (exported.GeneralMessage, bool) {
				return exported.GeneralMessage{}, false
			}
		}).
		Branch(
			When("the asset is not registered for the destination chain", func() {
				nexus.IsAssetRegisteredFunc = func(_ sdk.Context, _ exported.Chain, _ string) bool { return false }
			}).
				Then("should return error", func(t *testing.T) {
					_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

					assert.ErrorContains(t, err, "is not registered on chain")
					assert.False(t, errors.Is(err, wasmtypes.ErrUnknownMsg))
				}),

			When("the asset is registered for the destination chain", func() {
				nexus.IsAssetRegisteredFunc = func(_ sdk.Context, _ exported.Chain, _ string) bool { return true }
			}).
				Then("should lock the coin and set new message", func(t *testing.T) {
					moduleAccount := rand.AccAddr()
					account.GetModuleAddressFunc = func(_ string) sdk.AccAddress { return moduleAccount }
					lockableAsset := &exportedmock.LockableAssetMock{}
					lockableAsset.LockFromFunc = func(ctx sdk.Context, fromAddr sdk.AccAddress) error { return nil }
					nexus.NewLockableAssetFunc = func(ctx sdk.Context, ibc types.IBCKeeper, bank types.BankKeeper, coin sdk.Coin) (exported.LockableAsset, error) {
						return lockableAsset, nil
					}
					nexus.SetNewMessageFunc = func(_ sdk.Context, msg exported.GeneralMessage) error {
						return msg.ValidateBasic()
					}
					nexus.RouteMessageFunc = func(ctx sdk.Context, id string, _ ...exported.RoutingContext) error { return nil }

					_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

					assert.NoError(t, err)
					assert.Len(t, lockableAsset.LockFromCalls(), 1)
					assert.Equal(t, lockableAsset.LockFromCalls()[0].FromAddr, moduleAccount)
					assert.Len(t, nexus.SetNewMessageCalls(), 1)
					assert.Len(t, nexus.RouteMessageCalls(), 1)
					assert.NotNil(t, nexus.SetNewMessageCalls()[0].Msg.Asset)
				}),
		).
		Run(t)

	givenMessenger.
		When("the msg is encoded correctly without token and the gateway is set correctly", func() {
			msg = wasmvmtypes.CosmosMsg{
				Custom: []byte("{\"source_chain\":\"sourcechain\",\"source_address\":\"0xb860\",\"destination_chain\":\"Axelarnet\",\"destination_address\":\"axelarvaloper1zh9wrak6ke4n6fclj5e8yk397czv430ygs5jz7\",\"payload_hash\":[187,155,85,102,194,244,135,104,99,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241],\"source_tx_id\":[187,155,85,102,194,244,135,104,99,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241],\"source_tx_index\":100, \"id\": \"0x73657e3da2e404f474218fe2789462585d7f6060741bd312c862378cf67ca981-1\",\"token\":null}"),
			}

			nexus.GetParamsFunc = func(_ sdk.Context) types.Params {
				params := types.DefaultParams()
				params.Gateway = contractAddr

				return params
			}
		}).
		Branch(
			When("the destination chain is not registered", func() {
				nexus.GetChainFunc = func(_ sdk.Context, chain exported.ChainName) (exported.Chain, bool) {
					return exported.Chain{}, false
				}
			}).
				Then("should return error", func(t *testing.T) {
					_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

					assert.ErrorContains(t, err, "is not a registered chain")
					assert.False(t, errors.Is(err, wasmtypes.ErrUnknownMsg))
				}),

			When("the destination chain is registered", func() {
				nexus.GetChainFunc = func(ctx sdk.Context, chain exported.ChainName) (exported.Chain, bool) { return exported.Chain{}, true }

			}).
				When("the msg fails to be set", func() {
					nexus.GetMessageFunc = func(_ sdk.Context, _ string) (exported.GeneralMessage, bool) {
						return exported.GeneralMessage{}, false
					}
					nexus.SetNewMessageFunc = func(_ sdk.Context, _ exported.GeneralMessage) error {
						return fmt.Errorf("set msg error")
					}
				}).
				Then("should return error", func(t *testing.T) {
					_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)

					assert.ErrorContains(t, err, "set msg error")
					assert.False(t, errors.Is(err, wasmtypes.ErrUnknownMsg))
				}),
		).
		Run(t)

	givenMessenger.
		When("the gateway is set correctly", func() {
			nexus.GetParamsFunc = func(_ sdk.Context) types.Params {
				params := types.DefaultParams()
				params.Gateway = contractAddr

				return params
			}
		}).
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
			nexus.GetMessageFunc = func(_ sdk.Context, _ string) (exported.GeneralMessage, bool) {
				return exported.GeneralMessage{}, false
			}
			nexus.SetNewMessageFunc = func(_ sdk.Context, msg exported.GeneralMessage) error {
				return msg.ValidateBasic()
			}
			nexus.RouteMessageFunc = func(ctx sdk.Context, id string, _ ...exported.RoutingContext) error { return nil }
		}).
		When("the msg is encoded correctly", func() {
			msg = wasmvmtypes.CosmosMsg{
				Custom: []byte("{\"source_chain\":\"sourcechain\",\"source_address\":\"0xb860\",\"destination_chain\":\"Axelarnet\",\"destination_address\":\"axelarvaloper1zh9wrak6ke4n6fclj5e8yk397czv430ygs5jz7\",\"payload_hash\":[187,155,85,102,194,244,135,104,99,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241],\"source_tx_id\":[187,155,85,102,194,244,135,104,99,51,62,72,31,70,152,53,1,84,37,159,254,98,38,226,131,177,108,225,138,100,188,241],\"source_tx_index\":100, \"id\": \"0x73657e3da2e404f474218fe2789462585d7f6060741bd312c862378cf67ca981-1\"}"),
			}
		}).
		Branch(
			When("succeed to route message", func() {
				nexus.RouteMessageFunc = func(_ sdk.Context, id string, _ ...exported.RoutingContext) error { return nil }
			}).
				Branch(
					Then("should route the message", func(t *testing.T) {
						_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)
						assert.NoError(t, err)

						assert.Len(t, nexus.SetNewMessageCalls(), 1)
						assert.Equal(t, nexus.SetNewMessageCalls()[0].Msg.Recipient.Chain, axelarnet.Axelarnet)
						assert.Equal(t, nexus.SetNewMessageCalls()[0].Msg.Status, exported.Approved)
						assert.Nil(t, nexus.SetNewMessageCalls()[0].Msg.Asset)

						assert.Len(t, nexus.RouteMessageCalls(), 1)
						assert.Equal(t, nexus.SetNewMessageCalls()[0].Msg.ID, nexus.RouteMessageCalls()[0].ID)
					}),
					When("message already set", func() {

						nexus.GetMessageFunc = func(_ sdk.Context, _ string) (exported.GeneralMessage, bool) {
							return exported.GeneralMessage{}, true
						}
					}).Then("should be a no op", func(t *testing.T) {

						_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)
						assert.NoError(t, err)

						assert.Len(t, nexus.SetNewMessageCalls(), 0)

						assert.Len(t, nexus.RouteMessageCalls(), 0)
					}),

					When("failed to route message", func() {
						nexus.RouteMessageFunc = func(_ sdk.Context, id string, _ ...exported.RoutingContext) error { return fmt.Errorf("failed") }
					}).
						Then("should set message as processing", func(t *testing.T) {
							_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", msg)
							assert.NoError(t, err)

							assert.Len(t, nexus.SetNewMessageCalls(), 1)
							assert.Equal(t, nexus.SetNewMessageCalls()[0].Msg.Recipient.Chain, axelarnet.Axelarnet)
							assert.Equal(t, nexus.SetNewMessageCalls()[0].Msg.Status, exported.Approved)
							assert.Nil(t, nexus.SetNewMessageCalls()[0].Msg.Asset)

							assert.Len(t, nexus.RouteMessageCalls(), 1)
							assert.Equal(t, nexus.SetNewMessageCalls()[0].Msg.ID, nexus.RouteMessageCalls()[0].ID)
						}),
				),
		).
		Run(t)

}

func TestMessenger_DispatchMsg_WasmConnectionNotActivated(t *testing.T) {
	var (
		ctx       sdk.Context
		messenger keeper.Messenger
		nexus     *mock.NexusMock
	)

	contractAddr := rand.AccAddr()

	Given("a messenger", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
		nexus = &mock.NexusMock{
			LoggerFunc: func(ctx sdk.Context) log.Logger { return ctx.Logger() },
		}
		messenger = keeper.NewMessenger(nexus, &mock.IBCKeeperMock{}, &mock.BankKeeperMock{}, &mock.AccountKeeperMock{})
	}).
		When("wasm connection is not activated", func() {
			nexus.IsWasmConnectionActivatedFunc = func(_ sdk.Context) bool { return false }
		}).
		Then("should return error", func(t *testing.T) {
			_, _, err := messenger.DispatchMsg(ctx, contractAddr, "", wasmvmtypes.CosmosMsg{})

			assert.ErrorContains(t, err, "wasm connection is not activated")
		}).
		Run(t)
}
