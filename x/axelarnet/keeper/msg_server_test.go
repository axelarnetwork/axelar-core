package keeper_test

import (
	"crypto/sha256"
	"fmt"
	abci "github.com/tendermint/tendermint/abci/types"
	mathRand "math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"

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
		server = keeper.NewMsgServerImpl(nexusKeeper, &mock.BankKeeperMock{})
	}

	repeatCount := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgLink()
		_, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, nexusKeeper.LinkAddressesCalls(), 1)
		assert.Equal(t, msg.RecipientChain, nexusKeeper.GetChainCalls()[0].Chain)
	}).Repeat(repeatCount))

	t.Run("unknown chain", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgLink()
		nexusKeeper.GetChainFunc = func(sdk.Context, string) (nexus.Chain, bool) { return nexus.Chain{}, false }
		_, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("asset not registered", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgLink()
		nexusKeeper.IsAssetRegisteredFunc = func(sdk.Context, string, string) bool { return false }
		_, err := server.Link(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))
}

func TestHandleMsgConfirmDeposit(t *testing.T) {
	var (
		server      types.MsgServiceServer
		nexusKeeper *mock.NexusMock
		bankKeeper  *mock.BankKeeperMock
		ctx         sdk.Context
		msg         *types.ConfirmDepositRequest
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
			IsAssetRegisteredFunc:  func(sdk.Context, string, string) bool { return true },
			EnqueueForTransferFunc: func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) error { return nil },
		}
		bankKeeper = &mock.BankKeeperMock{
			BurnCoinsFunc:                    func(sdk.Context, string, sdk.Coins) error { return nil },
			SendCoinsFromAccountToModuleFunc: func(sdk.Context, sdk.AccAddress, string, sdk.Coins) error { return nil },
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		server = keeper.NewMsgServerImpl(nexusKeeper, bankKeeper)
	}

	repeatCount := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgConfirmDeposit()
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		events := ctx.EventManager().ABCIEvents()
		assert.NoError(t, err)
		assert.Len(t, testutils.Events(events).Filter(func(event abci.Event) bool { return event.Type == types.EventTypeDepositConfirmation }), 1)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 1)
		assert.Len(t, bankKeeper.BurnCoinsCalls(), 1)
		assert.Len(t, bankKeeper.SendCoinsFromAccountToModuleCalls(), 1)
		assert.Equal(t, msg.Token.Denom, nexusKeeper.EnqueueForTransferCalls()[0].Amount.Denom)
		assert.Equal(t, msg.Token.Amount, nexusKeeper.EnqueueForTransferCalls()[0].Amount.Amount)
	}).Repeat(repeatCount))

	t.Run("enqueue transfer failed", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgConfirmDeposit()
		nexusKeeper.EnqueueForTransferFunc = func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) error {
			return fmt.Errorf("failed")
		}

		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
	}).Repeat(repeatCount))

	t.Run("burn token failed", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgConfirmDeposit()
		bankKeeper.BurnCoinsFunc = func(sdk.Context, string, sdk.Coins) error {
			return fmt.Errorf("failed")
		}

		assert.Panics(t, func() { _, _ = server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg) }, "ConfirmDeposit did not panic when burn token failed")
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeatCount))

	t.Run("transfer from account to module failed", testutils.Func(func(t *testing.T) {
		setup()
		msg = randomMsgConfirmDeposit()
		bankKeeper.SendCoinsFromAccountToModuleFunc = func(sdk.Context, sdk.AccAddress, string, sdk.Coins) error {
			return fmt.Errorf("failed")
		}
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), msg)
		assert.Error(t, err)
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
	}).Repeat(repeatCount))
}

func TestHandleMsgExecutePendingTransfers(t *testing.T) {
	var (
		server      types.MsgServiceServer
		nexusKeeper *mock.NexusMock
		bankKeeper  *mock.BankKeeperMock
		ctx         sdk.Context
		msg         *types.ExecutePendingTransfersRequest

		transfers []nexus.CrossChainTransfer
	)
	setup := func() {

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
					NativeAsset:           rand.StrBetween(5, 20),
					SupportsForeignAssets: true,
				}, true
			},
			IsAssetRegisteredFunc:  func(sdk.Context, string, string) bool { return true },
			EnqueueForTransferFunc: func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) error { return nil },
		}
		bankKeeper = &mock.BankKeeperMock{
			MintCoinsFunc:                    func(sdk.Context, string, sdk.Coins) error { return nil },
			SendCoinsFromModuleToAccountFunc: func(sdk.Context, string, sdk.AccAddress, sdk.Coins) error { return nil },
		}
		ctx = sdk.NewContext(nil, tmproto.Header{Height: rand.PosI64()}, false, log.TestingLogger())
		server = keeper.NewMsgServerImpl(nexusKeeper, bankKeeper)
	}

	repeatCount := 20
	t.Run("happy path", testutils.Func(func(t *testing.T) {
		setup()
		msg = types.NewExecutePendingTransfersRequest(rand.Bytes(sdk.AddrLen))
		_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg)
		assert.NoError(t, err)
		assert.Len(t, bankKeeper.MintCoinsCalls(), len(transfers))
		assert.Len(t, bankKeeper.SendCoinsFromModuleToAccountCalls(), len(transfers))
		assert.Len(t, nexusKeeper.ArchivePendingTransferCalls(), len(transfers))

	}).Repeat(repeatCount))

	t.Run("mint token failed", testutils.Func(func(t *testing.T) {
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

	t.Run("transfer from module to account failed", testutils.Func(func(t *testing.T) {
		setup()
		bankKeeper.SendCoinsFromModuleToAccountFunc = func(sdk.Context, string, sdk.AccAddress, sdk.Coins) error {
			return fmt.Errorf("failed")
		}
		msg = types.NewExecutePendingTransfersRequest(rand.Bytes(sdk.AddrLen))
		assert.Panics(t, func() { _, _ = server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), msg) }, "ExecutePendingTransfers did not panic when transfer token failed")
		assert.Len(t, nexusKeeper.EnqueueForTransferCalls(), 0)
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
		rand.StrBetween(5, 100),
		rand.BytesBetween(5, 100),
		sdk.NewCoin("testDenom", sdk.NewInt(rand.I64Between(1, 10000000000))),
		rand.Bytes(sdk.AddrLen))
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
