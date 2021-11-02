package keeper_test

import (
	"crypto/sha256"
	"fmt"
	mathRand "math/rand"
	"testing"
	"time"

	ibctypes "github.com/cosmos/ibc-go/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/modules/core/23-commitment/types"
	ibcclient "github.com/cosmos/ibc-go/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/modules/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/ibc-go/testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

const (
	testChain = "cosmoschain-0"
	testToken = "stake"
)

func TestHandleMsgLink(t *testing.T) {
	var (
		server      types.MsgServiceServer
		nexusKeeper *mock.NexusMock
		ctx         sdk.Context
		msg         *types.LinkRequest
	)
	setup := func() {
		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					NativeAsset:           rand.StrBetween(5, 20),
					SupportsForeignAssets: true,
				}, true
			},
			IsAssetRegisteredFunc: func(sdk.Context, string, string) bool { return true },
			LinkAddressesFunc:     func(sdk.Context, nexus.CrossChainAddress, nexus.CrossChainAddress) {},
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		rtr := baseapp.NewRouter()
		msgServiceRtr := baseapp.NewMsgServiceRouter()
		server = keeper.NewMsgServerImpl(&mock.BaseKeeperMock{}, nexusKeeper, &mock.BankKeeperMock{}, &mock.IBCTransferKeeperMock{}, &mock.ChannelKeeperMock{}, &mock.AccountKeeperMock{}, msgServiceRtr, rtr)
	}

	repeatCount := 20
	t.Run("should return the linked address when the given chain and asset is registered", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgLink()
		_, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, nexusKeeper.LinkAddressesCalls(), 1)
		assert.Equal(t, msg.RecipientChain, nexusKeeper.GetChainCalls()[0].Chain)
	}).Repeat(repeatCount))

	t.Run("should return error when the given chain is unknown", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgLink()
		nexusKeeper.GetChainFunc = func(sdk.Context, string) (nexus.Chain, bool) { return nexus.Chain{}, false }
		_, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should return error when the given asset is not registered", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgLink()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		_, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))
}

