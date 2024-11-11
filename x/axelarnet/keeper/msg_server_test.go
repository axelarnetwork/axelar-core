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
	nexusmock "github.com/axelarnetwork/axelar-core/x/nexus/exported/mock"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	nexustypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
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
		server = keeper.NewMsgServerImpl(k, nexusK, &mock.BankKeeperMock{}, ibcK)
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
		server = keeper.NewMsgServerImpl(k, nexusK, bankK, ibcK)
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

	enqueueTransferSucceeds := When("enqueue transfer succeeds", func() {
		nexusK.EnqueueForTransferFunc = func(sdk.Context, nexus.CrossChainAddress, sdk.Coin) (nexus.TransferID, error) {
			return nexus.TransferID(rand.I64Between(1, 9999)), nil
		}
	})

	confirmToken := When("a confirm token deposit request is made", func() {
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
					When2(confirmToken).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When("recipient is not found", func() {
						nexusK.GetRecipientFunc = func(sdk.Context, nexus.CrossChainAddress) (nexus.CrossChainAddress, bool) {
							return nexus.CrossChainAddress{}, false
						}
					}).
					When2(confirmToken).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When("chain is not activated", func() {
						nexusK.IsChainActivatedFunc = func(sdk.Context, nexus.Chain) bool { return false }
					}).
					When2(confirmToken).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When("fails to create a lockable asset", func() {
						nexusK.NewLockableAssetFunc = func(ctx sdk.Context, ibc nexustypes.IBCKeeper, bank nexustypes.BankKeeper, coin sdk.Coin) (nexus.LockableAsset, error) {
							return nil, fmt.Errorf("failed to create lockable asset")
						}
					}).
					When2(confirmToken).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When("suceeded to create a lockable asset but lock fails", func() {
						nexusK.NewLockableAssetFunc = func(ctx sdk.Context, ibc nexustypes.IBCKeeper, bank nexustypes.BankKeeper, coin sdk.Coin) (nexus.LockableAsset, error) {
							lockbleCoin := &nexusmock.LockableAssetMock{
								LockFromFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress) error {
									return fmt.Errorf("failed to lock coin")
								},
							}

							return lockbleCoin, nil
						}
					}).
					When2(confirmToken).
					Then2(confirmDepositFails),

				whenDepositAddressHasBalance.
					When2(recipientIsFound).
					When2(chainIsActivated).
					When("succeeded to create a lockable asset but lock succeeds", func() {
						nexusK.NewLockableAssetFunc = func(ctx sdk.Context, ibc nexustypes.IBCKeeper, bank nexustypes.BankKeeper, coin sdk.Coin) (nexus.LockableAsset, error) {
							lockableAsset := &nexusmock.LockableAssetMock{
								LockFromFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress) error {
									return nil
								},
								GetAssetFunc: func() sdk.Coin {
									return sdk.NewCoin(req.Denom, sdk.NewInt(1e18))
								},
							}

							return lockableAsset, nil
						}
					}).
					When2(enqueueTransferSucceeds).
					When2(confirmToken).
					Then("confirm deposit succeeds", func(t *testing.T) {
						_, err := server.ConfirmDeposit(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)
					}),
			).Run(t)
	})
}

func TestHandleMsgExecutePendingTransfers(t *testing.T) {
	var (
		server        types.MsgServiceServer
		k             keeper.Keeper
		nexusK        *mock.NexusMock
		bankK         *mock.BankKeeperMock
		transferK     *mock.IBCTransferKeeperMock
		ctx           sdk.Context
		req           *types.ExecutePendingTransfersRequest
		lockableAsset *nexusmock.LockableAssetMock
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
			SubTransferFeeFunc:       func(sdk.Context, sdk.Coin) {},
			MarkTransferAsFailedFunc: func(sdk.Context, nexus.CrossChainTransfer) {},
		}
		bankK = &mock.BankKeeperMock{}
		transferK = &mock.IBCTransferKeeperMock{}
		ibcK := keeper.NewIBCKeeper(k, transferK)
		server = keeper.NewMsgServerImpl(k, nexusK, bankK, ibcK)
	})

	whenHasPendingTransfers := When("has pending transfers", func() {
		nexusK.GetTransfersForChainPaginatedFunc = func(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState, pageRequest *query.PageRequest) ([]nexus.CrossChainTransfer, *query.PageResponse, error) {
			return []nexus.CrossChainTransfer{randomTransfer(rand.Denom(3, 10), nexus.ChainName(rand.StrBetween(2, 10)))}, nil, nil
		}
	})

	unlocksCoinSucceeds := When("unlock coins succeeds", func() {
		lockableAsset = &nexusmock.LockableAssetMock{
			UnlockToFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress) error {
				return nil
			},
		}

		nexusK.NewLockableAssetFunc = func(ctx sdk.Context, ibc nexustypes.IBCKeeper, bank nexustypes.BankKeeper, coin sdk.Coin) (nexus.LockableAsset, error) {
			return lockableAsset, nil
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
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), 0)
					}),

				whenHasPendingTransfers.
					When("unlock coins fails", func() {
						lockableAsset = &nexusmock.LockableAssetMock{
							UnlockToFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress) error {
								return fmt.Errorf("failed to unlock coin")
							},
						}
						nexusK.NewLockableAssetFunc = func(ctx sdk.Context, ibc nexustypes.IBCKeeper, bank nexustypes.BankKeeper, coin sdk.Coin) (nexus.LockableAsset, error) {
							return lockableAsset, nil
						}
					}).
					When2(requestIsMade).
					Then("should mark the transfer as failed", func(t *testing.T) {
						_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)

						assert.Len(t, nexusK.ArchivePendingTransferCalls(), 0)
						assert.Len(t, nexusK.MarkTransferAsFailedCalls(), 1)
						assert.Len(t, ctx.EventManager().Events(), 0)
					}),

				whenHasPendingTransfers.
					When2(unlocksCoinSucceeds).
					When2(requestIsMade).
					Then("archive the transfer", func(t *testing.T) {
						_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), req)
						assert.NoError(t, err)

						assert.Len(t, nexusK.ArchivePendingTransferCalls(), 1)
						assert.Len(t, lockableAsset.UnlockToCalls(), 1)
					}),

				When("has many pending transfers", func() {
					nexusK.GetTransfersForChainPaginatedFunc = func(ctx sdk.Context, chain nexus.Chain, state nexus.TransferState, pageRequest *query.PageRequest) ([]nexus.CrossChainTransfer, *query.PageResponse, error) {
						return slices.Expand(func(int) nexus.CrossChainTransfer {
							return randomTransfer(rand.Denom(3, 10), nexus.ChainName(rand.StrBetween(2, 10)))
						}, int(pageRequest.Limit)), nil, nil
					}
				}).
					When2(unlocksCoinSucceeds).
					When2(requestIsMade).
					Then("mint coin and archive the transfer", func(t *testing.T) {
						transferLimit := int(k.GetTransferLimit(ctx))
						_, err := server.ExecutePendingTransfers(sdk.WrapSDKContext(ctx), req)

						assert.NoError(t, err)
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), transferLimit)
						assert.Len(t, lockableAsset.UnlockToCalls(), transferLimit)
					}),
			).Run(t)
	})
}

