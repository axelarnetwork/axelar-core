package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctypes "github.com/cosmos/ibc-go/v2/modules/apps/transfer/types"
	"github.com/stretchr/testify/assert"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	appParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/mock"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	nexustestutils "github.com/axelarnetwork/axelar-core/x/nexus/exported/testutils"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func setup() (sdk.Context, Keeper) {
	encCfg := appParams.MakeEncodingConfig()
	subspace := params.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey("axelarnetKey"), sdk.NewKVStoreKey("tAxelarnetKey"), "axelarnet")
	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())
	k := NewKeeper(encCfg.Codec, sdk.NewKVStoreKey("axelarnet"), subspace, &mock.ChannelKeeperMock{})

	return ctx, k
}

func TestCoin(t *testing.T) {
	var (
		ctx       sdk.Context
		nexusK    *mock.NexusMock
		bankK     *mock.BankKeeperMock
		transferK *mock.IBCTransferKeeperMock
		ibcK      IBCKeeper
		chain     nexus.Chain
		coin      coin
	)

	givenAKeeper := Given("a keeper", func() {
		ctx2, k := setup()
		ctx = ctx2
		nexusK = &mock.NexusMock{}
		bankK = &mock.BankKeeperMock{}
		transferK = &mock.IBCTransferKeeperMock{}
		ibcK = NewIBCKeeper(k, transferK, &mock.ChannelKeeperMock{})
		bankK.SendCoinsFromAccountToModuleFunc = func(sdk.Context, sdk.AccAddress, string, sdk.Coins) error {
			return nil
		}
		bankK.BurnCoinsFunc = func(sdk.Context, string, sdk.Coins) error {
			return nil
		}
		bankK.SendCoinsFunc = func(sdk.Context, sdk.AccAddress, sdk.AccAddress, sdk.Coins) error {
			return nil
		}
	})

	whenCoinIsNative := When("coin is native", func() {
		nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
			return exported.Axelarnet, true
		}
		coin = funcs.Must(newCoin(ctx, ibcK, nexusK, sdk.NewCoin("uaxl", sdk.NewInt(rand.PosI64()))))
	})

	whenCoinIsExternal := When("coin is external", func() {
		nexusK.GetChainByNativeAssetFunc = func(sdk.Context, string) (nexus.Chain, bool) {
			return nexustestutils.Chain(), true
		}
		nexusK.IsAssetRegisteredFunc = func(sdk.Context, nexus.Chain, string) bool {
			return true
		}
		coin = funcs.Must(newCoin(ctx, ibcK, nexusK, sdk.NewCoin(rand.Denom(5, 10), sdk.NewInt(rand.PosI64()))))
	})

	whenCoinIsICS20 := When("coin is from ICS20", func() {
		// setup
		path := testutils.RandomIBCPath()
		chain = nexustestutils.Chain()
		transferK.GetDenomTraceFunc = func(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) (ibctypes.DenomTrace, bool) {
			return ibctypes.DenomTrace{
				Path:      path,
				BaseDenom: rand.Denom(5, 10),
			}, true
		}

		ibcK.SetCosmosChain(ctx, types.CosmosChain{
			Name:       chain.Name,
			AddrPrefix: rand.StrBetween(1, 10),
			IBCPath:    path,
		})

		coin = funcs.Must(newCoin(ctx, ibcK, nexusK, sdk.NewCoin(testutils.RandomIBCDenom(), sdk.NewInt(rand.PosI64()))))
	})

	givenAKeeper.
		Branch(
			whenCoinIsNative.
				Then("coin type should be native", func(t *testing.T) {
					assert.Equal(t, types.CoinType(types.Native), coin.coinType)
				}),
			whenCoinIsExternal.
				Then("coin type should be external", func(t *testing.T) {
					assert.Equal(t, types.CoinType(types.External), coin.coinType)
				}),

			whenCoinIsICS20.
				Then("coin type should be ICS20", func(t *testing.T) {
					assert.Equal(t, types.CoinType(types.ICS20), coin.coinType)
				}),
		).Run(t)

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
