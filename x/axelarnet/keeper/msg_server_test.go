package keeper_test

import (
	"crypto/sha256"
	"fmt"
	mathRand "math/rand"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v2/modules/core/02-client/types"
	ibcclient "github.com/cosmos/ibc-go/v2/modules/core/exported"
	"github.com/stretchr/testify/assert"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	axelartestutils "github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestHandleMsgLink(t *testing.T) {
	var (
		server types.MsgServiceServer
		k      keeper.Keeper
		nexusK *mock.NexusMock
		ctx    sdk.Context
		req    *types.LinkRequest
	)

	givenMsgServer := Given("an axelarnet msg server", func() {
		ctx, k, _ = setup()
		k.InitGenesis(ctx, types.DefaultGenesisState())
		nexusK = &mock.NexusMock{}
		ibcK := keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{}, &mock.ChannelKeeperMock{})
		server = keeper.NewMsgServerImpl(k, nexusK, &mock.BankKeeperMock{}, &mock.IBCTransferKeeperMock{}, &mock.AccountKeeperMock{}, ibcK)
	})

	whenChainIsRegistered := When("chain is registered", func() {
		nexusK.GetChainFunc = func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{
				Name:                  chain,
				SupportsForeignAssets: true,
				Module:                rand.Str(10),
			}, true
		}
	})

	whenAssetIsRegistered := When("asset is registered", func() {
		nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool { return true }
	})

	linkSucceeds := When("link addresses succeeds", func() {
		nexusK.LinkAddressesFunc = func(sdk.Context, nexus.CrossChainAddress, nexus.CrossChainAddress) error {
			return nil
		}
	})
	requestIsMade := When("a link request is made", func() {
		req = types.NewLinkRequest(
			rand.AccAddr(),
			rand.StrBetween(5, 100),
			rand.StrBetween(5, 100),
			rand.StrBetween(5, 100))
	})

	linkFails := Then("link addresses request fails", func(t *testing.T) {
		_, err := server.Link(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
	})

	t.Run("link addresses", func(t *testing.T) {
		givenMsgServer.
			Branch(
				whenChainIsRegistered.
					When2(whenAssetIsRegistered).
					When2(linkSucceeds).
					When2(requestIsMade).
					Then("link succeeds", func(t *testing.T) {
						_, err := server.Link(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
					}),
				When("chain is not registered", func() {
					nexusK.GetChainFunc = func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
						return nexus.Chain{}, false
					}
				}).
					When2(requestIsMade).
					Then2(linkFails),

				whenChainIsRegistered.
					When("asset is not registered", func() {
						nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool { return false }
					}).
					When2(requestIsMade).
					Then2(linkFails),
			).Run(t)
	})
}

