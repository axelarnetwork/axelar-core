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

	slices.ForEach(state.KeygenSessions, func(keygenSession types.KeygenSession) { k.setKeygenSession(ctx, keygenSession) })
	slices.ForEach(state.Keys, func(key types.Key) { k.setKey(ctx, key) })
	slices.ForEach(state.SigningSessions, func(signingSession types.SigningSession) { k.setSigningSession(ctx, signingSession) })
	slices.ForEach(state.KeyEpochs, func(keyEpoch types.KeyEpoch) { k.setKeyEpoch(ctx, keyEpoch) })

	keyEpochsByChain := slices.GroupBy(state.KeyEpochs, func(keyEpoch types.KeyEpoch) nexus.ChainName { return keyEpoch.GetChain() })
	for chain, keyEpochs := range keyEpochsByChain {
		sort.SliceStable(keyEpochs, func(i, j int) bool { return keyEpochs[i].Epoch < keyEpochs[j].Epoch })

		key := funcs.MustOk(k.getKey(ctx, keyEpochs[len(keyEpochs)-1].GetKeyID()))
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
