package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// InitGenesis initializes the state from a genesis file
func (k Keeper) InitGenesis(ctx sdk.Context, state *types.GenesisState) {
	k.setParams(ctx, state.Params)

	slices.ForEach(state.KeygenSessions, withContext(ctx, k.setKeygenSession))
	slices.ForEach(state.Keys, withContext(ctx, k.setKey))
	slices.ForEach(state.SigningSessions, withContext(ctx, k.setSigningSession))
	slices.ForEach(state.KeyEpochs, withContext(ctx, k.setKeyEpoch))

	keyEpochsByChain := slices.GroupBy(state.KeyEpochs, func(keyEpoch types.KeyEpoch) nexus.ChainName { return keyEpoch.GetChain() })
	for chain, keyEpochs := range keyEpochsByChain {
		sort.SliceStable(keyEpochs, func(i, j int) bool { return keyEpochs[i].Epoch < keyEpochs[j].Epoch })
		latestKeyEpoch := slices.Last(keyEpochs)

		key := funcs.MustOk(k.getKey(ctx, latestKeyEpoch.GetKeyID()))
		switch key.State {
		case exported.Assigned:
			k.setKeyRotationCount(ctx, chain, uint64(len(keyEpochs)-1))
		case exported.Active:
			k.setKeyRotationCount(ctx, chain, uint64(len(keyEpochs)))
		default:
			panic(fmt.Errorf("invalid state for key %s", key.GetID()))
		}
	}

	k.setSigningSessionCount(ctx, uint64(len(state.SigningSessions)))
}

// ExportGenesis generates a genesis file from the state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(
		k.getParams(ctx),
		k.getKeygenSessions(ctx),
		k.getSigningSessions(ctx),
		k.getKeys(ctx),
		k.getKeyEpochs(ctx),
	)
}

func withContext[T any](ctx sdk.Context, fn func(sdk.Context, T)) func(T) {
	return func(t T) {
		fn(ctx, t)
	}
}
