package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func randPayload() []byte {
	bytesType := funcs.Must(abi.NewType("bytes", "bytes", nil))
	stringType := funcs.Must(abi.NewType("string", "string", nil))
	stringArrayType := funcs.Must(abi.NewType("string[]", "string[]", nil))

	argNum := int(rand.I64Between(1, 10))

	var args abi.Arguments
	for i := 0; i < argNum; i += 1 {
		args = append(args, abi.Argument{Type: stringType})
	}

	schema := abi.Arguments{{Type: stringType}, {Type: stringArrayType}, {Type: stringArrayType}, {Type: bytesType}}
	payload := funcs.Must(
		schema.Pack(
			rand.StrBetween(5, 10),
			slices.Expand2(func() string { return rand.Str(5) }, argNum),
			slices.Expand2(func() string { return "string" }, argNum),
			funcs.Must(args.Pack(slices.Expand2(func() interface{} { return "string" }, argNum)...)),
		),
	)

	return append(funcs.Must(hexutil.Decode(types.CosmWasmV1)), payload...)
}

func randMsg(status nexus.GeneralMessage_Status, payload []byte, token ...*sdk.Coin) nexus.GeneralMessage {
	var asset *sdk.Coin
	if len(token) > 0 {
		asset = token[0]
	}

	return nexus.GeneralMessage{
		ID: rand.NormalizedStr(10),
		Sender: nexus.CrossChainAddress{
			Chain:   nexustestutils.RandomChain(),
			Address: rand.NormalizedStr(42),
		},
		Recipient: nexus.CrossChainAddress{
			Chain:   nexustestutils.RandomChain(),
			Address: rand.NormalizedStr(42),
		},
		PayloadHash:   evmtestutils.RandomHash().Bytes(),
		Status:        status,
		Asset:         asset,
		SourceTxID:    evmtestutils.RandomHash().Bytes(),
		SourceTxIndex: uint64(rand.I64Between(0, 100)),
	}
}

