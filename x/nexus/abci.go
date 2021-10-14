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

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, k types.Nexus, staking types.StakingKeeper) []abci.ValidatorUpdate {
	for _, chain := range k.GetChains(ctx) {
		if k.IsChainActivated(ctx, chain) {
			continue
		}

		sumConsensusPower := sdk.ZeroInt()
		maintainers := k.GetChainMaintainers(ctx, chain)

		for _, maintainer := range maintainers {
			validator := staking.Validator(ctx, maintainer)
			if validator == nil {
				continue
			}

			if !validator.IsBonded() || validator.IsJailed() {
				continue
			}

			sumConsensusPower = sumConsensusPower.AddRaw(validator.GetConsensusPower(staking.PowerReduction(ctx)))
		}

		if utils.NewThreshold(sumConsensusPower.Int64(), staking.GetLastTotalPower(ctx).Int64()).GTE(k.GetParams(ctx).ChainActivationThreshold) {
			k.ActivateChain(ctx, chain)

			k.Logger(ctx).Info(fmt.Sprintf("chain %s activated", chain.Name))
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeChain,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueActivated),
					sdk.NewAttribute(types.AttributeKeyChain, chain.Name),
				),
			)
		}
	}

	return nil
}