func TestHandleMsgConfirmDeposit(t *testing.T) {
	var (
		server    types.MsgServiceServer
		k         keeper.Keeper
		nexusK    *mock.NexusMock
		bankK     *mock.BankKeeperMock
		transferK *mock.IBCTransferKeeperMock
		ctx       sdk.Context
		req       *types.ConfirmDepositRequest
	)

	ibcPath := randomIBCPath()
	chain := nexustestutils.Chain()
	givenMsgServer := Given("an axelarnet msg server", func() {
		ctx, k, _ = setup()
		k.InitGenesis(ctx, types.DefaultGenesisState())
		k.SetCosmosChain(ctx, types.CosmosChain{
			Name:       chain.Name,
			AddrPrefix: rand.StrBetween(1, 10),
		})

		nexusK = &mock.NexusMock{
			GetChainFunc: func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) {
				return chain, true
			},
			GetChainByNativeAssetFunc: func(sdk.Context, string) (nexus.Chain, bool) {
				return chain, true
			},
		}
		bankK = &mock.BankKeeperMock{}
		transferK = &mock.IBCTransferKeeperMock{
			GetDenomTraceFunc: func(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctypes.DenomTrace, bool) {
				return ibctypes.DenomTrace{
					Path:      ibcPath,
					BaseDenom: rand.Denom(5, 10),
				}, true
			},
		}
		ibcK := keeper.NewIBCKeeper(k, transferK, &mock.ChannelKeeperMock{})
		server = keeper.NewMsgServerImpl(k, nexusK, bankK, transferK, &mock.AccountKeeperMock{}, ibcK)
	})

	recipientIsFound := When("recipient is found", func() {
		nexusK.GetRecipientFunc = func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, true
		}
	})

	whenDepositAddressHasBalance := When("deposit address has balance", func() {
		bankK.GetBalanceFunc = func(_ sdk.Context, _ sdk.AccAddress, denom string) sdk.Coin {
			return sdk.NewCoin(denom, sdk.NewInt(rand.I64Between(1, 1e18)))
		}
	})

	chainIsActivated := When("chain is activated", func() {
		nexusK.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool { return true }
	})

	assetIsRegistered := When("asset is registered", func() {
		nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool { return true }
	})

	assetIsLinkedToCosmosChain := When("asset is linked to a cosmos chain", func() {
		nexusK.GetChainByNativeAssetFunc = func(ctx sdk.Context, asset string) (nexus.Chain, bool) {
			return chain, true
		}
	})

	sendCoinSucceeds := When("send to module account succeeds", func() {
		bankK.SendCoinsFunc = func(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error {
			return nil
		}
	})

	pathIsRegistered := When("denom path matches registered path", func() {
		err := k.SetIBCPath(ctx, chain.Name, ibcPath)
		assert.NoError(t, err)
	})

	enqueueTransferSucceeds := When("enqueue transfer succeeds", func() {
		nexusK.EnqueueForTransferFunc = func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) (nexus.TransferID, error) {
			return nexus.TransferID(rand.I64Between(1, 9999)), nil
		}
	})

	confirmExternalICS20TokenRequest := When("a confirm external ICS20 token deposit request is made", func() {
		req = randomMsgConfirmDeposit()
		req.Denom = fmt.Sprintf("ibc/%s", rand.HexStr(64))
	})

	confirmNativeAXLRequest := When("a confirm native AXL token deposit request is made", func() {
		req = randomMsgConfirmDeposit()
		req.Denom = exported.NativeAsset
	})

	confirmExternalERC20Token := When("a confirm external ERC20 token deposit request is made", func() {
		req = randomMsgConfirmDeposit()
	})

	confirmDepositFails := Then("confirm deposit request fails", func(t *testing.T) {
		_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
	})

	t.Run("confirm deposit", func(t *testing.T) {
		givenMsgServer.
			Branch(
				When("deposit address holds no funds", func() {
					bankK.GetBalanceFunc = func(_ sdk.Context, _ sdk.AccAddress, denom string) sdk.Coin {
						return sdk.NewCoin(denom, sdk.ZeroInt())
					}
				}).
					When2(confirmExternalICS20TokenRequest).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When("recipient is not found", func() {
						nexusK.GetRecipientFunc = func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
							return nexus.CrossChainAddress{}, false
						}
					}).
					When2(confirmExternalICS20TokenRequest).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When("chain is not activated", func() {
						nexusK.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool { return false }
					}).
					When2(confirmExternalICS20TokenRequest).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When("asset is not registered", func() {
						nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool { return false }
						nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
							return nexus.Chain{}, false
						}
					}).
					When2(confirmExternalICS20TokenRequest).
					When("confirm an invalid IBC denom", func() {
						req.Denom = fmt.Sprintf("ibc/%s", rand.HexStr(50))
					}).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When2(assetIsLinkedToCosmosChain).
					When("denom path does not match registered path", func() {
						funcs.MustNoErr(k.SetIBCPath(ctx, chain.Name, randomIBCPath()))
					}).
					When2(confirmExternalICS20TokenRequest).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When2(assetIsLinkedToCosmosChain).
					When2(pathIsRegistered).
					When("send to escrow account fails", func() {
						bankK.SendCoinsFunc = func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
							return fmt.Errorf("failed to send %s from %s to %s", amt.String(), fromAddr.String(), toAddr.String())
						}
					}).
					When2(confirmExternalICS20TokenRequest).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When2(assetIsLinkedToCosmosChain).
					When2(pathIsRegistered).
					When2(sendCoinSucceeds).
					When("enqueue transfer fails", func() {
						nexusK.EnqueueForTransferFunc = func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) (nexus.TransferID, error) {
							return nexus.TransferID(0), fmt.Errorf("failed to enqueue tranfer")
						}
					}).
					When2(confirmExternalICS20TokenRequest).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When2(assetIsLinkedToCosmosChain).
					When2(pathIsRegistered).
					When2(sendCoinSucceeds).
					When2(enqueueTransferSucceeds).
					When2(confirmExternalICS20TokenRequest).
					Then("confirm external IBC deposit succeeds", func(t *testing.T) {
						_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
					}),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When("is native asset on Axelarnet", func() {
						nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
							return exported.Axelarnet, true
						}
					}).
					When("send to escrow account fails", func() {
						bankK.SendCoinsFunc = func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
							return fmt.Errorf("failed to send %s from %s to %s", amt.String(), fromAddr.String(), toAddr.String())
						}
					}).
					When2(confirmNativeAXLRequest).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When("is native asset on Axelarnet", func() {
						nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
							return exported.Axelarnet, true
						}
					}).
					When2(sendCoinSucceeds).
					When2(enqueueTransferSucceeds).
					When2(confirmNativeAXLRequest).
					Then("confirm native AXL deposit succeeds", func(t *testing.T) {
						_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
					}),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When("asset is not registered", func() {
						nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool { return false }
					}).
					When2(confirmExternalERC20Token).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When2(assetIsRegistered).
					When("send coins to module account succeeds", func() {
						bankK.SendCoinsFromAccountToModuleFunc = func(sdk.Context, sdk.AccAddress, string, sdk.Coins) error {
							return nil
						}
					}).
					When("burn coin succeeds", func() {
						bankK.BurnCoinsFunc = func(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
							return nil
						}
					}).
					When2(enqueueTransferSucceeds).
					When2(confirmExternalERC20Token).
					Then("confirm external ERC20 deposit succeeds", func(t *testing.T) {
						_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
					}),
			).Run(t)
	})
}