func TestNewMessageRoute(t *testing.T) {
	var (
		ctx        sdk.Context
		routingCtx nexus.RoutingContext
		msg        nexus.GeneralMessage
		route      nexus.MessageRoute

		k         keeper.Keeper
		feegrantK *mock.FeegrantKeeperMock
		ibcK      *mock.IBCKeeperMock
		bankK     *mock.BankKeeperMock
		nexusK    *mock.NexusMock
		accountK  *mock.AccountKeeperMock
	)

	givenMessageRoute := Given("the message route", func() {
		ctx, k, _, feegrantK = setup()

		ibcK = &mock.IBCKeeperMock{}
		bankK = &mock.BankKeeperMock{}
		nexusK = &mock.NexusMock{}
		accountK = &mock.AccountKeeperMock{}

		route = keeper.NewMessageRoute(k, ibcK, feegrantK, bankK, nexusK, accountK)
	})

	givenMessageRoute.
		When("payload is nil", func() {
			routingCtx = nexus.RoutingContext{Payload: nil}
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, route(ctx, routingCtx, msg), "payload is required")
		}).
		Run(t)

	givenMessageRoute.
		When("the message cannot be translated", func() {
			routingCtx = nexus.RoutingContext{
				Sender:     rand.AccAddr(),
				FeeGranter: nil,
				Payload:    rand.Bytes(100),
			}
			msg = randMsg(nexus.Processing, routingCtx.Payload)
		}).
		Then("should return error", func(t *testing.T) {
			assert.ErrorContains(t, route(ctx, routingCtx, msg), "invalid payload")
		}).
		Run(t)

	whenTheMessageCanBeTranslated := When("the message can be translated", func() {
		routingCtx = nexus.RoutingContext{
			Sender:  rand.AccAddr(),
			Payload: randPayload(),
		}
	})

	givenMessageRoute.
		When2(whenTheMessageCanBeTranslated).
		When("the message has no token transfer", func() {
			msg = randMsg(nexus.Processing, routingCtx.Payload)
		}).
		Branch(
			When("the fee granter is not set", func() {
				routingCtx.FeeGranter = nil
			}).
				Then("should deduct the fee from the sender", func(t *testing.T) {
					bankK.SendCoinsFunc = func(_ sdk.Context, _, _ sdk.AccAddress, _ sdk.Coins) error { return nil }
					ibcK.SendMessageFunc = func(_ context.Context, _ nexus.CrossChainAddress, _ sdk.Coin, _, _ string) error {
						return nil
					}

					assert.NoError(t, route(ctx, routingCtx, msg))

					assert.Len(t, bankK.SendCoinsCalls(), 1)
					assert.Equal(t, routingCtx.Sender, bankK.SendCoinsCalls()[0].FromAddr)
					assert.Equal(t, types.AxelarGMPAccount, bankK.SendCoinsCalls()[0].ToAddr)
					assert.Equal(t, sdk.NewCoins(sdk.NewCoin(exported.NativeAsset, sdk.OneInt())), bankK.SendCoinsCalls()[0].Amt)

					assert.Len(t, ibcK.SendMessageCalls(), 1)
					assert.Equal(t, msg.Recipient, ibcK.SendMessageCalls()[0].Recipient)
					assert.Equal(t, sdk.NewCoin(exported.NativeAsset, sdk.OneInt()), ibcK.SendMessageCalls()[0].Asset)
					assert.Equal(t, msg.ID, ibcK.SendMessageCalls()[0].ID)
				}),

			When("the fee granter is set", func() {
				routingCtx.FeeGranter = rand.AccAddr()
			}).
				Then("should deduct the fee from the fee granter", func(t *testing.T) {
					feegrantK.UseGrantedFeesFunc = func(_ sdk.Context, granter, _ sdk.AccAddress, _ sdk.Coins, _ []sdk.Msg) error {
						return nil
					}
					bankK.SendCoinsFunc = func(_ sdk.Context, _, _ sdk.AccAddress, _ sdk.Coins) error { return nil }
					ibcK.SendMessageFunc = func(_ context.Context, _ nexus.CrossChainAddress, _ sdk.Coin, _, _ string) error {
						return nil
					}

					assert.NoError(t, route(ctx, routingCtx, msg))

					assert.Len(t, feegrantK.UseGrantedFeesCalls(), 1)
					assert.Equal(t, routingCtx.FeeGranter, feegrantK.UseGrantedFeesCalls()[0].Granter)
					assert.Equal(t, routingCtx.Sender, feegrantK.UseGrantedFeesCalls()[0].Grantee)
					assert.Equal(t, sdk.NewCoins(sdk.NewCoin(exported.NativeAsset, sdk.OneInt())), feegrantK.UseGrantedFeesCalls()[0].Fee)

					assert.Len(t, bankK.SendCoinsCalls(), 1)
					assert.Equal(t, routingCtx.FeeGranter, bankK.SendCoinsCalls()[0].FromAddr)
					assert.Equal(t, types.AxelarGMPAccount, bankK.SendCoinsCalls()[0].ToAddr)
					assert.Equal(t, sdk.NewCoins(sdk.NewCoin(exported.NativeAsset, sdk.OneInt())), bankK.SendCoinsCalls()[0].Amt)

					assert.Len(t, ibcK.SendMessageCalls(), 1)
					assert.Equal(t, msg.Recipient, ibcK.SendMessageCalls()[0].Recipient)
					assert.Equal(t, sdk.NewCoin(exported.NativeAsset, sdk.OneInt()), ibcK.SendMessageCalls()[0].Asset)
					assert.Equal(t, msg.ID, ibcK.SendMessageCalls()[0].ID)
				}),
		).
		Run(t)

	givenMessageRoute.
		When2(whenTheMessageCanBeTranslated).
		When("the message has token transfer", func() {
			coin := rand.Coin()
			msg = randMsg(nexus.Processing, routingCtx.Payload, &coin)
		}).
		Then("should deduct from the corresponding account", func(t *testing.T) {
			nexusK.GetChainByNativeAssetFunc = func(_ sdk.Context, _ string) (nexus.Chain, bool) {
				return exported.Axelarnet, true
			}
			bankK.SendCoinsFunc = func(_ sdk.Context, _, _ sdk.AccAddress, _ sdk.Coins) error { return nil }
			ibcK.SendMessageFunc = func(_ context.Context, _ nexus.CrossChainAddress, _ sdk.Coin, _, _ string) error {
				return nil
			}

			assert.NoError(t, route(ctx, routingCtx, msg))

			assert.Len(t, bankK.SendCoinsCalls(), 1)
			assert.Equal(t, types.GetEscrowAddress(msg.Asset.Denom), bankK.SendCoinsCalls()[0].FromAddr)
			assert.Equal(t, types.AxelarGMPAccount, bankK.SendCoinsCalls()[0].ToAddr)
			assert.Equal(t, sdk.NewCoins(*msg.Asset), bankK.SendCoinsCalls()[0].Amt)

			assert.Len(t, ibcK.SendMessageCalls(), 1)
			assert.Equal(t, msg.Recipient, ibcK.SendMessageCalls()[0].Recipient)
			assert.Equal(t, *msg.Asset, ibcK.SendMessageCalls()[0].Asset)
			assert.Equal(t, msg.ID, ibcK.SendMessageCalls()[0].ID)
		}).
		Run(t)
}