func TestHandleMsgConfirmDeposit(t *testing.T) {
	var (
		server          types.MsgServiceServer
		axelarnetKeeper *mock.BaseKeeperMock
		nexusKeeper     *mock.NexusMock
		bankKeeper      *mock.BankKeeperMock
		transferKeeper  *mock.IBCTransferKeeperMock
		ctx             sdk.Context
		msg             *types.ConfirmDepositRequest
	)
	setup := func() {
		ibcPath := randomIBCPath()
		axelarnetKeeper = &mock.BaseKeeperMock{
			GetIBCPathFunc: func(sdk.Context, string) (string, bool) {
				return ibcPath, true
			},
			GetCosmosChainFunc: func(sdk.Context, string) (string, bool) {
				return "cosmoshub", true
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					NativeAsset:           rand.StrBetween(5, 20),
					SupportsForeignAssets: true,
				}, true
			},
			IsAssetRegisteredFunc:  func(sdk.Context, string, string) bool { return true },
			EnqueueForTransferFunc: func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) error { return nil },
			AddToChainTotalFunc:    func(_ sdk.Context, _ nexus.Chain, _ sdk.Coin) {},
		}
		bankKeeper = &mock.BankKeeperMock{
			BurnCoinsFunc:                    func(sdk.Context, string, sdk.Coins) error { return nil },
			SendCoinsFromAccountToModuleFunc: func(sdk.Context, sdk.AccAddress, string, sdk.Coins) error { return nil },
			SendCoinsFunc:                    func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error { return nil },
		}
		transferKeeper = &mock.IBCTransferKeeperMock{
			GetDenomTraceFunc: func(sdk.Context, tmbytes.HexBytes) (ibctypes.DenomTrace, bool) {
				return ibctypes.DenomTrace{
					Path:      ibcPath,
					BaseDenom: randomDenom(),
				}, true
			},
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		rtr := baseapp.NewRouter()
		msgServiceRtr := baseapp.NewMsgServiceRouter()
		server = keeper.NewMsgServerImpl(axelarnetKeeper, nexusKeeper, bankKeeper, transferKeeper, &mock.ChannelKeeperMock{}, &mock.AccountKeeperMock{}, msgServiceRtr, rtr)
	}

	repeatCount := 20
	t.Run("should enqueue Transfer in nexus keeper when registered tokens are sent from burner address to bank keeper, and burned", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgConfirmDeposit()
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		events := ctx.EventManager().ABCIEvents()
		assert.NoError(t, err)
		assert.Len(t, testutils.Events(events).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeDepositConfirmation }), 1)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 1)
		assert.Len(t, bankKeeper.SendCoinsFromAccountToModuleCalls(), 1)
		assert.Len(t, bankKeeper.BurnCoinsCalls(), 1)
		assert.Equal(t, msg.Token.Denom, nexusKeeper.EnqueueForTransferCalls()[0].Amount.Denom)
		assert.Equal(t, msg.Token.Amount, nexusKeeper.EnqueueForTransferCalls()[0].Amount.Amount)
	}).Repeat(repeatCount))

	t.Run("should return error when EnqueueForTransfer in nexus keeper failed", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgConfirmDeposit()
		nexusKeeper.EnqueueForTransferFunc = func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) error {
			return fmt.Errorf("failed")
		}

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should panic when BurnCoins in bank keeper failed", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgConfirmDeposit()
		bankKeeper.BurnCoinsFunc = func(sdk.Context, string, sdk.Coins) error {
			return fmt.Errorf("failed")
		}

		assert.Panics(t, func() { _, _ = server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg) }, "ConfirmDeposit did not panic when burn token failed")
	}).Repeat(repeatCount))

	t.Run("should return error when SendCoinsFromAccountToModule in bank keeper failed", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgConfirmDeposit()
		bankKeeper.SendCoinsFromAccountToModuleFunc = func(sdk.Context, sdk.AccAddress, string, sdk.Coins) error {
			return fmt.Errorf("failed")
		}
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should enqueue Transfer in nexus keeper when registered ICS20 tokens are sent from burner address to escrow address", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		msg = randomMsgConfirmDeposit()
		msg.Token.Denom = randomIBCDenom()
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		events := ctx.EventManager().ABCIEvents()
		assert.NoError(t, err)
		assert.Len(t, testutils.Events(events).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeDepositConfirmation }), 1)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 1)
		assert.Len(t, nexusKeeper.AddToChainTotalCalls(), 1)
		assert.Len(t, bankKeeper.SendCoinsCalls(), 1)
		assert.Equal(t, msg.Token.Amount, nexusKeeper.EnqueueForTransferCalls()[0].Amount.Amount)
	}).Repeat(repeatCount))

	t.Run("should return error when ICS20 token hash not found in IBCTransferKeeper", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		transferKeeper.GetDenomTraceFunc = func(sdk.Context, tmbytes.HexBytes) (ibctypes.DenomTrace, bool) {
			return ibctypes.DenomTrace{}, false
		}
		msg = randomMsgConfirmDeposit()
		msg.Token.Denom = randomIBCDenom()
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should return error when ICS20 token path not registered in axelarnet keeper", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		axelarnetKeeper.GetIBCPathFunc = func(sdk.Context, string) (string, bool) {
			return "", false
		}
		msg = randomMsgConfirmDeposit()
		msg.Token.Denom = randomIBCDenom()
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should return error when ICS20 token tracing path does not match registered path in axelarnet keeper", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		axelarnetKeeper.GetIBCPathFunc = func(sdk.Context, string) (string, bool) {
			return randomIBCPath(), true
		}
		msg = randomMsgConfirmDeposit()
		msg.Token.Denom = randomIBCDenom()
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should return error when SendCoins in bank keeper failed", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		bankKeeper.SendCoinsFunc = func(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error {
			return fmt.Errorf("failed")
		}
		msg = randomMsgConfirmDeposit()
		msg.Token.Denom = randomIBCDenom()
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should enqueue Transfer in nexus keeper when native axelar tokens are sent from burner address to escrow address", testutils.Func(func(t *testing.T) {
		setup()

		msg = randomMsgConfirmDeposit()
		msg.Token.Denom = exported.Axelarnet.NativeAsset
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		events := ctx.EventManager().ABCIEvents()
		assert.NoError(t, err)
		assert.Len(t, testutils.Events(events).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeDepositConfirmation }), 1)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 1)
		assert.Len(t, bankKeeper.SendCoinsCalls(), 1)
		assert.Equal(t, msg.Token.Denom, nexusKeeper.EnqueueForTransferCalls()[0].Amount.Denom)
		assert.Equal(t, msg.Token.Amount, nexusKeeper.EnqueueForTransferCalls()[0].Amount.Amount)
	}).Repeat(repeatCount))

	t.Run("should return error when asset is not a valid IBC denom and not registered", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		msg = randomMsgConfirmDeposit()
		msg.Token.Denom = "ibc" + randomDenom()
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)

		assert.Error(t, err)
	}).Repeat(repeatCount))
}

