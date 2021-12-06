package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/permission/types"
)

// InitGenesis initializes the reward module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.setParams(ctx, genState.Params)

	if genState.GovernanceKey != nil {
		k.setGovernanceKey(ctx, *genState.GovernanceKey)
	}

	for _, account := range genState.GovAccounts {
		k.setGovAccount(ctx, account)
	}
}

// ExportGenesis returns the reward module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	governanceKey, ok := k.GetGovernanceKey(ctx)
	if !ok {
		return types.NewGenesisState(
			k.GetParams(ctx),
			nil,
			k.getGovAccounts(ctx),
		)
	}

	return types.NewGenesisState(
		k.GetParams(ctx),
		&governanceKey,
		k.getGovAccounts(ctx),
	)
}
