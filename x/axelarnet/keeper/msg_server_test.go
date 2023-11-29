package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	mathRand "math/rand"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	ibctypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcclient "github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	axelartestutils "github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	evmtypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	evmtestutils "github.com/axelarnetwork/axelar-core/x/evm/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
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
		ctx, k, _, _ = setup()
		k.InitGenesis(ctx, types.DefaultGenesisState())
		nexusK = &mock.NexusMock{}
		ibcK := keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{})
		server = keeper.NewMsgServerImpl(k, nexusK, &mock.BankKeeperMock{}, &mock.AccountKeeperMock{}, ibcK)
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

	ibcPath := axelartestutils.RandomIBCPath()
	denomTrace := ibctypes.DenomTrace{
		Path:      ibcPath,
		BaseDenom: rand.Denom(5, 10),
	}

	chain := nexustestutils.RandomChain()
	givenMsgServer := Given("an axelarnet msg server", func() {
		ctx, k, _, _ = setup()
		k.InitGenesis(ctx, types.DefaultGenesisState())
		funcs.MustNoErr(k.SetCosmosChain(ctx, types.CosmosChain{
			Name:       chain.Name,
			AddrPrefix: rand.StrBetween(1, 10),
			IBCPath:    axelartestutils.RandomIBCPath(),
		}))

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
				return denomTrace, true
			},
		}
		ibcK := keeper.NewIBCKeeper(k, transferK)
		server = keeper.NewMsgServerImpl(k, nexusK, bankK, &mock.AccountKeeperMock{}, ibcK)
	})

	recipientIsFound := When("recipient is found", func() {
		nexusK.GetRecipientFunc = func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
			return nexus.CrossChainAddress{}, true
		}
	})

	whenDepositAddressHasBalance := When("deposit address has balance", func() {
		bankK.SpendableBalanceFunc = func(_ sdk.Context, _ sdk.AccAddress, denom string) sdk.Coin {
			// need to compare the balance so cannot make it random
			return sdk.NewCoin(denom, sdk.NewInt(1e18))
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

	enqueueTransferSucceeds := When("enqueue transfer succeeds", func() {
		nexusK.EnqueueForTransferFunc = func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) (nexus.TransferID, error) {
			return nexus.TransferID(rand.I64Between(1, 9999)), nil
		}
	})

	confirmExternalICS20TokenRequest := When("a confirm external ICS20 token deposit request is made", func() {
		req = randomMsgConfirmDeposit()
		req.Denom = denomTrace.IBCDenom()
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
					bankK.SpendableBalanceFunc = func(_ sdk.Context, _ sdk.AccAddress, denom string) sdk.Coin {
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
		ctx, k, _, _ = setup()
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
		ibcK := keeper.NewIBCKeeper(k, transferK)
		server = keeper.NewMsgServerImpl(k, nexusK, bankK, accountK, ibcK)
	})

	whenAssetOriginsFromExternalCosmosChain := When("asset is from external cosmos chain", func() {
		chain := nexustestutils.RandomChain()
		funcs.MustNoErr(k.SetCosmosChain(ctx, types.CosmosChain{
			Name:       chain.Name,
			AddrPrefix: rand.StrBetween(1, 10),
			IBCPath:    axelartestutils.RandomIBCPath(),
		}))
		nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
			return chain, true
		}
	})

	hasPendingTransfers := When("has pending transfers", func() {
		nexusK.GetTransfersForChainPaginatedFunc = func(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState, pageRequest *query.PageRequest) ([]nexus.CrossChainTransfer, *query.PageResponse, error) {
			return []nexus.CrossChainTransfer{randomTransfer(rand.Denom(3, 10), nexus.ChainName(rand.StrBetween(2, 10)))}, nil, nil
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
					nexusK.GetTransfersForChainPaginatedFunc = func(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState, pageRequest *query.PageRequest) ([]nexus.CrossChainTransfer, *query.PageResponse, error) {
						return []nexus.CrossChainTransfer{}, nil, nil
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
						return nexustestutils.RandomChain(), true
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

				When("asset is registered", func() {
					nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool {
						return true
					}
					nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
						return nexustestutils.RandomChain(), true
					}
				}).
					When("has many pending transfers", func() {
						nexusK.GetTransfersForChainPaginatedFunc = func(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState, pageRequest *query.PageRequest) ([]nexus.CrossChainTransfer, *query.PageResponse, error) {
							return slices.Expand(func(int) nexus.CrossChainTransfer {
								return randomTransfer(rand.Denom(3, 10), nexus.ChainName(rand.StrBetween(2, 10)))
							}, int(pageRequest.Limit)), nil, nil
						}
					}).
					When("mint coins succeeds", func() {
						bankK.MintCoinsFunc = func(sdk.Context, string, sdk.Coins) error {
							return nil
						}
					}).
					When2(sendCoinSucceeds).
					When2(requestIsMade).
					Then("mint coin and archive the transfer", func(t *testing.T) {
						transferLimit := int(k.GetTransferLimit(ctx))
						_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Len(t, bankK.MintCoinsCalls(), transferLimit)
						assert.Len(t, bankK.SendCoinsCalls(), transferLimit)
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), transferLimit)
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
		ctx, k, _, _ = setup()
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
			funcs.MustNoErr(k.SetCosmosChain(ctx, c))
		})

		nexusK = &mock.NexusMock{
			GetChainFunc: func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
				return nexus.Chain{Name: chain}, true
			},
			GetChainByNativeAssetFunc: func(sdk.Context, string) (nexus.Chain, bool) {
				return nexustestutils.RandomChain(), true
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
		ibcK := keeper.NewIBCKeeper(k, transferK)
		server = keeper.NewMsgServerImpl(k, nexusK, bankK, accountK, ibcK)
	})

	whenAssetOriginsFromExternalCosmosChain := When("asset is from external cosmos chain", func() {
		nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
			chainName := cosmosChains[rand.I64Between(0, int64(len(cosmosChains)))].Name
			return nexus.Chain{Name: chainName}, true
		}

	})
	hasPendingTranfers := When("has pending transfers", func() {
		nexusK.GetTransfersForChainPaginatedFunc = func(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState, pageRequest *query.PageRequest) ([]nexus.CrossChainTransfer, *query.PageResponse, error) {
			var transfers []nexus.CrossChainTransfer
			for i := int64(0); i < rand.I64Between(1, 5); i++ {
				chainName := chain.Name
				transfers = append(transfers, randomTransfer(rand.Denom(3, 10), chainName))
			}
			transfersNum += len(transfers)
			return transfers, nil, nil
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
					nexusK.GetTransfersForChainPaginatedFunc = func(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState, pageRequest *query.PageRequest) ([]nexus.CrossChainTransfer, *query.PageResponse, error) {
						return []nexus.CrossChainTransfer{}, nil, nil
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
		ctx, k, channelK, _ = setup()
		k.InitGenesis(ctx, types.DefaultGenesisState())
		cosmosChain := axelartestutils.RandomCosmosChain()
		chain = nexus.Chain{
			Name:                  cosmosChain.Name,
			SupportsForeignAssets: true,
			Module:                types.ModuleName,
		}
		cosmosChain.Name = chain.Name
		path = cosmosChain.IBCPath
		funcs.MustNoErr(k.SetCosmosChain(ctx, cosmosChain))
		funcs.MustNoErr(k.SetChainByIBCPath(ctx, path, cosmosChain.Name))

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

		ibcK := keeper.NewIBCKeeper(k, i)
		server = keeper.NewMsgServerImpl(k, n, b, a, ibcK)
	})

	requestIsMade := When("a retry failed transfer request is made", func() {
		req = types.NewRetryIBCTransferRequest(
			rand.AccAddr(),
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

			whenTransferIsFailed.
				When("chain is not activated", func() {
					n.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool { return false }
				}).
				When2(requestIsMade).
				Then("should return error", func(t *testing.T) {
					_, err := server.RetryIBCTransfer(sdk.WrapSDKContext(ctx), req)
					assert.ErrorContains(t, err, "not activated")
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

func TestAddCosmosBasedChain(t *testing.T) {
	var (
		server types.MsgServiceServer
		k      keeper.Keeper
		nexusK *mock.NexusMock
		ctx    sdk.Context
		req    *types.AddCosmosBasedChainRequest
	)
	repeats := 20

	givenMsgServer := Given("an axelarnet msg server", func() {
		ctx, k, _, _ = setup()
		k.InitGenesis(ctx, types.DefaultGenesisState())
		nexusK = &mock.NexusMock{
			GetChainFunc:              func(ctx sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) { return nexus.Chain{}, false },
			GetChainByNativeAssetFunc: func(ctx sdk.Context, asset string) (nexus.Chain, bool) { return nexus.Chain{}, false },
			SetChainFunc:              func(ctx sdk.Context, chain nexus.Chain) {},
			RegisterAssetFunc: func(ctx sdk.Context, chain nexus.Chain, asset nexus.Asset, limit sdk.Uint, window time.Duration) error {
				return nil
			},
		}
		ibcK := keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{})
		server = keeper.NewMsgServerImpl(k, nexusK, &mock.BankKeeperMock{}, &mock.AccountKeeperMock{}, ibcK)
	})

	addChainRequest := When("an add cosmos based chain request is created", func() {
		req = types.NewAddCosmosBasedChainRequest(
			rand.AccAddr(),
			rand.StrBetween(1, 20),
			rand.StrBetween(1, 10),
			slices.Expand(func(idx int) nexus.Asset { return nexus.NewAsset(rand.Denom(3, 10), true) }, int(rand.I64Between(0, 5))),
			axelartestutils.RandomIBCPath(),
		)
	})

	requestFails := func(msg string) ThenStatement {
		return Then("add cosmos chain request fails", func(t *testing.T) {
			_, err := server.AddCosmosBasedChain(sdk.WrapSDKContext(ctx), req)
			assert.ErrorContains(t, err, msg)
		})
	}

	validationFails := func(msg string) ThenStatement {
		return Then("add cosmos chain validation fails", func(t *testing.T) {
			err := req.ValidateBasic()
			assert.ErrorContains(t, err, msg)
		})
	}

	givenMsgServer.
		When2(addChainRequest).
		Branch(
			When("chain name is invalid", func() {
				req.CosmosChain = "invalid_name"
			}).
				Then2(validationFails("invalid cosmos chain name")),

			When("invalid addr prefix", func() {
				req.AddrPrefix = "invalid_prefix"
			}).
				Then2(validationFails("invalid address prefix")),

			When("invalid asset", func() {
				req.NativeAssets = []nexus.Asset{{Denom: "invalid_asset", IsNativeAsset: true}}
			}).
				Then2(validationFails("invalid denomination")),

			When("invalid asset denom", func() {
				req.NativeAssets = []nexus.Asset{{Denom: "invalid@denom", IsNativeAsset: true}}
			}).
				Then2(validationFails("invalid denomination")),

			When("duplicate assets", func() {
				asset := nexus.Asset{Denom: rand.Denom(3, 10), IsNativeAsset: true}
				req.NativeAssets = []nexus.Asset{asset, asset}
			}).
				Then2(validationFails("duplicate asset")),

			When("invalid ibc path", func() {
				req.IBCPath = "invalid path"
			}).Then2(validationFails("invalid IBC path")),

			When("non native asset", func() {
				req.NativeAssets = []nexus.Asset{{Denom: rand.Denom(3, 10), IsNativeAsset: false}}
			}).
				Then2(validationFails("is not specified as a native asset")),
		).
		Run(t, repeats)

	givenMsgServer.
		When2(addChainRequest).
		Branch(
			When("chain is already registered", func() {
				nexusK.GetChainFunc = func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
					return nexus.Chain{
						Name:                  chain,
						SupportsForeignAssets: true,
						Module:                rand.Str(10),
					}, true
				}
			}).
				Then2(requestFails("already registered")),

			When("asset is already registered", func() {
				req.NativeAssets = []nexus.Asset{{Denom: rand.Denom(3, 10), IsNativeAsset: true}}
				nexusK.RegisterAssetFunc = func(ctx sdk.Context, chain nexus.Chain, asset nexus.Asset, limit sdk.Uint, window time.Duration) error {
					return fmt.Errorf("asset already registered")
				}
			}).
				Then2(requestFails("asset already registered")),
		).
		Run(t, repeats)

	givenMsgServer.
		When2(addChainRequest).
		Then("chain is added", func(t *testing.T) {
			_, err := server.AddCosmosBasedChain(sdk.WrapSDKContext(ctx), req)
			assert.NoError(t, err)

			chain, ok := k.GetCosmosChainByName(ctx, req.CosmosChain)
			assert.True(t, ok)
			assert.Equal(t, req.CosmosChain, chain.Name)
			assert.Equal(t, req.AddrPrefix, chain.AddrPrefix)
		}).
		Run(t, repeats)
}

func TestRouteMessage(t *testing.T) {
	var (
		server types.MsgServiceServer
		nexusK *mock.NexusMock
		ctx    sdk.Context
	)

	req := types.RouteMessageRequest{
		ID:         rand.Str(10),
		Sender:     rand.AccAddr(),
		Feegranter: rand.AccAddr(),
		Payload:    rand.BytesBetween(5, 100),
	}

	givenMsgServer := Given("an axelarnet msg server", func() {
		c, k, _, _ := setup()
		ctx = c

		nexusK = &mock.NexusMock{}
		ibcK := keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{})
		bankK := &mock.BankKeeperMock{}
		accountK := &mock.AccountKeeperMock{}
		server = keeper.NewMsgServerImpl(k, nexusK, bankK, accountK, ibcK)
	})

	givenMsgServer.
		When("route message successfully", func() {
			nexusK.RouteMessageFunc = func(_ sdk.Context, _ string, _ ...nexus.RoutingContext) error { return nil }
		}).
		Then("should route the correct message", func(t *testing.T) {
			_, err := server.RouteMessage(sdk.WrapSDKContext(ctx), &req)

			assert.NoError(t, err)
			assert.Len(t, nexusK.RouteMessageCalls(), 1)
			assert.Equal(t, nexusK.RouteMessageCalls()[0].RoutingCtx[0].Sender, req.Sender)
			assert.Equal(t, nexusK.RouteMessageCalls()[0].RoutingCtx[0].FeeGranter, req.Feegranter)
			assert.Equal(t, nexusK.RouteMessageCalls()[0].RoutingCtx[0].Payload, req.Payload)
			assert.Equal(t, nexusK.RouteMessageCalls()[0].ID, req.ID)
		}).
		Run(t)
}

func TestHandleCallContract(t *testing.T) {
	var (
		server types.MsgServiceServer
		k      keeper.Keeper
		nexusK *mock.NexusMock
		b      *mock.BankKeeperMock
		ctx    sdk.Context
		req    *types.CallContractRequest
		msg    nexus.GeneralMessage
	)

	givenMsgServer := Given("an axelarnet msg server", func() {
		ctx, k, _, _ = setup()
		k.InitGenesis(ctx, types.DefaultGenesisState())
		nexusK = &mock.NexusMock{}
		ibcK := keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{})
		b = &mock.BankKeeperMock{}
		server = keeper.NewMsgServerImpl(k, nexusK, b, &mock.AccountKeeperMock{}, ibcK)
		count := 0
		nexusK.GenerateMessageIDFunc = func(ctx sdk.Context) (string, []byte, uint64) {
			count++
			hash := sha256.Sum256(ctx.TxBytes())
			return fmt.Sprintf("%s-%x", hex.EncodeToString(hash[:]), count), hash[:], uint64(count)
		}
		b.SendCoinsFunc = func(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error { return nil }
		nexusK.GetChainByNativeAssetFunc = func(_ sdk.Context, asset string) (nexus.Chain, bool) { return exported.Axelarnet, true }
	})

	whenChainIsRegistered := When("chain is registered", func() {
		nexusK.GetChainFunc = func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
			return nexus.Chain{
				Name:                  chain,
				SupportsForeignAssets: true,
				Module:                evmtypes.ModuleName,
				KeyType:               tss.Multisig,
			}, true
		}
	})

	whenChainIsActivated := When("chain is activated", func() {
		nexusK.IsChainActivatedFunc = func(_ sdk.Context, chain nexus.Chain) bool {
			return true
		}
	})

	whenAddressIsValid := When("address is valid", func() {
		nexusK.ValidateAddressFunc = func(_ sdk.Context, address nexus.CrossChainAddress) error {
			return nil
		}
	})

	whenSetNewMessageSucceeds := When("set new message succeeds", func() {
		nexusK.SetNewMessageFunc = func(_ sdk.Context, m nexus.GeneralMessage) error {
			msg = m
			return m.ValidateBasic()
		}
	})

	requestIsMade := When("a call contract request is made", func() {
		req = types.NewCallContractRequest(
			rand.AccAddr(),
			nexustestutils.RandomChain().Name.String(),
			evmtestutils.RandomAddress().Hex(),
			rand.BytesBetween(5, 1000),
			&types.Fee{Amount: rand.Coin(), Recipient: rand.AccAddr()})
	})

	callFails := Then("call contract request fails", func(t *testing.T) {
		_, err := server.CallContract(sdk.WrapSDKContext(ctx), req)
		assert.Error(t, err)
	})

	validationFails := Then("call contract validation fails", func(t *testing.T) {
		err := req.ValidateBasic()
		assert.Error(t, err)
	})

	t.Run("call contract", func(t *testing.T) {
		givenMsgServer.
			Branch(
				whenChainIsRegistered.
					When2(whenChainIsActivated).
					When2(whenAddressIsValid).
					When2(whenSetNewMessageSucceeds).
					When2(requestIsMade).
					Then("call contract succeeds", func(t *testing.T) {
						err := req.ValidateBasic()
						assert.NoError(t, err)
						_, err = server.CallContract(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Equal(t, msg.Status, nexus.Approved)
						assert.Equal(t, msg.GetSourceChain(), nexus.ChainName(exported.Axelarnet.Name))
						assert.Equal(t, msg.GetSourceAddress(), req.Sender.String())
						assert.Equal(t, msg.GetDestinationAddress(), req.ContractAddress)
						assert.Equal(t, msg.GetDestinationChain(), req.Chain)

						payloadHash := crypto.Keccak256(req.Payload)
						assert.Equal(t, msg.PayloadHash, payloadHash)
					}),
				whenChainIsRegistered.
					When2(whenChainIsActivated).
					When2(whenAddressIsValid).
					When2(whenSetNewMessageSucceeds).
					When2(requestIsMade).
					When("destination is cosmos", func() {

						nexusK.GetChainFunc = func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
							return nexus.Chain{
								Name:                  chain,
								SupportsForeignAssets: true,
								Module:                exported.ModuleName,
								KeyType:               tss.Multisig,
							}, true
						}
					}).
					Then("call contract succeeds", func(t *testing.T) {
						err := req.ValidateBasic()
						assert.NoError(t, err)
						_, err = server.CallContract(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Equal(t, msg.Status, nexus.Approved)
						assert.Equal(t, msg.GetSourceChain(), nexus.ChainName(exported.Axelarnet.Name))
						assert.Equal(t, msg.GetSourceAddress(), req.Sender.String())
						assert.Equal(t, msg.GetDestinationAddress(), req.ContractAddress)
						assert.Equal(t, msg.GetDestinationChain(), req.Chain)

						payloadHash := crypto.Keccak256(req.Payload)
						assert.Equal(t, msg.PayloadHash, payloadHash)
					}),
				whenChainIsRegistered.
					When2(whenChainIsActivated).
					When2(whenAddressIsValid).
					When2(whenSetNewMessageSucceeds).
					When2(requestIsMade).
					When("fee is nil", func() {
						req.Fee = nil
					}).
					Then("call contract succeeds", func(t *testing.T) {
						err := req.ValidateBasic()
						assert.NoError(t, err)
						_, err = server.CallContract(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
						assert.Equal(t, msg.Status, nexus.Approved)
						assert.Equal(t, msg.GetSourceChain(), nexus.ChainName(exported.Axelarnet.Name))
						assert.Equal(t, msg.GetSourceAddress(), req.Sender.String())
						assert.Equal(t, msg.GetDestinationAddress(), req.ContractAddress)
						assert.Equal(t, msg.GetDestinationChain(), req.Chain)

						payloadHash := crypto.Keccak256(req.Payload)
						assert.Equal(t, msg.PayloadHash, payloadHash)
					}),
				whenChainIsRegistered.
					When2(whenChainIsActivated).
					When2(whenAddressIsValid).
					When2(requestIsMade).
					When("fee is zero", func() {
						req.Fee.Amount.Amount = sdk.NewInt(0)
					}).
					Then2(validationFails),
				whenChainIsRegistered.
					When2(whenChainIsActivated).
					When2(whenAddressIsValid).
					When("set new message fails", func() {
						nexusK.SetNewMessageFunc = func(_ sdk.Context, m nexus.GeneralMessage) error {
							return fmt.Errorf("failed to set message")
						}
					}).
					Then2(callFails),
				whenChainIsRegistered.
					When2(whenChainIsActivated).
					When("address is not valid", func() {
						nexusK.ValidateAddressFunc = func(_ sdk.Context, address nexus.CrossChainAddress) error {
							return fmt.Errorf("address is invalid")
						}
					}).
					Then2(callFails),
				whenChainIsRegistered.
					When("chain is not activated", func() {
						nexusK.IsChainActivatedFunc = func(_ sdk.Context, chain nexus.Chain) bool { return false }
					}).
					Then2(callFails),
				When("chain is not registered", func() {
					nexusK.GetChainFunc = func(_ sdk.Context, chain nexus.ChainName) (nexus.Chain, bool) {
						return nexus.Chain{}, false
					}
				}).
					Then2(callFails),
			).Run(t)
	})
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
