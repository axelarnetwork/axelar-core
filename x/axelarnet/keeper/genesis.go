package keeper

import (
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"
	"golang.org/x/exp/maps"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/utils/slices"
)

// InitGenesis initializes the axelarnet module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.setParams(ctx, genState.Params)
	if len(genState.CollectorAddress) > 0 {
		if err := k.SetFeeCollector(ctx, genState.CollectorAddress); err != nil {
			panic(err)
		}
	}

	for _, chain := range genState.Chains {
		k.SetCosmosChain(ctx, chain)
	}

	if err := k.validateIBCTransferQueueState(genState.TransferQueue, routeTransferQueueName); err != nil {
		panic(err)
	}
	k.GetIBCTransferQueue(ctx).(utils.GeneralKVQueue).ImportState(genState.TransferQueue)

	for _, t := range genState.IBCTransfers {
		k.setTransfer(ctx, t)
	}

	seqKeys := maps.Keys(genState.SeqIDMapping)
	sort.SliceStable(seqKeys, func(i, j int) bool { return strings.Compare(seqKeys[i], seqKeys[j]) < 0 })
	slices.ForEach(seqKeys, func(seqKey string) {
		k.getStore(ctx).SetNew(key.FromBz([]byte(seqKey)), &gogoprototypes.UInt64Value{Value: genState.SeqIDMapping[seqKey]})
	})
}

// ExportGenesis returns the reward module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	collector, _ := k.GetFeeCollector(ctx)

	return types.NewGenesisState(
		k.getParams(ctx),
		collector,
		k.getCosmosChains(ctx),
		k.GetIBCTransferQueue(ctx).(utils.GeneralKVQueue).ExportState(),
		k.getIBCTransfers(ctx),
		k.getSeqIDMappings(ctx),
	)
}
