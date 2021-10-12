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
	externalChainVotingInflationRate := k.GetParams(ctx).ExternalChainVotingInflationRate
	externalChainVotingReward := sdk.NewCoin(
		m.GetParams(ctx).MintDenom,
		totalStakingSupply.ToDec().Mul(externalChainVotingInflationRate).QuoInt64(int64(blocksPerYear)).RoundInt(),
	)

	for _, chain := range n.GetChains(ctx) {
		externalChainVotingRewardPool := k.GetPool(ctx, chain.Name)

		for _, maintainer := range n.GetChainMaintainers(ctx, chain) {
			externalChainVotingRewardPool.AddReward(maintainer, externalChainVotingReward)
		}
	}

	return nil
}
