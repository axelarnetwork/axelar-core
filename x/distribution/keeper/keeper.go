package keeper

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distributionTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

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
func (k Keeper) AllocateTokens(ctx context.Context, _ int64, _ []abci.VoteInfo) error {
	// fetch and clear the collected fees for distribution, since this is
	// called in BeginBlock, collected fees will be from the previous block
	// (and distributed to the previous proposer)
	feeCollector := k.authKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	feesCollectedInt := k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())
	feesCollected := sdk.NewDecCoinsFromCoins(feesCollectedInt...)

	// transfer collected fees to the distribution module account
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, distributionTypes.ModuleName, feesCollectedInt)
	if err != nil {
		return err
	}

	feePool, err := k.FeePool.Get(ctx)
	if err != nil {
		return err
	}

	communityTaxRate := funcs.Must(k.GetCommunityTax(ctx))
	communityPoolAmount := feesCollected.MulDecTruncate(communityTaxRate)
	remaining := feesCollected.Sub(communityPoolAmount)

	// truncate the remaining coins, return remainder to community pool
	feesToBurn, truncationRemainder := remaining.TruncateDecimal()
	communityPoolAmount = communityPoolAmount.Add(truncationRemainder...)

	// allocate community funding
	feePool.CommunityPool = feePool.CommunityPool.Add(communityPoolAmount...)
	err = k.FeePool.Set(ctx, feePool)
	if err != nil {
		return err
	}

	// burn the rest
	err = k.bankKeeper.BurnCoins(ctx, distributionTypes.ModuleName, feesToBurn)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	events.Emit(sdkCtx, &types.FeesBurned{
		Coins: feesToBurn,
	})

	// track cumulative burned fees
	feesBurned := slices.Map(feesToBurn, types.WithBurnedPrefix)
	err = k.bankKeeper.MintCoins(ctx, distributionTypes.ModuleName, feesBurned)
	if err != nil {
		return err
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, distributionTypes.ModuleName, types.ZeroAddress, feesBurned)
}

// BeginBlocker mirrors the cosmos-sdk distribution keeper's BeginBlocker
// (cosmos-sdk/x/distribution/keeper/abci.go) so the custom AllocateTokens
// defined on this keeper is used instead of the SDK's. The SDK's external
// community pool handling is omitted because the app does not wire
// x/protocolpool.
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	start := telemetry.Now()
	defer telemetry.ModuleMeasureSince(distributionTypes.ModuleName, start, telemetry.MetricKeyBeginBlocker)

	// determine the total power signing the block
	var previousTotalPower int64
	for _, voteInfo := range ctx.VoteInfos() {
		previousTotalPower += voteInfo.Validator.Power
	}

	// TODO this is Tendermint-dependent
	// ref https://github.com/cosmos/cosmos-sdk/issues/3095
	if ctx.BlockHeight() > 1 {
		if err := k.AllocateTokens(ctx, previousTotalPower, ctx.VoteInfos()); err != nil {
			return err
		}
	}

	// record the proposer for when we pay out on the next block
	consAddr := sdk.ConsAddress(ctx.BlockHeader().ProposerAddress)
	return k.SetPreviousProposerConsAddr(ctx, consAddr)
}
