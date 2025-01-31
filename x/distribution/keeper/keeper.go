package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/distribution/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// Keeper wraps the distribution keeper to customize fee allocation mechanism
type Keeper struct {
	distribution.Keeper

	authKeeper       types.AccountKeeper
	bankKeeper       types.BankKeeper
	stakingKeeper    types.StakingKeeper
	feeCollectorName string
}

func NewKeeper(
	k distribution.Keeper, ak types.AccountKeeper, bk types.BankKeeper,
	sk types.StakingKeeper, feeCollectorName string,
) Keeper {
	return Keeper{
		Keeper:           k,
		authKeeper:       ak,
		bankKeeper:       bk,
		stakingKeeper:    sk,
		feeCollectorName: feeCollectorName,
	}
}

// AllocateTokens modifies the fee distribution by:
// - Allocating the community tax portion to the community pool
// - Burning all remaining tokens instead of distributing to validators
func (k Keeper) AllocateTokens(ctx sdk.Context, _, _ int64, _ sdk.ConsAddress, _ []abci.VoteInfo) {
	// fetch and clear the collected fees for distribution, since this is
	// called in BeginBlock, collected fees will be from the previous block
	// (and distributed to the previous proposer)
	feeCollector := k.authKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	feesCollectedInt := k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	// transfer collected fees to the distribution module account
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, distributionTypes.ModuleName, feesCollectedInt)
	if err != nil {
		panic(err)
	}

	feePool := k.GetFeePool(ctx)

	communityTaxRate := k.GetCommunityTax(ctx)
	communityPoolAmount := feesCollected.MulDecTruncate(communityTaxRate)
	remaining := feesCollected.Sub(communityPoolAmount)

	// truncate the remaining coins, return remainder to community pool
	feeToBurn, truncationRemainder := remaining.TruncateDecimal()
	communityPoolAmount = communityPoolAmount.Add(truncationRemainder...)

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(communityPoolAmount...)
	k.SetFeePool(ctx, feePool)

	// burn the rest
	funcs.MustNoErr(k.bankKeeper.BurnCoins(ctx, distributionTypes.ModuleName, feeToBurn))
	events.Emit(ctx, &types.FeeBurnedEvent{
		Coins: feeToBurn,
	})

	// track cumulative burned fee
	feeBurned := slices.Map(feeToBurn, types.WithBurnedPrefix)
	funcs.MustNoErr(k.bankKeeper.MintCoins(ctx, distributionTypes.ModuleName, feeBurned))
	funcs.MustNoErr(k.bankKeeper.SendCoinsFromModuleToAccount(ctx, distributionTypes.ModuleName, types.ZeroAddress, feeBurned))
}
