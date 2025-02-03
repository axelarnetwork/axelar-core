package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/testutils/fake"
	"github.com/axelarnetwork/axelar-core/testutils/rand"
	axelarnettypes "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/distribution/keeper"
	"github.com/axelarnetwork/axelar-core/x/distribution/types"
	"github.com/axelarnetwork/axelar-core/x/distribution/types/mock"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
	. "github.com/axelarnetwork/utils/test"
)

func TestAllocateTokens(t *testing.T) {
	var (
		k           keeper.Keeper
		accBalances map[string]sdk.Coins
		bk          *mock.BankKeeperMock
		fees        sdk.Coins
	)

	ctx := sdk.NewContext(fake.NewMultiStore(), tmproto.Header{}, false, log.TestingLogger())

	fees = sdk.NewCoins(sdk.NewCoin(axelarnettypes.NativeAsset, sdk.NewInt(rand.PosI64())))
	accBalances = map[string]sdk.Coins{
		authtypes.NewModuleAddress(authtypes.FeeCollectorName).String(): fees,
	}

	Given("an axelar distribution keeper", func() {
		encCfg := params.MakeEncodingConfig()
		subspace := paramstypes.NewSubspace(encCfg.Codec, encCfg.Amino, sdk.NewKVStoreKey(distributiontypes.StoreKey), sdk.NewKVStoreKey("tKey"), distributiontypes.ModuleName)
		ak := &mock.AccountKeeperMock{
			GetModuleAccountFunc: func(ctx sdk.Context, name string) authtypes.ModuleAccountI {
				return authtypes.NewEmptyModuleAccount(name)
			},
			GetModuleAddressFunc: func(name string) sdk.AccAddress {
				return authtypes.NewModuleAddress(name)
			},
		}
		bk = &mock.BankKeeperMock{
			GetAllBalancesFunc: func(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
				return accBalances[addr.String()]
			},
			SendCoinsFromModuleToModuleFunc: func(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
				senderModule = authtypes.NewModuleAddress(senderModule).String()
				recipientModule = authtypes.NewModuleAddress(recipientModule).String()

				accBalances[senderModule] = accBalances[senderModule].Sub(amt)
				accBalances[recipientModule] = accBalances[recipientModule].Add(amt...)

				return nil
			},
			SendCoinsFromModuleToAccountFunc: func(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
				senderModule = authtypes.NewModuleAddress(senderModule).String()

				accBalances[senderModule] = accBalances[senderModule].Sub(amt)
				accBalances[recipientAddr.String()] = accBalances[recipientAddr.String()].Add(amt...)

				return nil
			},
			BurnCoinsFunc: func(ctx sdk.Context, name string, amt sdk.Coins) error {
				acc := authtypes.NewModuleAddress(name).String()
				accBalances[acc] = accBalances[acc].Sub(amt)

				return nil
			},
			MintCoinsFunc: func(ctx sdk.Context, name string, amt sdk.Coins) error {
				acc := authtypes.NewModuleAddress(name).String()
				accBalances[acc] = accBalances[acc].Add(amt...)

				return nil
			},
		}
		sk := &mock.StakingKeeperMock{
			ValidatorByConsAddrFunc: func(ctx sdk.Context, addr sdk.ConsAddress) stakingtypes.ValidatorI {
				seed := []byte("key")
				consKey := ed25519.GenPrivKeyFromSecret(seed).PubKey()
				pk := secp256k1.GenPrivKeyFromSecret(seed)
				valAddr := sdk.ValAddress(pk.PubKey().Address().Bytes())
				return funcs.Must(stakingtypes.NewValidator(valAddr, consKey, stakingtypes.Description{}))
			},
		}

		distriK := distribution.NewKeeper(encCfg.Codec, sdk.NewKVStoreKey(distributiontypes.StoreKey), subspace, ak, bk, sk, authtypes.FeeCollectorName, map[string]bool{})
		k = keeper.NewKeeper(distriK, ak, bk, sk, authtypes.FeeCollectorName)
		k.SetFeePool(ctx, distributiontypes.FeePool{CommunityPool: sdk.DecCoins{}})
		k.SetParams(ctx, distributiontypes.DefaultParams())
	}).
		When("allocate tokens", func() {
			k.AllocateTokens(ctx, 0, 1, sdk.ConsAddress{}, nil)
		}).
		Then("allocate to community pool and burn the rest", func(t *testing.T) {
			assert.Len(t, bk.BurnCoinsCalls(), 1)

			feesBurnedType := proto.MessageName(&types.FeesBurned{})
			assert.Len(t, slices.Filter(ctx.EventManager().Events(), func(e sdk.Event) bool {
				return e.Type == feesBurnedType
			}), 1)

			burned, tax := expectedBurnAndTax(ctx, k, fees)
			expectedBurnedFees := sdk.NewCoins(slices.Map(burned, types.WithBurnedPrefix)...)

			assert.Equal(t, expectedBurnedFees, accBalances[types.ZeroAddress.String()])
			assert.Equal(t, k.GetFeePool(ctx).CommunityPool, tax)
		}).
		Run(t)
}

func expectedBurnAndTax(ctx sdk.Context, k keeper.Keeper, fee sdk.Coins) (sdk.Coins, sdk.DecCoins) {
	feesDec := sdk.NewDecCoinsFromCoins(fee...)
	tax := feesDec.MulDecTruncate(k.GetCommunityTax(ctx))
	burnAmt, remainder := feesDec.Sub(tax).TruncateDecimal()

	return burnAmt, tax.Add(remainder...)
}