func TestHandleMsgExecutePendingTransfers(t *testing.T) {
	var (
		server    types.MsgServiceServer
		k         keeper.Keeper
		nexusK    *mock.NexusMock
		bankK     *mock.BankKeeperMock
		transferK *mock.IBCTransferKeeperMock
		ctx       sdk.Context
		req       *types.ExecutePendingTransfersRequest
	)

	givenMsgServer := Given("an axelarnet msg server", func() {
		ctx, k, _ = setup()
		k.InitGenesis(ctx, types.DefaultGenesisState())
		funcs.MustNoErr(k.SetFeeCollector(ctx, rand.AccAddr()))
		nexusK = &mock.NexusMock{
			GetChainFunc: func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) {
				return exported.Axelarnet, true
			},
			ArchivePendingTransferFunc: func(sdk.Context, nexus.CrossChainTransfer) {},
			GetTransferFeesFunc: func(sdk.Context) sdk.Coins {
				return sdk.Coins{}
			},
			SubTransferFeeFunc: func(sdk.Context, sdk.Coin) {},
		}
		bankK = &mock.BankKeeperMock{}
		transferK = &mock.IBCTransferKeeperMock{}
		accountK := &mock.AccountKeeperMock{
			GetModuleAddressFunc: func(moduleName string) sdk.AccAddress {
				return rand.AccAddr()
			},
		}
		ibcK := keeper.NewIBCKeeper(k, transferK, &mock.ChannelKeeperMock{})
		server = keeper.NewMsgServerImpl(k, nexusK, bankK, transferK, accountK, ibcK)
	})

	whenAssetOriginsFromExternalCosmosChain := When("asset is from external cosmos chain", func() {
		chain := nexustestutils.Chain()
		k.SetCosmosChain(ctx, types.CosmosChain{
			Name:       chain.Name,
			AddrPrefix: rand.StrBetween(1, 10),
		})
		assert.NotPanics(t, func() {
			funcs.MustNoErr(k.SetIBCPath(ctx, chain.Name, randomIBCPath()))
		})
		nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
			return chain, true
		}
	})

	hasPendingTransfers := When("has pending transfers", func() {
		nexusK.GetTransfersForChainFunc = func(sdk.Context, nexus.Chain, nexus.TransferState) []nexus.CrossChainTransfer {
			return []nexus.CrossChainTransfer{randomTransfer(rand.Denom(2, 10), nexus.ChainName(rand.StrBetween(2, 10)))}
		}
	})

	sendCoinSucceeds := When("send coins succeeds", func() {
		bankK.SendCoinsFunc = func(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error {
			return nil
		}
	})

	requestIsMade := When("an execute pending transfer request is made", func() {
		req = types.NewExecutePendingTransfersRequest(rand.AccAddr())
	})

	t.Run("execute pending transfers", func(t *testing.T) {
		givenMsgServer.
			Branch(
				When("no pending transfer", func() {
					nexusK.GetTransfersForChainFunc = func(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState) []nexus.CrossChainTransfer {
						return []nexus.CrossChainTransfer{}
					}
				}).
					When2(requestIsMade).
					Then("do nothing", func(t *testing.T) {
						_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Len(t, bankK.MintCoinsCalls(), 0)
						assert.Len(t, bankK.SendCoinsCalls(), 0)
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), 0)
					}),

				whenAssetOriginsFromExternalCosmosChain.
					When2(hasPendingTransfers).
					When("send coins fails", func() {
						bankK.SendCoinsFunc = func(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
							return fmt.Errorf("failed to send %s from %s to %s", amt.String(), fromAddr.String(), toAddr.String())
						}
					}).
					When2(requestIsMade).
					Then("should not archive the transfer", func(t *testing.T) {
						_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Len(t, bankK.MintCoinsCalls(), 0)
						assert.Len(t, bankK.SendCoinsCalls(), 1)
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), 0)
					}),

				whenAssetOriginsFromExternalCosmosChain.
					When2(hasPendingTransfers).
					When2(sendCoinSucceeds).
					When2(requestIsMade).
					Then("archive the transfer", func(t *testing.T) {
						_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Len(t, bankK.MintCoinsCalls(), 0)
						assert.Len(t, bankK.SendCoinsCalls(), 1)
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), 1)
					}),

				When("asset is native on Axelarnet", func() {
					nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
						return exported.Axelarnet, true
					}
				}).
					When2(hasPendingTransfers).
					When2(sendCoinSucceeds).
					When2(requestIsMade).
					Then("send coin and archive the transfer", func(t *testing.T) {
						_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Len(t, bankK.MintCoinsCalls(), 0)
						assert.Len(t, bankK.SendCoinsCalls(), 1)
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), 1)
					}),

				When("asset is not registered", func() {
					nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool {
						return false
					}
				}).
					When2(requestIsMade).
					Then("should panic", func(t *testing.T) {
						assert.Panics(t, func() {
							_, _ = server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), req)
						})
					}),

				When("asset is registered", func() {
					nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool {
						return true
					}
					nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
						return nexustestutils.Chain(), true
					}
				}).
					When2(hasPendingTransfers).
					When("mint coins succeeds", func() {
						bankK.MintCoinsFunc = func(sdk.Context, string, sdk.Coins) error {
							return nil
						}
					}).
					When2(sendCoinSucceeds).
					When2(requestIsMade).
					Then("mint coin and archive the transfer", func(t *testing.T) {
						_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Len(t, bankK.MintCoinsCalls(), 1)
						assert.Len(t, bankK.SendCoinsCalls(), 1)
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), 1)
					}),
			).Run(t)
	})
}

