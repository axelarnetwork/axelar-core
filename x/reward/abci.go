package reward

import (
	"github.com/axelarnetwork/axelar-core/x/reward/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker is called at the beginning of every block
func BeginBlocker(ctx sdk.Context, _ abci.RequestBeginBlock, _ types.Rewarder) {
}

// EndBlocker is called at the end of every block, process external chain voting inflation
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, k types.Rewarder, n types.Nexus, m types.Minter) []abci.ValidatorUpdate {
	totalStakingSupply := m.StakingTokenSupply(ctx)
	blocksPerYear := m.GetParams(ctx).BlocksPerYear
	inflationRate := k.GetParams(ctx).ExternalChainVotingInflationRate
	denom := m.GetParams(ctx).MintDenom
	amountPerChain := totalStakingSupply.ToDec().Mul(inflationRate).QuoInt64(int64(blocksPerYear))

	for _, chain := range n.GetChains(ctx) {
		externalChainVotingRewardPool := k.GetPool(ctx, chain.Name)
		maintainers := n.GetChainMaintainers(ctx, chain)
		if len(maintainers) == 0 {
			continue
		}

		reward := sdk.NewCoin(
			denom,
			amountPerChain.QuoInt64(int64(len(maintainers))).RoundInt(),
		)
		for _, maintainer := range maintainers {
			externalChainVotingRewardPool.AddReward(maintainer, reward)
		}
	}

	return nil
}
