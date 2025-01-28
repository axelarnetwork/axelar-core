package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/distribution/types"
	"github.com/axelarnetwork/utils/funcs"
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
func (k Keeper) AllocateTokens(
	ctx sdk.Context, _, totalPreviousPower int64,
	previousProposer sdk.ConsAddress, _ []abci.VoteInfo,
) {
	logger := k.Logger(ctx)

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

	// temporary workaround to keep CanWithdrawInvariant happy
	// general discussions here: https://github.com/cosmos/cosmos-sdk/issues/2906#issuecomment-441867634
	feePool := k.GetFeePool(ctx)
	if totalPreviousPower == 0 {
		feePool.CommunityPool = feePool.CommunityPool.Add(feesCollected...)
		k.SetFeePool(ctx, feePool)
		return
	}

	communityTaxRate := k.GetCommunityTax(ctx)
	communityPoolAmount := feesCollected.MulDecTruncate(communityTaxRate)
	remaining := feesCollected.Sub(communityPoolAmount)

	// truncate the remaining coins, return remainder to community pool
	feeToBurn, remainder := remaining.TruncateDecimal()
	communityPoolAmount = communityPoolAmount.Add(remainder...)

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(communityPoolAmount...)
	k.SetFeePool(ctx, feePool)

	// burn the rest
	funcs.MustNoErr(k.bankKeeper.BurnCoins(ctx, distributionTypes.ModuleName, feeToBurn))
	events.Emit(ctx, &types.FeeBurnedEvent{
		Coins: feeToBurn,
	})

	// keep the error log from the original implementation
	proposerValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, previousProposer)
	if proposerValidator == nil {
		// previous proposer can be unknown if say, the unbonding period is 1 block, so
		// e.g. a validator undelegates at block X, it's removed entirely by
		// block X+1's endblock, then X+2 we need to refer to the previous
		// proposer for X+1, but we've forgotten about them.
		logger.Error(fmt.Sprintf(
			"WARNING: Attempt to allocate proposer rewards to unknown proposer %s. "+
				"This should happen only if the proposer unbonded completely within a single block, "+
				"which generally should not happen except in exceptional circumstances (or fuzz testing). "+
				"We recommend you investigate immediately.",
			previousProposer.String()))
	}
}