func TestHandleMsgExecutePendingTransfers(t *testing.T) {
	var (
		server          types.MsgServiceServer
		axelarnetKeeper *mock.BaseKeeperMock
		nexusKeeper     *mock.NexusMock
		bankKeeper      *mock.BankKeeperMock
		accountKeeper   *mock.AccountKeeperMock
		ctx             sdk.Context
		msg             *types.ExecutePendingTransfersRequest

		transfers []nexus.CrossChainTransfer
	)
	setup := func() {
		axelarnetKeeper = &mock.BaseKeeperMock{
			GetIBCPathFunc: func(sdk.Context, string) (string, bool) {
				return "", false
			},
			GetCosmosChainFunc: func(sdk.Context, string) (string, bool) {
				return testChain, true
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetTransfersForChainFunc: func(sdk.Context, nexus.Chain, nexus.TransferState) []nexus.CrossChainTransfer {
				transfers = []nexus.CrossChainTransfer{}
				for i := int64(0); i < rand.I64Between(1, 50); i++ {
					transfer := randomTransfer(testToken, testChain)
					transfers = append(transfers, transfer)
				}
				return transfers
			},
			ArchivePendingTransferFunc: func(sdk.Context, nexus.CrossChainTransfer) {},
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					NativeAsset:           randomDenom(),
					SupportsForeignAssets: true,
				}, true
			},
			IsAssetRegisteredFunc:  func(sdk.Context, string, string) bool { return true },
			EnqueueForTransferFunc: func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) error { return nil },
		}
		bankKeeper = &mock.BankKeeperMock{
			MintCoinsFunc: func(sdk.Context, string, sdk.Coins) error { return nil },
			SendCoinsFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error { return nil },
		}
		accountKeeper = &mock.AccountKeeperMock{
			GetModuleAddressFunc: func(string) sdk.AccAddress { return rand.AccAddr() },
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		rtr := baseapp.NewRouter()
		msgServiceRtr := baseapp.NewMsgServiceRouter()
		server = keeper.NewMsgServerImpl(axelarnetKeeper, nexusKeeper, bankKeeper, &mock.IBCTransferKeeperMock{}, &mock.ChannelKeeperMock{}, accountKeeper, msgServiceRtr, rtr)
	}

	repeatCount := 20
	t.Run("should mint and send token to recipients, and archive pending transfers when get pending transfers from nexus keeper ", testutils.Func(func(t *testing.T) {
		setup()
		msg = types.NewExecutePendingTransfersRequest(rand.AccAddr())
		_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, bankKeeper.MintCoinsCalls(), len(transfers))
		assert.Len(t, bankKeeper.SendCoinsCalls(), len(transfers))
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
	}).Repeat(repeatCount))

	t.Run("should return error when MintCoins in bank keeper failed", testutils.Func(func(t *testing.T) {
		setup()
		bankKeeper.MintCoinsFunc = func(sdk.Context, string, sdk.Coins) error {
			return fmt.Errorf("failed")
		}
		msg = types.NewExecutePendingTransfersRequest(rand.AccAddr())
		_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should send ICS20 token from escrow account to recipients, and archive pending transfers \\"+
		"when pending transfer asset is origined from cosmos chain", testutils.Func(func(t *testing.T) {
		setup()
		axelarnetKeeper.GetIBCPathFunc = func(sdk.Context, string) (string, bool) {
			return randomIBCPath(), true
		}

		msg = types.NewExecutePendingTransfersRequest(rand.AccAddr())
		_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, bankKeeper.SendCoinsCalls(), len(transfers))
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
	}).Repeat(repeatCount))

	t.Run("should send axelar native token from escrow account to recipients, and archive pending transfers \\"+
		"when pending transfer asset is axelar native token", testutils.Func(func(t *testing.T) {
		setup()
		transfers = []nexus.CrossChainTransfer{}
		for i := int64(0); i < rand.I64Between(1, 50); i++ {
			transfer := randomTransfer(exported.Axelarnet.NativeAsset, testChain)
			transfers = append(transfers, transfer)
		}
		nexusKeeper.GetTransfersForChainFunc = func(sdk.Context, nexus.Chain, nexus.TransferState) []nexus.CrossChainTransfer {
			return transfers
		}
		msg = types.NewExecutePendingTransfersRequest(rand.AccAddr())
		_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, bankKeeper.SendCoinsCalls(), len(transfers))
		assert.Len(t, bankKeeper.MintCoinsCalls(), 0)
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
	}).Repeat(repeatCount))
}

