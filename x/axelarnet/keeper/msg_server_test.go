package keeper_test

import (
	"crypto/sha256"
	"fmt"
	mathRand "math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
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
	btc "github.com/axelarnetwork/axelar-core/x/bitcoin/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
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
		server = keeper.NewMsgServerImpl(&mock.BaseKeeperMock{}, nexusKeeper, &mock.BankKeeperMock{}, &mock.IBCTransferKeeperMock{})
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
			GetIBCPathFunc: func(sdk.Context, string) string {
				return ibcPath
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
		server = keeper.NewMsgServerImpl(axelarnetKeeper, nexusKeeper, bankKeeper, transferKeeper)
	}

	repeatCount := 20
	t.Run("should enqueue transfer in nexus keeper when registered tokens are sent from burner address to bank keeper, and burned", testutils.Func(func(t *testing.T) {
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
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeatCount))

	t.Run("should return error when SendCoinsFromAccountToModule in bank keeper failed", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgConfirmDeposit()
		bankKeeper.SendCoinsFromAccountToModuleFunc = func(sdk.Context, sdk.AccAddress, string, sdk.Coins) error {
			return fmt.Errorf("failed")
		}
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeatCount))

	t.Run("should enqueue transfer in nexus keeper when registered ICS20 tokens are sent from burner address to escrow address", testutils.Func(func(t *testing.T) {
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
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
		assert.Len(t, nexusKeeper.AddToChainTotalCalls(), 0)
		assert.Len(t, bankKeeper.SendCoinsCalls(), 0)

	}).Repeat(repeatCount))

	t.Run("should return error when ICS20 token path not registered in axelarnet keeper", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		axelarnetKeeper.GetIBCPathFunc = func(sdk.Context, string) string {
			return ""
		}
		msg = randomMsgConfirmDeposit()
		msg.Token.Denom = randomIBCDenom()
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
		assert.Len(t, nexusKeeper.AddToChainTotalCalls(), 0)
		assert.Len(t, bankKeeper.SendCoinsCalls(), 0)

	}).Repeat(repeatCount))

	t.Run("should return error when ICS20 token tracing path does not match registered path in axelarnet keeper", testutils.Func(func(t *testing.T) {
		setup()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		axelarnetKeeper.GetIBCPathFunc = func(sdk.Context, string) string {
			return randomIBCPath()
		}
		msg = randomMsgConfirmDeposit()
		msg.Token.Denom = randomIBCDenom()
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
		assert.Len(t, nexusKeeper.AddToChainTotalCalls(), 0)
		assert.Len(t, bankKeeper.SendCoinsCalls(), 0)

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
		assert.Len(t, bankKeeper.SendCoinsCalls(), 1)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
		assert.Len(t, nexusKeeper.AddToChainTotalCalls(), 0)

	}).Repeat(repeatCount))

	t.Run("should enqueue transfer in nexus keeper when native axelar tokens are sent from burner address to escrow address", testutils.Func(func(t *testing.T) {
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
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)

	}).Repeat(repeatCount))
}

