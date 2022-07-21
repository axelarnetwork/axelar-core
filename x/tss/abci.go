package tss

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/utils/slices"
)

// BeginBlocker check for infraction evidence or downtime of validators
// on every begin block
func BeginBlocker(_ sdk.Context, _ abci.RequestBeginBlock, _ keeper.Keeper) {}

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, req abci.RequestEndBlock, keeper keeper.Keeper, m types.MultiSigKeeper, n types.Nexus) []abci.ValidatorUpdate {
	if ctx.BlockHeight() > 0 && (ctx.BlockHeight()%keeper.GetHeartbeatPeriodInBlocks(ctx)) == 0 {
		emitHeartbeatEvent(ctx, m, n)
	}

	return nil
}

func emitHeartbeatEvent(ctx sdk.Context, m types.MultiSigKeeper, n types.Nexus) {
	var keyInfos []types.KeyInfo

	for _, chain := range n.GetChains(ctx) {
		keyIDs := m.GetActiveKeyIDs(ctx, chain.Name)

		keyInfos = append(
			keyInfos,
			slices.Map(keyIDs, func(keyID multisig.KeyID) types.KeyInfo {
				return types.KeyInfo{KeyID: exported.KeyID(keyID), KeyType: exported.Multisig}
			})...,
		)
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeHeartBeat,
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueSend),
		sdk.NewAttribute(types.AttributeKeyKeyInfos, string(types.ModuleCdc.LegacyAmino.MustMarshalJSON(keyInfos))),
	))
}
