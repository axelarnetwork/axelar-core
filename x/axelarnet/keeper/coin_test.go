package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	"github.com/stretchr/testify/assert"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestCoin(t *testing.T) {
	var (
		ctx       sdk.Context
		nexusK    *mock.NexusMock
		bankK     *mock.BankKeeperMock
		transferK *mock.IBCTransferKeeperMock
		ibcK      keeper.IBCKeeper
		chain     nexus.Chain
		coin      keeper.Coin
	)

	givenAKeeper := Given("a keeper", func() {
		ctx2, k, _ := setup()
		ctx = ctx2
		nexusK = &mock.NexusMock{}
		bankK = &mock.BankKeeperMock{}
		transferK = &mock.IBCTransferKeeperMock{}
		ibcK = keeper.NewIBCKeeper(k, transferK, &mock.ChannelKeeperMock{})
		bankK.SendCoinsFromAccountToModuleFunc = func(sdk.Context, sdk.AccAddress, string, sdk.Coins) error {
			return nil
		}
		bankK.BurnCoinsFunc = func(sdk.Context, string, sdk.Coins) error {
			return nil
		}
		bankK.SendCoinsFunc = func(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error {
			return nil
		}
		bankK.GetBalanceFunc = func(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
			return coin.Coin
		}
	})

	whenCoinIsNative := When("coin is native", func() {
		nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
			return exported.Axelarnet, true
		}
		coin = funcs.Must(keeper.NewCoin(ctx, ibcK, nexusK, sdk.NewCoin("uaxl", sdk.NewInt(rand.PosI64()))))
	})

	whenCoinIsExternal := When("coin is external", func() {
		nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
			return nexustestutils.Chain(), true
		}
		nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool {
			return true
		}
		coin = funcs.Must(keeper.NewCoin(ctx, ibcK, nexusK, sdk.NewCoin(rand.Denom(5, 10), sdk.NewInt(rand.PosI64()))))
	})

	whenCoinIsICS20 := When("coin is from ICS20", func() {
		// setup
		path := testutils.RandomIBCPath()
		chain = nexustestutils.Chain()
		trace := ibctypes.DenomTrace{
			Path:      path,
			BaseDenom: rand.Denom(5, 10),
		}
		transferK.GetDenomTraceFunc = func(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctypes.DenomTrace, bool) {
			return trace, true
		}

		ibcK.SetCosmosChain(ctx, types.CosmosChain{
			Name:       chain.Name,
			AddrPrefix: rand.StrBetween(1, 10),
			IBCPath:    path,
		})

		coin = funcs.Must(keeper.NewCoin(ctx, ibcK, nexusK, sdk.NewCoin(trace.IBCDenom(), sdk.NewInt(rand.PosI64()))))

		bankK.GetBalanceFunc = func(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
			return sdk.NewCoin(trace.IBCDenom(), coin.Amount)
		}
	})

	givenAKeeper.
		Branch(
			whenCoinIsNative.
				Then("should Lock native coin in escrow account", func(t *testing.T) {
					err := coin.Lock(bankK, rand.AccAddr())
					assert.NoError(t, err)
					assert.Len(t, bankK.SendCoinsCalls(), 1)
				}),
			whenCoinIsExternal.
				Then("should burn external token", func(t *testing.T) {
					err := coin.Lock(bankK, rand.AccAddr())
					assert.NoError(t, err)
					assert.Len(t, bankK.SendCoinsFromAccountToModuleCalls(), 1)
					assert.Len(t, bankK.BurnCoinsCalls(), 1)
				}),

			whenCoinIsICS20.
				Then("should Lock ICS20 coin in escrow account", func(t *testing.T) {
					nexusK.GetChainFunc = func(sdk.Context, nexus.ChainName) (nexus.Chain, bool) {
						return chain, true
					}
					nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
						return chain, true
					}

					err := coin.Lock(bankK, rand.AccAddr())
					assert.NoError(t, err)
					assert.Len(t, bankK.SendCoinsCalls(), 1)
				}),
		).Run(t)
}