func TestHandleMsgExecutePendingTransfers(t *testing.T) {
	var (
		server          types.MsgServiceServer
		axelarnetKeeper *mock.BaseKeeperMock
		nexusKeeper     *mock.NexusMock
		bankKeeper      *mock.BankKeeperMock
		ctx             sdk.Context
		msg             *types.ExecutePendingTransfersRequest

		transfers []nexus.CrossChainTransfer
	)
	setup := func() {
		axelarnetKeeper = &mock.BaseKeeperMock{
			GetIBCPathFunc: func(sdk.Context, string) string {
				return ""
			},
		}
		nexusKeeper = &mock.NexusMock{
			GetTransfersForChainFunc: func(sdk.Context, nexus.Chain, nexus.TransferState) []nexus.CrossChainTransfer {
				transfers = []nexus.CrossChainTransfer{}
				for i := int64(0); i < rand.I64Between(1, 50); i++ {
					transfer := randomTransfer()
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
			MintCoinsFunc:                    func(sdk.Context, string, sdk.Coins) error { return nil },
			SendCoinsFromModuleToAccountFunc: func(sdk.Context, string, sdk.AccAddress, sdk.Coins) error { return nil },
			SendCoinsFunc:                    func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error { return nil },
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		server = keeper.NewMsgServerImpl(axelarnetKeeper, nexusKeeper, bankKeeper, &mock.IBCTransferKeeperMock{})
	}

	repeatCount := 20
	t.Run("should mint and send token to recipients, and archive pending transfers when get pending transfers from nexus keeper ", testutils.Func(func(t *testing.T) {
		setup()
		msg = types.NewExecutePendingTransfersRequest(rand.Bytes(sdk.AddrLen))
		_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, bankKeeper.MintCoinsCalls(), len(transfers))
		assert.Len(t, bankKeeper.SendCoinsFromModuleToAccountCalls(), len(transfers))
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))

	}).Repeat(repeatCount))

	t.Run("should return error when MintCoins in bank keeper failed", testutils.Func(func(t *testing.T) {
		setup()
		bankKeeper.MintCoinsFunc = func(sdk.Context, string, sdk.Coins) error {
			return fmt.Errorf("failed")
		}
		msg = types.NewExecutePendingTransfersRequest(rand.Bytes(sdk.AddrLen))
		_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
		assert.Len(t, bankKeeper.SendCoinsFromModuleToAccountCalls(), 0)
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), 0)
	}).Repeat(repeatCount))

	t.Run("should panic when SendCoinsFromModuleToAccount in bank keeper failed", testutils.Func(func(t *testing.T) {
		setup()
		bankKeeper.SendCoinsFromModuleToAccountFunc = func(sdk.Context, string, sdk.AccAddress, sdk.Coins) error {
			return fmt.Errorf("failed")
		}
		msg = types.NewExecutePendingTransfersRequest(rand.Bytes(sdk.AddrLen))
		assert.Panics(t, func() { _, _ = server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg) }, "ExecutePendingTransfers did not panic when transfer token failed")
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeatCount))

	t.Run("should send ICS20 token from escrow account to recipients, and archive pending transfers \\"+
		"when pending transfer asset is registered with an IBC tracing path ", testutils.Func(func(t *testing.T) {
		setup()
		axelarnetKeeper.GetIBCPathFunc = func(sdk.Context, string) string {
			return randomIBCPath()
		}

		msg = types.NewExecutePendingTransfersRequest(rand.Bytes(sdk.AddrLen))
		_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, bankKeeper.SendCoinsCalls(), len(transfers))
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))

	}).Repeat(repeatCount))

	t.Run("should send axelarnet native token from escrow account to recipients, and archive pending transfers \\"+
		"when pending transfer asset axelarnet native token", testutils.Func(func(t *testing.T) {
		setup()
		transfers = []nexus.CrossChainTransfer{}
		for i := int64(0); i < rand.I64Between(1, 50); i++ {
			transfer := randomTransfer()
			transfers = append(transfers, transfer)
		}
		nativeAssetCount := int(rand.I64Between(1, int64(len(transfers)+1)))
		for i := 0; i < nativeAssetCount; i++ {
			transfers[i].Asset.Denom = exported.Axelarnet.NativeAsset
		}
		nexusKeeper.GetTransfersForChainFunc = func(sdk.Context, nexus.Chain, nexus.TransferState) []nexus.CrossChainTransfer {
			return transfers
		}
		msg = types.NewExecutePendingTransfersRequest(rand.Bytes(sdk.AddrLen))
		_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, bankKeeper.SendCoinsCalls(), nativeAssetCount)
		assert.Len(t, bankKeeper.MintCoinsCalls(), len(transfers)-nativeAssetCount)
		assert.Len(t, bankKeeper.SendCoinsFromModuleToAccountCalls(), len(transfers)-nativeAssetCount)
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
		server = keeper.NewMsgServerImpl(axelarnetKeeper, &mock.NexusMock{}, &mock.BankKeeperMock{}, &mock.IBCTransferKeeperMock{})

	}

	repeatCount := 20
	t.Run("should register an IBC tracing path for an asset if not registered yet", testutils.Func(func(t *testing.T) {
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
		assert.Len(t, axelarnetKeeper.RegisterIBCPathCalls(), 1)
	}).Repeat(repeatCount))
}

func randomMsgLink() *types.LinkRequest {
	return types.NewLinkRequest(
		rand.Bytes(sdk.AddrLen),
		rand.StrBetween(5, 100),
		rand.StrBetween(5, 100),
		rand.StrBetween(5, 100))

}

func randomMsgConfirmDeposit() *types.ConfirmDepositRequest {
	return types.NewConfirmDepositRequest(
		rand.Bytes(sdk.AddrLen),
		rand.BytesBetween(5, 100),
		sdk.NewCoin(randomDenom(), sdk.NewInt(rand.I64Between(1, 10000000000))),
		rand.Bytes(sdk.AddrLen))
}
func randomMsgRegisterIBCPath() *types.RegisterIBCPathRequest {
	return types.NewRegisterIBCPathRequest(
		rand.Bytes(sdk.AddrLen),
		randomDenom(),
		randomIBCPath(),
	)

}

func randomTransfer() nexus.CrossChainTransfer {
	hash := sha256.Sum256(rand.BytesBetween(20, 50))
	ranAddr := sdk.AccAddress(hash[:20]).String()

	return nexus.CrossChainTransfer{
		Recipient: nexus.CrossChainAddress{Chain: exported.Axelarnet, Address: ranAddr},
		Asset:     sdk.NewInt64Coin(btc.Bitcoin.NativeAsset, rand.I64Between(1, 10000000000)),
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
