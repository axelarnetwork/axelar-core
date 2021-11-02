package reward

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/reward/types"
	snapshot "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	tsstypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// BeginBlocker is called at the beginning of every block
func BeginBlocker(ctx sdk.Context, _ abci.RequestBeginBlock, _ types.Rewarder) {}

// EndBlocker is called at the end of every block, process external chain voting inflation
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, k types.Rewarder, n types.Nexus, m types.Minter, s types.Staker, t types.Tss, ss types.Snapshotter) []abci.ValidatorUpdate {
	handleExternalChainVotingInflation(ctx, k, n, m)
	handleTssInflation(ctx, k, m, s, t, ss)

	return nil
}

func handleTssInflation(ctx sdk.Context, k types.Rewarder, m types.Minter, s types.Staker, t types.Tss, ss types.Snapshotter) {
	rewardPool := k.GetPool(ctx, tsstypes.ModuleName)
	minter := m.GetMinter(ctx)
	mintParams := m.GetParams(ctx)
	totalAmount := minter.BlockProvision(mintParams).Amount.ToDec().Mul(k.GetParams(ctx).TssRelativeInflationRate)

	var validators []stakingtypes.Validator
	totalConsensusPower := sdk.ZeroInt()

	validatorIterFn := func(_ int64, v stakingtypes.ValidatorI) bool {
		if !t.IsOperatorAvailable(ctx, v.GetOperator()) {
			return false
		}

		validator := v.(stakingtypes.Validator)
		illegibility, err := ss.GetValidatorIllegibility(ctx, &validator)
		if err != nil {
			panic(err)
		}

		if !illegibility.Is(snapshot.None) {
			return false
		}

		totalConsensusPower = totalConsensusPower.AddRaw(validator.GetConsensusPower(s.PowerReduction(ctx)))
		validators = append(validators, validator)

		return false
	}
	s.IterateBondedValidatorsByPower(ctx, validatorIterFn)

	for _, validator := range validators {
		// Each validator receives reward weighted by consensus power
		amount := totalAmount.MulInt64(validator.GetConsensusPower(s.PowerReduction(ctx))).QuoInt(totalConsensusPower).RoundInt()
		rewardPool.AddReward(
			validator.GetOperator(),
			sdk.NewCoin(mintParams.MintDenom, amount),
		)
	}

}

func handleExternalChainVotingInflation(ctx sdk.Context, k types.Rewarder, n types.Nexus, m types.Minter) {
	totalStakingSupply := m.StakingTokenSupply(ctx)
	blocksPerYear := m.GetParams(ctx).BlocksPerYear
	inflationRate := k.GetParams(ctx).ExternalChainVotingInflationRate
	denom := m.GetParams(ctx).MintDenom
	amountPerChain := totalStakingSupply.ToDec().Mul(inflationRate).QuoInt64(int64(blocksPerYear))

	for _, chain := range n.GetChains(ctx) {
		rewardPool := k.GetPool(ctx, chain.Name)
		maintainers := n.GetChainMaintainers(ctx, chain)
		if len(maintainers) == 0 {
			continue
		}

		reward := sdk.NewCoin(
			denom,
			amountPerChain.QuoInt64(int64(len(maintainers))).RoundInt(),
		)
		for _, maintainer := range maintainers {
			// Each maintainer receives equal amount of reward
			rewardPool.AddReward(maintainer, reward)
		}
	}
}