func TestHandleMsgRegisterIBCPath(t *testing.T) {
	var (
		server types.MsgServiceServer
		k      keeper.Keeper
		ctx    sdk.Context
		req    *types.RegisterIBCPathRequest
	)

	givenMsgServer := Given("an axelarnet msg server", func() {
		ctx, k, _ = setup()
		k.InitGenesis(ctx, types.DefaultGenesisState())
		ibcK := keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{}, &mock.ChannelKeeperMock{})
		server = keeper.NewMsgServerImpl(k, &mock.NexusMock{}, &mock.BankKeeperMock{}, &mock.IBCTransferKeeperMock{}, &mock.AccountKeeperMock{}, ibcK)
	})

	whenChainIsACosmosChain := When("chain is a cosmos chain", func() {
		k.SetCosmosChain(ctx, types.CosmosChain{Name: req.Chain})
	})

	requestIsMade := When("a register IBC path request is made", func() {
		req = types.NewRegisterIBCPathRequest(
			rand.AccAddr(),
			rand.Denom(5, 10),
			randomIBCPath(),
		)
	})

	registerFailed := Then("register IBC path request fails", func(t *testing.T) {
		_, err := server.RegisterIBCPath(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
	})

	t.Run("register IBC path", func(t *testing.T) {
		givenMsgServer.
			Branch(
				When("chain is not a cosmos chain", func() {}).
					When2(requestIsMade).
					Then2(registerFailed),

				whenChainIsACosmosChain.
					When("path is already registered", func() {
						funcs.MustNoErr(k.SetIBCPath(ctx, req.Chain, randomIBCPath()))
					}).
					When2(requestIsMade).
					Then2(registerFailed),

				requestIsMade.
					When2(whenChainIsACosmosChain).
					When("path is not registered", func() {}).
					Then("register path", func(t *testing.T) {
						_, err := server.RegisterIBCPath(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
					}),
			).Run(t)
	})
}

func TestHandleMsgRouteIBCTransfers(t *testing.T) {
	var (
		server       types.MsgServiceServer
		k            keeper.Keeper
		nexusK       *mock.NexusMock
		bankK        *mock.BankKeeperMock
		transferK    *mock.IBCTransferKeeperMock
		ctx          sdk.Context
		req          *types.RouteIBCTransfersRequest
		cosmosChains []types.CosmosChain
		transfersNum int
	)

	givenMsgServer := Given("an axelarnet msg server", func() {
		ctx, k, _ = setup()
		k.InitGenesis(ctx, types.DefaultGenesisState())
		transfersNum = 0
		cosmosChains = slices.Expand(func(i int) types.CosmosChain {
			return types.CosmosChain{
				Name:       nexus.ChainName(fmt.Sprintf("cosmoschain-%d", i)),
				IBCPath:    fmt.Sprintf("transfer/channel-%d", i),
				AddrPrefix: fmt.Sprintf("cosmos%d", i),
			}
		}, 5)

		slices.ForEach(cosmosChains, func(c types.CosmosChain) {
			k.SetCosmosChain(ctx, c)
		})

		nexusK = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{Name: chain}, true
			},
			GetChainByNativeAssetFunc: func(sdk.Context, string) (nexus.Chain, bool) {
				return nexustestutils.Chain(), true
			},
			ArchivePendingTransferFunc: func(sdk.Context, nexus.CrossChainTransfer) {},
		}
		bankK = &mock.BankKeeperMock{}
		transferK = &mock.IBCTransferKeeperMock{}
		accountK := &mock.AccountKeeperMock{
			GetModuleAddressFunc: func(string) sdk.AccAddress {
				return rand.AccAddr()
			},
		}
		ibcK := keeper.NewIBCKeeper(k, transferK, &mock.ChannelKeeperMock{})
		server = keeper.NewMsgServerImpl(k, nexusK, bankK, transferK, accountK, ibcK)
	})

	whenAssetOriginsFromExternalCosmosChain := When("asset is from external cosmos chain", func() {
		nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
			chainName := cosmosChains[rand.I64Between(0, int64(len(cosmosChains)))].Name
			return nexus.Chain{Name: chainName}, true
		}

	})
	hasPendingTranfers := When("has pending transfers", func() {
		nexusK.GetTransfersForChainFunc = func(_ sdk.Context, chain nexus.Chain, _ nexus.TransferState) []nexus.CrossChainTransfer {
			var transfers []nexus.CrossChainTransfer
			for i := int64(0); i < rand.I64Between(1, 5); i++ {
				chainName := chain.Name
				transfers = append(transfers, randomTransfer(rand.Denom(2, 10), chainName))
			}
			transfersNum += len(transfers)
			return transfers
		}
	})

	requestIsMade := When("a route IBC transfers request is made", func() {
		req = types.NewRouteIBCTransfersRequest(rand.AccAddr())
	})

	doNothing := Then("do nothing", func(t *testing.T) {
		_, err := server.RouteIBCTransfers(sdk.WrapSDKContext(ctx), req)
		assert.NoError(t, err)
		assert.Len(t, bankK.MintCoinsCalls(), 0)
		assert.Len(t, nexusK.ArchivePendingTransferCalls(), 0)
	})

	t.Run("route IBC transfers", func(t *testing.T) {
		givenMsgServer.
			Branch(
				When("dest chain is not found", func() {
					nexusK.GetChainFunc = func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) {
						return nexus.Chain{}, false
					}
				}).
					When2(requestIsMade).
					Then2(doNothing),

				When("dest chain is Axelarnet", func() {
					nexusK.GetChainFunc = func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) {
						return exported.Axelarnet, true
					}
				}).
					When2(requestIsMade).
					Then2(doNothing),

				When("no pending transfer", func() {
					nexusK.GetTransfersForChainFunc = func(sdk.Context, nexus.Chain, nexus.TransferState) []nexus.CrossChainTransfer {
						return []nexus.CrossChainTransfer{}
					}
				}).
					When2(requestIsMade).
					Then2(doNothing),

				whenAssetOriginsFromExternalCosmosChain.
					When2(hasPendingTranfers).
					When2(requestIsMade).
					Then("archive the transfer", func(t *testing.T) {
						_, err := server.RouteIBCTransfers(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Len(t, bankK.MintCoinsCalls(), 0)
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), transfersNum, fmt.Sprintf("expected %d got %d", transfersNum, len(nexusK.ArchivePendingTransferCalls())))
					}),

				When("asset is native on Axelarnet", func() {
					nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
						return exported.Axelarnet, true
					}
				}).
					When2(hasPendingTranfers).
					When2(requestIsMade).
					Then("send coin, archive the transfer", func(t *testing.T) {
						_, err := server.RouteIBCTransfers(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Len(t, bankK.MintCoinsCalls(), 0)
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), transfersNum)
					}),

				When("asset is not registered", func() {
					nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool {
						return false
					}
				}).
					When2(requestIsMade).
					Then("should panic", func(t *testing.T) {
						assert.Panics(t, func() {
							_, _ = server.RouteIBCTransfers(sdk.WrapSDKContext(ctx), req)
						})
					}),

				When("asset is registered", func() {
					nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool {
						return true
					}
				}).
					When2(hasPendingTranfers).
					When("mint succeeds", func() {
						bankK.MintCoinsFunc = func(sdk.Context, string, sdk.Coins) error {
							return nil
						}
					}).
					When2(requestIsMade).
					Then("mint coin, archive the transfer", func(t *testing.T) {
						_, err := server.RouteIBCTransfers(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Len(t, bankK.MintCoinsCalls(), transfersNum)
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), transfersNum)
					}),
			).Run(t)
	})
}