func TestHandleMsgRegisterIBCPath(t *testing.T) {
	var (
		server          types.MsgServiceServer
		axelarnetKeeper *mock.BaseKeeperMock
		ctx             sdk.Context
		msg             *types.RegisterIBCPathRequest
	)
	setup := func() {
		axelarnetKeeper = &mock.BaseKeeperMock{
			RegisterIBCPathFunc: func(sdk.Context, string, string) error { return nil },
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		rtr := baseapp.NewRouter()
		msgServiceRtr := baseapp.NewMsgServiceRouter()

		server = keeper.NewMsgServerImpl(axelarnetKeeper, &mock.NexusMock{}, &mock.BankKeeperMock{}, &mock.IBCTransferKeeperMock{}, &mock.ChannelKeeperMock{}, &mock.AccountKeeperMock{}, msgServiceRtr, rtr)
	}

	repeatCount := 20
	t.Run("should register an IBC tracing path for an chain if not registered yet", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgRegisterIBCPath()
		_, err := server.RegisterIBCPath(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, axelarnetKeeper.RegisterIBCPathCalls(), 1)
	}).Repeat(repeatCount))

	t.Run("should return error if an asset is already registered", testutils.Func(func(t *testing.T) {
		setup()
		axelarnetKeeper.RegisterIBCPathFunc = func(sdk.Context, string, string) error { return fmt.Errorf("failed") }
		msg = randomMsgRegisterIBCPath()
		_, err := server.RegisterIBCPath(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))
}

func TestHandleMsgRefundRequest(t *testing.T) {
	var (
		server          types.MsgServiceServer
		axelarnetKeeper *mock.BaseKeeperMock
		bankKeeper      *mock.BankKeeperMock
		ctx             sdk.Context
		router          sdk.Router
		msg             *types.RefundMsgRequest
	)
	setup := func() {
		axelarnetKeeper = &mock.BaseKeeperMock{
			LoggerFunc: func(ctx sdk.Context) log.Logger { return log.TestingLogger() },
			GetPendingRefundFunc: func(sdk.Context, types.RefundMsgRequest) (sdk.Coin, bool) {
				return sdk.NewCoin("uaxl", sdk.NewInt(1000)), true
			},
			DeletePendingRefundFunc: func(sdk.Context, types.RefundMsgRequest) { return },
		}
		bankKeeper = &mock.BankKeeperMock{
			SendCoinsFromModuleToAccountFunc: func(sdk.Context, string, sdk.AccAddress, sdk.Coins) error { return nil },
		}
		var tssHandler = func(_ sdk.Context, _ sdk.Msg) (*sdk.Result, error) {
			return &sdk.Result{}, nil
		}

		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		router = baseapp.NewRouter()
		router.AddRoute(sdk.NewRoute("tss", tssHandler))
		msgServiceRtr := baseapp.NewMsgServiceRouter()

		server = keeper.NewMsgServerImpl(axelarnetKeeper, &mock.NexusMock{}, bankKeeper, &mock.IBCTransferKeeperMock{}, &mock.ChannelKeeperMock{}, &mock.AccountKeeperMock{}, msgServiceRtr, router)
	}

	repeatCount := 20

	t.Run("should return error when unpack invalid inner message", testutils.Func(func(t *testing.T) {
		setup()

		any := cdctypes.Any{
			TypeUrl: rand.StrBetween(5, 20),
			Value:   rand.Bytes(int(rand.I64Between(100, 1000))),
		}
		msg = &types.RefundMsgRequest{
			Sender:       rand.AccAddr(),
			InnerMessage: &any,
		}

		_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should return error when failed to route inner message", testutils.Func(func(t *testing.T) {
		setup()

		msg = types.NewRefundMsgRequest(rand.AccAddr(), randomMsgLink())

		_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should return error when failed to executed inner message", testutils.Func(func(t *testing.T) {
		setup()

		var evmHandler = func(_ sdk.Context, _ sdk.Msg) (*sdk.Result, error) {
			return &sdk.Result{}, fmt.Errorf("failed to execute message")
		}
		router.AddRoute(sdk.NewRoute("evm", evmHandler))
		voteReq := &evmtypes.VoteConfirmChainRequest{Name: rand.StrBetween(5, 20)}
		msg = types.NewRefundMsgRequest(rand.AccAddr(), voteReq)

		_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("should not refund transaction fee when no pending refund", testutils.Func(func(t *testing.T) {
		setup()
		axelarnetKeeper.GetPendingRefundFunc = func(sdk.Context, types.RefundMsgRequest) (sdk.Coin, bool) { return sdk.Coin{}, false }

		msg = types.NewRefundMsgRequest(rand.AccAddr(), &tsstypes.HeartBeatRequest{})

		_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
	}).Repeat(repeatCount))

	t.Run("should refund transaction fee when executed inner message successfully", testutils.Func(func(t *testing.T) {
		setup()

		msg = types.NewRefundMsgRequest(rand.AccAddr(), &tsstypes.HeartBeatRequest{})

		_, err := server.RefundMsg(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, axelarnetKeeper.GetPendingRefundCalls(), 1)
		assert.Len(t, bankKeeper.SendCoinsFromModuleToAccountCalls(), 1)
	}).Repeat(repeatCount))
}

func TestHandleMsgRouteIBCTransfers(t *testing.T) {
	var (
		server          types.MsgServiceServer
		axelarnetKeeper *mock.BaseKeeperMock
		nexusKeeper     *mock.NexusMock
		bankKeeper      *mock.BankKeeperMock
		channelKeeper   *mock.ChannelKeeperMock
		transferKeeper  *mock.IBCTransferKeeperMock
		accountKeeper   *mock.AccountKeeperMock
		ctx             sdk.Context
		msg             *types.RouteIBCTransfersRequest

		transfers []nexus.CrossChainTransfer
	)
	setup := func() {
		ibcPath := randomIBCPath()
		axelarnetKeeper = &mock.BaseKeeperMock{
			GetIBCPathFunc: func(sdk.Context, string) (string, bool) {
				return ibcPath, true
			},
			GetCosmosChainFunc: func(sdk.Context, string) (string, bool) {
				return "cosmoschain", true
			},
			GetCosmosChainsFunc: func(sdk.Context) []string {
				var chains []string
				chains = append(chains, "cosmoschain")
				return chains
			},

			GetRouteTimeoutWindowFunc: func(ctx sdk.Context) uint64 { return 10 },
			SetPendingIBCTransferFunc: func(ctx sdk.Context, portID, channelID string, sequence uint64, value types.IBCTransfer) {},
		}
		nexusKeeper = &mock.NexusMock{
			GetTransfersForChainFunc: func(sdk.Context, nexus.Chain, nexus.TransferState) []nexus.CrossChainTransfer {
				transfers = []nexus.CrossChainTransfer{}
				for i := int64(0); i < rand.I64Between(1, 50); i++ {
					transfer := randomTransfer(testToken, testChain)
					transfers = append(transfers, transfer)
				}
				return transfers
			},
			ArchivePendingTransferFunc: func(sdk.Context, nexus.CrossChainTransfer) {},
			GetChainFunc: func(_ sdk.Context, chain string) (nexus.Chain, bool) {
				return nexus.Chain{
					Name:                  chain,
					NativeAsset:           randomDenom(),
					SupportsForeignAssets: true,
				}, true
			},
			IsAssetRegisteredFunc: func(sdk.Context, string, string) bool { return true },
		}
		bankKeeper = &mock.BankKeeperMock{
			MintCoinsFunc: func(sdk.Context, string, sdk.Coins) error { return nil },
		}
		channelKeeper = &mock.ChannelKeeperMock{
			GetChannelClientStateFunc: func(sdk.Context, string, string) (string, ibcclient.ClientState, error) {
				return "07-tendermint-0", clientState(), nil
			},
			GetNextSequenceSendFunc: func(ctx sdk.Context, portID, channelID string) (uint64, bool) { return uint64(rand.PosI64()), true },
		}
		transferKeeper = &mock.IBCTransferKeeperMock{
			SendTransferFunc: func(ctx sdk.Context, sourcePort, sourceChannel string, token sdk.Coin, sender sdk.AccAddress, receiver string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64) error {
				return nil
			},
		}
		accountKeeper = &mock.AccountKeeperMock{
			GetModuleAddressFunc: func(string) sdk.AccAddress { return rand.AccAddr() },
		}

		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		rtr := baseapp.NewRouter()
		msgServiceRtr := baseapp.NewMsgServiceRouter()
		server = keeper.NewMsgServerImpl(axelarnetKeeper, nexusKeeper, bankKeeper, transferKeeper, channelKeeper, accountKeeper, msgServiceRtr, rtr)
	}
	repeatCount := 20
	t.Run("should route ibc token back to cosmos chains, and archive pending transfers when get pending transfers from nexus keeper", testutils.Func(func(t *testing.T) {
		setup()
		msg = types.NewRouteIBCTransfersRequest(rand.AccAddr())
		_, err := server.RouteIBCTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
		assert.Len(t, axelarnetKeeper.SetPendingIBCTransferCalls(), len(transfers))
	}).Repeat(repeatCount))

	t.Run("should mint wrapped token and route to cosmos chains, and archive pending transfers when get pending transfers from nexus keeper", testutils.Func(func(t *testing.T) {
		setup()
		axelarnetKeeper.GetCosmosChainFunc = func(sdk.Context, string) (string, bool) { return "", false }
		msg = types.NewRouteIBCTransfersRequest(rand.AccAddr())
		_, err := server.RouteIBCTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))
		assert.Len(t, bankKeeper.MintCoinsCalls(), len(transfers))
		assert.Len(t, axelarnetKeeper.SetPendingIBCTransferCalls(), len(transfers))
	}).Repeat(repeatCount))

	t.Run("should return error when no path registered for cosmos chain", testutils.Func(func(t *testing.T) {
		setup()
		axelarnetKeeper.GetIBCPathFunc = func(sdk.Context, string) (string, bool) { return "", false }
		msg = types.NewRouteIBCTransfersRequest(rand.AccAddr())
		_, err := server.RouteIBCTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))
}

func randomMsgLink() *types.LinkRequest {
	return types.NewLinkRequest(
		rand.AccAddr(),
		rand.StrBetween(5, 100),
		rand.StrBetween(5, 100),
		rand.StrBetween(5, 100))

}

func randomMsgConfirmDeposit() *types.ConfirmDepositRequest {
	return types.NewConfirmDepositRequest(
		rand.AccAddr(),
		rand.BytesBetween(5, 100),
		sdk.NewCoin(randomDenom(), sdk.NewInt(rand.I64Between(1, 10000000000))),
		rand.AccAddr())
}
func randomMsgRegisterIBCPath() *types.RegisterIBCPathRequest {
	return types.NewRegisterIBCPathRequest(
		rand.AccAddr(),
		randomDenom(),
		randomIBCPath(),
	)

}

func randomTransfer(asset string, chain string) nexus.CrossChainTransfer {
	hash := sha256.Sum256(rand.BytesBetween(20, 50))
	ranAddr := sdk.AccAddress(hash[:20]).String()
	c := nexus.Chain{Name: chain, NativeAsset: "cosmos", SupportsForeignAssets: true}

	return nexus.CrossChainTransfer{
		Recipient: nexus.CrossChainAddress{Chain: c, Address: ranAddr},
		Asset:     sdk.NewInt64Coin(asset, rand.I64Between(1, 10000000000)),
		ID:        mathRand.Uint64(),
	}
}

func randomIBCDenom() string {
	return "ibc/" + rand.HexStr(64)
}

func randomDenom() string {
	d := rand.Strings(3, 10).WithAlphabet([]rune("abcdefghijklmnopqrstuvwxyz")).Take(1)
	return d[0]
}

func clientState() *ibctmtypes.ClientState {
	return ibctmtypes.NewClientState(
		"07-tendermint-0",
		ibctmtypes.DefaultTrustLevel,
		time.Hour*24*7*2,
		time.Hour*24*7*3,
		time.Second*10,
		clienttypes.NewHeight(0, 5),
		commitmenttypes.GetSDKSpecs(),
		ibctesting.UpgradePath,
		false,
		false,
	)
}