func TestHandleMsgRouteIBCTransfers(t *testing.T) {
	var (
		server        types.MsgServiceServer
		k             keeper.Keeper
		nexusK        *mock.NexusMock
		bankK         *mock.BankKeeperMock
		transferK     *mock.IBCTransferKeeperMock
		ctx           sdk.Context
		req           *types.RouteIBCTransfersRequest
		cosmosChains  []types.CosmosChain
		transfersNum  int
		lockableAsset *nexusmock.LockableAssetMock
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
		ibcK := keeper.NewIBCKeeper(k, transferK)
		server = keeper.NewMsgServerImpl(k, nexusK, bankK, ibcK)
	})

	whenHasPendingTranfers := When("has pending transfers", func() {
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

				whenHasPendingTranfers.
					When2(requestIsMade).
					When("unlock coin succeeds", func() {
						lockableAsset = &nexusmock.LockableAssetMock{
							UnlockToFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress) error {
								return nil
							},
							GetCoinFunc: func(ctx sdk.Context) sdk.Coin {
								return sdk.NewCoin(rand.Denom(3, 10), sdk.NewInt(1e18))
							},
						}

						nexusK.NewLockableAssetFunc = func(ctx sdk.Context, ibc nexustypes.IBCKeeper, bank nexustypes.BankKeeper, coin sdk.Coin) (nexus.LockableAsset, error) {
							return lockableAsset, nil
						}
					}).
					Then("archive the transfer", func(t *testing.T) {
						_, err := server.RouteIBCTransfers(sdk.WrapSDKContext(ctx), req)

						assert.NoError(t, err)
						assert.Len(t, nexusK.ArchivePendingTransferCalls(), transfersNum, fmt.Sprintf("expected %d got %d", transfersNum, len(nexusK.ArchivePendingTransferCalls())))
						assert.Len(t, lockableAsset.UnlockToCalls(), transfersNum)
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
			NewLockableAssetFunc: func(ctx sdk.Context, ibc nexustypes.IBCKeeper, bank nexustypes.BankKeeper, coin sdk.Coin) (nexus.LockableAsset, error) {
				lockableAsset := &nexusmock.LockableAssetMock{
					UnlockToFunc: func(ctx sdk.Context, fromAddr sdk.AccAddress) error {
						return nil
					},
				}

				return lockableAsset, nil
			},
		}

		channelK.GetNextSequenceSendFunc = func(sdk.Context, string, string) (uint64, bool) {
			return uint64(rand.I64Between(1, 99999)), true
		}
		channelK.GetChannelClientStateFunc = func(sdk.Context, string, string) (string, ibcclient.ClientState, error) {
			return "07-tendermint-0", axelartestutils.ClientState(), nil
		}

		ibcK := keeper.NewIBCKeeper(k, i)
		server = keeper.NewMsgServerImpl(k, n, b, ibcK)
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
		server = keeper.NewMsgServerImpl(k, nexusK, &mock.BankKeeperMock{}, ibcK)
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
		server = keeper.NewMsgServerImpl(k, nexusK, bankK, ibcK)
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
		nexusK = &mock.NexusMock{
			NewLockableAssetFunc: func(ctx sdk.Context, ibc nexustypes.IBCKeeper, bank nexustypes.BankKeeper, coin sdk.Coin) (nexus.LockableAsset, error) {
				lockableAsset := &nexusmock.LockableAssetMock{
					GetAssetFunc: func() sdk.Coin { return req.Fee.Amount },
				}

				return lockableAsset, nil
			},
		}
		ibcK := keeper.NewIBCKeeper(k, &mock.IBCTransferKeeperMock{})
		b = &mock.BankKeeperMock{}
		server = keeper.NewMsgServerImpl(k, nexusK, b, ibcK)
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
