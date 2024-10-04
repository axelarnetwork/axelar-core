package keeper

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types/testutils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/types/mock"
	"github.com/axelarnetwork/utils/funcs"
	. "github.com/axelarnetwork/utils/test"
)

func TestLockableCoin(t *testing.T) {
	var (
		ctx   sdk.Context
		nexus *mock.NexusMock
		ibc   *mock.IBCKeeperMock
		bank  *mock.BankKeeperMock

		coin  sdk.Coin
		trace ibctypes.DenomTrace
	)

	givenKeeper := Given("the nexus keeper", func() {
		ctx = sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

		nexus = &mock.NexusMock{}
		ibc = &mock.IBCKeeperMock{}
		bank = &mock.BankKeeperMock{}
	})

	whenCoinIsNative := When("coin is native", func() {
		coin = sdk.NewCoin(rand.Denom(5, 10), sdk.NewInt(rand.PosI64()))
		nexus.GetChainByNativeAssetFunc = func(ctx sdk.Context, asset string) (exported.Chain, bool) {
			if asset == coin.Denom {
				return axelarnet.Axelarnet, true
			}

			return exported.Chain{}, false
		}
	})

	whenCoinIsExternal := When("coin is external", func() {
		coin = sdk.NewCoin(rand.Denom(5, 10), sdk.NewInt(rand.PosI64()))
		nexus.GetChainByNativeAssetFunc = func(ctx sdk.Context, asset string) (exported.Chain, bool) {
			return exported.Chain{}, false
		}
		nexus.IsAssetRegisteredFunc = func(ctx sdk.Context, chain exported.Chain, asset string) bool {
			return chain == axelarnet.Axelarnet && asset == coin.Denom
		}
	})

	whenCoinIsICS20 := When("coin is ICS20", func() {
		path := testutils.RandomIBCPath()
		trace = ibctypes.DenomTrace{
			Path:      path,
			BaseDenom: rand.Denom(5, 10),
		}

		ibc.ParseIBCDenomFunc = func(ctx sdk.Context, ibcDenom string) (ibctypes.DenomTrace, error) {
			if ibcDenom == coin.Denom {
				return trace, nil
			}

			return ibctypes.DenomTrace{}, fmt.Errorf("denom not found")
		}
		ibc.GetIBCPathFunc = func(ctx sdk.Context, chain exported.ChainName) (string, bool) {
			if chain == axelarnet.Axelarnet.Name {
				return path, true
			}

			return "", false
		}
		nexus.GetChainByNativeAssetFunc = func(ctx sdk.Context, asset string) (exported.Chain, bool) {
			if asset == trace.BaseDenom {
				return axelarnet.Axelarnet, true
			}

			return exported.Chain{}, false
		}

		coin = sdk.NewCoin(trace.IBCDenom(), sdk.NewInt(rand.PosI64()))
	})

	t.Run("NewLockableCoin, GetCoin and GetOriginalCoin", func(t *testing.T) {
		givenKeeper.
			When2(whenCoinIsICS20).
			Then("should create a new lockable coin of type ICS20", func(t *testing.T) {
				lockableCoin, err := newLockableCoin(ctx, nexus, ibc, bank, coin)

				assert.NoError(t, err)
				assert.Equal(t, types.CoinType(types.ICS20), lockableCoin.coinType)
				assert.Equal(t, sdk.NewCoin(trace.GetBaseDenom(), coin.Amount), lockableCoin.GetCoin())
				assert.Equal(t, coin, lockableCoin.GetOriginalCoin(ctx))
			}).
			Run(t)

		givenKeeper.
			When2(whenCoinIsNative).
			Then("should create a new lockable coin of type native", func(t *testing.T) {
				lockableCoin, err := newLockableCoin(ctx, nexus, ibc, bank, coin)

				assert.NoError(t, err)
				assert.Equal(t, types.CoinType(types.Native), lockableCoin.coinType)
				assert.Equal(t, coin, lockableCoin.GetCoin())
				assert.Equal(t, coin, lockableCoin.GetOriginalCoin(ctx))
			}).
			Run(t)

		givenKeeper.
			When2(whenCoinIsExternal).
			Then("should create a new lockable coin of type external", func(t *testing.T) {
				lockableCoin, err := newLockableCoin(ctx, nexus, ibc, bank, coin)

				assert.NoError(t, err)
				assert.Equal(t, types.CoinType(types.External), lockableCoin.coinType)
				assert.Equal(t, coin, lockableCoin.GetCoin())
				assert.Equal(t, coin, lockableCoin.GetOriginalCoin(ctx))
			}).
			Run(t)
	})

	t.Run("LockFrom", func(t *testing.T) {
		givenKeeper.
			When2(whenCoinIsICS20).
			Then("should lock the coin", func(t *testing.T) {
				bank.SendCoinsFunc = func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error { return nil }

				lockableCoin := funcs.Must(newLockableCoin(ctx, nexus, ibc, bank, coin))
				fromAddr := rand.AccAddr()

				err := lockableCoin.LockFrom(ctx, fromAddr)

				assert.NoError(t, err)
				assert.Len(t, bank.SendCoinsCalls(), 1)
				assert.Equal(t, sdk.NewCoins(lockableCoin.GetOriginalCoin(ctx)), bank.SendCoinsCalls()[0].Amt)
				assert.Equal(t, fromAddr, bank.SendCoinsCalls()[0].FromAddr)
				assert.Equal(t, exported.GetEscrowAddress(lockableCoin.GetOriginalCoin(ctx).Denom), bank.SendCoinsCalls()[0].ToAddr)
			}).
			Run(t)

		givenKeeper.
			When2(whenCoinIsNative).
			Then("should lock the coin", func(t *testing.T) {
				bank.SendCoinsFunc = func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error { return nil }

				lockableCoin := funcs.Must(newLockableCoin(ctx, nexus, ibc, bank, coin))
				fromAddr := rand.AccAddr()

				err := lockableCoin.LockFrom(ctx, fromAddr)

				assert.NoError(t, err)
				assert.Len(t, bank.SendCoinsCalls(), 1)
				assert.Equal(t, sdk.NewCoins(lockableCoin.GetOriginalCoin(ctx)), bank.SendCoinsCalls()[0].Amt)
				assert.Equal(t, fromAddr, bank.SendCoinsCalls()[0].FromAddr)
				assert.Equal(t, exported.GetEscrowAddress(lockableCoin.GetOriginalCoin(ctx).Denom), bank.SendCoinsCalls()[0].ToAddr)
			}).
			Run(t)

		givenKeeper.
			When2(whenCoinIsExternal).
			Then("should burn the coin", func(t *testing.T) {
				bank.SendCoinsFromAccountToModuleFunc = func(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
					return nil
				}
				bank.BurnCoinsFunc = func(ctx sdk.Context, moduleName string, amt sdk.Coins) error { return nil }

				lockableCoin := funcs.Must(newLockableCoin(ctx, nexus, ibc, bank, coin))
				fromAddr := rand.AccAddr()

				err := lockableCoin.LockFrom(ctx, fromAddr)

				assert.NoError(t, err)
				assert.Len(t, bank.SendCoinsFromAccountToModuleCalls(), 1)
				assert.Equal(t, fromAddr, bank.SendCoinsFromAccountToModuleCalls()[0].SenderAddr)
				assert.Equal(t, types.ModuleName, bank.SendCoinsFromAccountToModuleCalls()[0].RecipientModule)
				assert.Equal(t, sdk.NewCoins(lockableCoin.GetOriginalCoin(ctx)), bank.SendCoinsFromAccountToModuleCalls()[0].Amt)
				assert.Len(t, bank.BurnCoinsCalls(), 1)
				assert.Equal(t, types.ModuleName, bank.BurnCoinsCalls()[0].ModuleName)
				assert.Equal(t, sdk.NewCoins(lockableCoin.GetOriginalCoin(ctx)), bank.BurnCoinsCalls()[0].Amt)
			}).
			Run(t)
	})

	t.Run("UnlockTo", func(t *testing.T) {
		givenKeeper.
			When2(whenCoinIsICS20).
			Then("should unlock the coin", func(t *testing.T) {
				bank.SendCoinsFunc = func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error { return nil }

				lockableCoin := funcs.Must(newLockableCoin(ctx, nexus, ibc, bank, coin))
				toAddr := rand.AccAddr()

				err := lockableCoin.UnlockTo(ctx, toAddr)

				assert.NoError(t, err)
				assert.Len(t, bank.SendCoinsCalls(), 1)
				assert.Equal(t, sdk.NewCoins(lockableCoin.GetOriginalCoin(ctx)), bank.SendCoinsCalls()[0].Amt)
				assert.Equal(t, exported.GetEscrowAddress(lockableCoin.GetOriginalCoin(ctx).Denom), bank.SendCoinsCalls()[0].FromAddr)
				assert.Equal(t, toAddr, bank.SendCoinsCalls()[0].ToAddr)
			}).
			Run(t)

		givenKeeper.
			When2(whenCoinIsNative).
			Then("should unlock the coin", func(t *testing.T) {
				bank.SendCoinsFunc = func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error { return nil }

				lockableCoin := funcs.Must(newLockableCoin(ctx, nexus, ibc, bank, coin))
				toAddr := rand.AccAddr()

				err := lockableCoin.UnlockTo(ctx, toAddr)

				assert.NoError(t, err)
				assert.Len(t, bank.SendCoinsCalls(), 1)
				assert.Equal(t, sdk.NewCoins(lockableCoin.GetOriginalCoin(ctx)), bank.SendCoinsCalls()[0].Amt)
				assert.Equal(t, exported.GetEscrowAddress(lockableCoin.GetOriginalCoin(ctx).Denom), bank.SendCoinsCalls()[0].FromAddr)
				assert.Equal(t, toAddr, bank.SendCoinsCalls()[0].ToAddr)
			}).
			Run(t)

		givenKeeper.
			When2(whenCoinIsExternal).
			Then("should mint the coin", func(t *testing.T) {
				bank.MintCoinsFunc = func(ctx sdk.Context, moduleName string, amt sdk.Coins) error { return nil }
				bank.SendCoinsFromModuleToAccountFunc = func(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
					return nil
				}

				lockableCoin := funcs.Must(newLockableCoin(ctx, nexus, ibc, bank, coin))
				toAddr := rand.AccAddr()

				err := lockableCoin.UnlockTo(ctx, toAddr)

				assert.NoError(t, err)
				assert.Len(t, bank.MintCoinsCalls(), 1)
				assert.Equal(t, types.ModuleName, bank.MintCoinsCalls()[0].ModuleName)
				assert.Equal(t, sdk.NewCoins(lockableCoin.GetOriginalCoin(ctx)), bank.MintCoinsCalls()[0].Amt)
				assert.Len(t, bank.SendCoinsFromModuleToAccountCalls(), 1)
				assert.Equal(t, types.ModuleName, bank.SendCoinsFromModuleToAccountCalls()[0].SenderModule)
				assert.Equal(t, toAddr, bank.SendCoinsFromModuleToAccountCalls()[0].RecipientAddr)
			}).
			Run(t)
	})
}
