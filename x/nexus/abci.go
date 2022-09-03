package nexus

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ types.Nexus) {}

// EndBlocker called every block, checking the chain maintainers of all activated chains
// - if a chain maintainer has missed voting for too many polls, then it will be de-registered
// - if a chain maintainer has voted incorrectly for too many polls, then it will be de-registered
// - if a chain maintainer does not active proxy set, then it will be de-registered
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, n types.Nexus, r types.RewardKeeper, s types.Snapshotter) ([]abci.ValidatorUpdate, error) {
	if err := checkChainMaintainers(ctx, n, r, s); err != nil {
		return nil, err
	}

	return nil, nil
}

func checkChainMaintainers(ctx sdk.Context, n types.Nexus, r types.RewardKeeper, s types.Snapshotter) error {
	for _, chain := range n.GetChains(ctx) {
		if !n.IsChainActivated(ctx, chain) {
			continue
		}

		rewardPool := r.GetPool(ctx, chain.Name.String())
		params := n.GetParams(ctx)
		window := int(params.ChainMaintainerCheckWindow)

		for _, maintainerState := range n.GetChainMaintainerStates(ctx, chain) {
			missingVoteCount := maintainerState.CountMissingVotes(window)
			incorrectVoteCount := maintainerState.CountIncorrectVotes(window)
			_, hasProxyActive := s.GetProxy(ctx, maintainerState.GetAddress())

			if hasProxyActive &&
				utils.NewThreshold(int64(missingVoteCount), int64(window)).LTE(params.ChainMaintainerMissingVoteThreshold) &&
				utils.NewThreshold(int64(incorrectVoteCount), int64(window)).LTE(params.ChainMaintainerIncorrectVoteThreshold) {
				continue
			}

			rewardPool.ClearRewards(maintainerState.GetAddress())
			if err := n.RemoveChainMaintainer(ctx, chain, maintainerState.GetAddress()); err != nil {
				return err
			}

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeChainMaintainer,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDeregister),
					sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
					sdk.NewAttribute(types.AttributeKeyChainMaintainerAddress, maintainerState.GetAddress().String()),
				),
			)

			n.Logger(ctx).Info(fmt.Sprintf("deregistered validator %s as maintainer for chain %s", maintainerState.GetAddress().String(), chain.Name))
		}
	}

	return nil
}