func TestRetryIBCTransfer(t *testing.T) {
	var (
		server   types.MsgServiceServer
		k        keeper.Keeper
		n        *mock.NexusMock
		b        *mock.BankKeeperMock
		i        *mock.IBCTransferKeeperMock
		a        *mock.AccountKeeperMock
		channelK *mock.ChannelKeeperMock
		ctx      sdk.Context
		chain    nexus.Chain
		req      *types.RetryIBCTransferRequest
		path     string
		transfer types.IBCTransfer
	)

	givenMessageServer := Given("a message server", func() {
		ctx, k, channelK = setup()
		k.InitGenesis(ctx, types.DefaultGenesisState())
		chain = nexustestutils.Chain()
		path = randomIBCPath()
		k.SetCosmosChain(ctx, types.CosmosChain{Name: chain.Name})
		funcs.MustNoErr(k.SetIBCPath(ctx, chain.Name, path))

		b = &mock.BankKeeperMock{}
		a = &mock.AccountKeeperMock{}
		i = &mock.IBCTransferKeeperMock{
			SendTransferFunc: func(sdk.Context, string, string, sdk.Coin, sdk.AccAddress, string, clienttypes.Height, uint64) error {
				return nil
			},
		}

		n = &mock.NexusMock{
			GetChainFunc: func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) {
				return chain, true
			},
			IsChainActivatedFunc: func(sdk.Context, nexus.Chain) bool { return true },
			EnqueueForTransferFunc: func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) (nexus.TransferID, error) {
				return nexus.TransferID(rand.I64Between(1, 9999)), nil
			},
		}
		channelK.GetNextSequenceSendFunc = func(sdk.Context, string, string) (uint64, bool) {
			return uint64(rand.I64Between(1, 99999)), true
		}
		channelK.GetChannelClientStateFunc = func(sdk.Context, string, string) (string, ibcclient.ClientState, error) {
			return "07-tendermint-0", axelartestutils.ClientState(), nil
		}
		ibcK := keeper.NewIBCKeeper(k, i, channelK)
		server = keeper.NewMsgServerImpl(k, n, b, i, a, ibcK)
	})

	requestIsMade := When("a retry failed transfer request is made", func() {
		req = types.NewRetryIBCTransferRequest(
			rand.AccAddr(),
			chain.Name,
			transfer.ID,
		)
	})

	whenTransferIsFailed := When("transfer is failed", func() {
		transfer = axelartestutils.RandomIBCTransfer()
		transfer.ChannelID = strings.Split(path, "/")[1]
		funcs.MustNoErr(k.EnqueueIBCTransfer(ctx, transfer))
		funcs.MustNoErr(k.SetTransferFailed(ctx, transfer.ID))
	})

	givenMessageServer.
		Branch(
			When("transfer is not found", func() {}).
				When2(requestIsMade).
				Then("should return error", func(t *testing.T) {
					_, err := server.RetryIBCTransfer(sdk.WrapSDKContext(ctx), req)
					assert.Error(t, err)
				}),

			When("transfer is not failed", func() {
				transfer := axelartestutils.RandomIBCTransfer()
				funcs.MustNoErr(k.EnqueueIBCTransfer(ctx, transfer))
				funcs.MustNoErr(k.SetTransferCompleted(ctx, transfer.ID))
			}).
				When2(requestIsMade).
				Then("should return error", func(t *testing.T) {
					_, err := server.RetryIBCTransfer(sdk.WrapSDKContext(ctx), req)
					assert.Error(t, err)
				}),

			When("ibc path does not match", func() {
				transfer := axelartestutils.RandomIBCTransfer()
				funcs.MustNoErr(k.EnqueueIBCTransfer(ctx, transfer))
				funcs.MustNoErr(k.SetTransferFailed(ctx, transfer.ID))
			}).
				When2(requestIsMade).
				Then("should return error", func(t *testing.T) {
					_, err := server.RetryIBCTransfer(sdk.WrapSDKContext(ctx), req)
					assert.Error(t, err)
				}),

			whenTransferIsFailed.
				When("ibc path matches", func() {}).
				When("send transfer succeeds", func() {}).
				When2(requestIsMade).
				Then("retry succeeds", func(t *testing.T) {
					_, err := server.RetryIBCTransfer(sdk.WrapSDKContext(ctx), req)
					assert.NoError(t, err)
					retiedTransfer, ok := k.GetTransfer(ctx, transfer.ID)
					assert.True(t, ok)
					assert.Equal(t, types.TransferPending, retiedTransfer.Status)
				}),
		).Run(t)

}

func randomMsgConfirmDeposit() *types.ConfirmDepositRequest {
	return types.NewConfirmDepositRequest(
		rand.AccAddr(),
		rand.Denom(5, 10),
		rand.AccAddr())
}

func randomTransfer(asset string, chain nexus.ChainName) nexus.CrossChainTransfer {
	hash := sha256.Sum256(rand.BytesBetween(20, 50))
	ranAddr := sdk.AccAddress(hash[:20]).String()

	return nexus.NewPendingCrossChainTransfer(
		mathRand.Uint64(),
		nexus.CrossChainAddress{
			Chain: nexus.Chain{
				Name:                  chain,
				SupportsForeignAssets: true,
				Module:                rand.Str(10),
			},
			Address: ranAddr,
		},
		sdk.NewInt64Coin(asset, rand.I64Between(1, 10000000000)),
	)
}
