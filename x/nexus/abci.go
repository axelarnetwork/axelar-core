package nexus

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// EndBlocker called every block
func EndBlocker(ctx sdk.Context, n types.Nexus, r types.RewardKeeper, s types.Snapshotter) ([]abci.ValidatorUpdate, error) {
	checkChainMaintainers(ctx, n, r, s)

	routeQueuedMessages(ctx, n)

	return nil, nil
}

// checkChainMaintainers checks the chain maintainers of all activated chains
// - if a chain maintainer has missed voting for too many polls, then it will be de-registered
// - if a chain maintainer has voted incorrectly for too many polls, then it will be de-registered
// - if a chain maintainer does not active proxy set, then it will be de-registered
func checkChainMaintainers(ctx sdk.Context, n types.Nexus, r types.RewardKeeper, s types.Snapshotter) {
	for _, chain := range n.GetChains(ctx) {
		if !n.IsChainActivated(ctx, chain) {
			continue
		}

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

			deregistered := utils.RunCached(ctx, n, func(cachedCtx sdk.Context) (bool, error) {
				r.GetPool(cachedCtx, chain.Name.String()).ClearRewards(maintainerState.GetAddress())

				if err := n.RemoveChainMaintainer(cachedCtx, chain, maintainerState.GetAddress()); err != nil {
					return false, err
				}

				cachedCtx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeChainMaintainer,
						sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
						sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueDeregister),
						sdk.NewAttribute(types.AttributeKeyChain, chain.Name.String()),
						sdk.NewAttribute(types.AttributeKeyChainMaintainerAddress, maintainerState.GetAddress().String()),
					),
				)

				return true, nil
			})

			if !deregistered {
				n.Logger(ctx).Error(fmt.Sprintf("failed to deregister validator %s as maintainer for chain %s", maintainerState.GetAddress().String(), chain.Name))
				continue
			}

			n.Logger(ctx).Info(fmt.Sprintf("deregistered validator %s as maintainer for chain %s", maintainerState.GetAddress().String(), chain.Name))
		}
	}
}

func routeQueuedMessages(ctx sdk.Context, n types.Nexus) {
	params := n.GetParams(ctx)

	for i := uint64(0); i < params.EndBlockerLimit; i++ {
		msg, ok := n.DequeueRouteMessage(ctx)
		if !ok {
			break
		}

		// try routing the message
		_ = utils.RunCached(ctx, n, func(ctx sdk.Context) (struct{}, error) {
			return struct{}{}, n.RouteMessage(ctx, msg.ID)
		})
	}
}
