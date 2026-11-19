package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

// InitGenesis initializes the permission module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.setParams(ctx, genState.Params)

	if genState.GovernanceKey != nil {
		k.setGovernanceKey(ctx, *genState.GovernanceKey)
	}

	for _, account := range genState.GovAccounts {
		k.setGovAccount(ctx, account)
	}

	k.setPauseLevel(ctx, genState.PauseLevel)
}

// ExportGenesis returns the permission module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	var genState *types.GenesisState

	governanceKey, ok := k.GetGovernanceKey(ctx)
	if !ok {
		genState = types.NewGenesisState(
			k.GetParams(ctx),
			nil,
			k.getGovAccounts(ctx),
		)
	} else {
		genState = types.NewGenesisState(
			k.GetParams(ctx),
			&governanceKey,
			k.getGovAccounts(ctx),
		)
	}

	genState.PauseLevel = k.GetPauseLevel(ctx)

	return genState
}
