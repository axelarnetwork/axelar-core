package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoprototypes "github.com/gogo/protobuf/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// InitGenesis initializes the axelarnet module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	if len(genState.CollectorAddress) > 0 {
		if err := k.SetFeeCollector(ctx, genState.CollectorAddress); err != nil {
			panic(err)
		}
	}

	slices.ForEach(genState.Chains, func(c types.CosmosChain) {
		if _, ok := k.GetCosmosChainByName(ctx, c.Name); ok {
			panic(fmt.Sprintf("cosmos chain %s already registered", c.Name))
		}

		if _, ok := k.GetChainNameByIBCPath(ctx, c.IBCPath); ok {
			panic(fmt.Sprintf("ibc path %s already registered", c.IBCPath))
		}

		funcs.MustNoErr(k.SetCosmosChain(ctx, c))

		// axelarnet does not have an ibc path
		if !c.Name.Equals(exported.Axelarnet.Name) {
			funcs.MustNoErr(k.SetChainByIBCPath(ctx, c.IBCPath, c.Name))
		}
	})

	funcs.MustNoErr(k.validateIBCTransferQueueState(genState.TransferQueue, ibcTransferQueueName))

	k.GetIBCTransferQueue(ctx).(utils.GeneralKVQueue).ImportState(genState.TransferQueue)

	slices.ForEach(genState.IBCTransfers, func(t types.IBCTransfer) { funcs.MustNoErr(k.setTransfer(ctx, t)) })

	sortedKeys := types.SortedMapKeys(genState.SeqIDMapping, strings.Compare)
	slices.ForEach(sortedKeys, func(seqKey string) {
		funcs.MustNoErr(k.getStore(ctx).SetNewValidated(
			key.FromBz([]byte(seqKey)),
			utils.NoValidation(&gogoprototypes.UInt64Value{Value: genState.SeqIDMapping[seqKey]}),
		))
	})
}

// ExportGenesis returns the reward module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	collector, _ := k.GetFeeCollector(ctx)

	return types.NewGenesisState(
		k.GetParams(ctx),
		collector,
		k.getCosmosChains(ctx),
		k.GetIBCTransferQueue(ctx).(utils.GeneralKVQueue).ExportState(),
		k.getIBCTransfers(ctx),
		k.getSeqIDMappings(ctx),
	)
}
