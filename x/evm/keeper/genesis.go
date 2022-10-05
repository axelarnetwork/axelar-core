package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

// InitGenesis initializes the state from a genesis file
func (k BaseKeeper) InitGenesis(ctx sdk.Context, state types.GenesisState) {
	for _, chain := range state.Chains {
		ck := k.ForChain(chain.Params.Chain).(chainKeeper)

		ck.SetParams(ctx, chain.Params)

		for _, burner := range chain.BurnerInfos {
			ck.SetBurnerInfo(ctx, burner)
		}

		if err := ck.validateCommandQueueState(chain.CommandQueue, commandQueueName); err != nil {
			panic(err)
		}
		ck.getCommandQueue(ctx).ImportState(chain.CommandQueue)

		for _, deposit := range chain.ConfirmedDeposits {
			ck.SetDeposit(ctx, deposit, types.DepositStatus_Confirmed)
		}

		for _, deposit := range chain.BurnedDeposits {
			ck.SetDeposit(ctx, deposit, types.DepositStatus_Burned)
		}

		var latestBatch types.CommandBatchMetadata
		for _, batch := range chain.CommandBatches {
			ck.setCommandBatchMetadata(ctx, batch)
			latestBatch = batch
		}

		if latestBatch.Status != types.BatchNonExistent {
			ck.setLatestBatchMetadata(ctx, latestBatch)
			ck.SetLatestSignedCommandBatchID(ctx, latestBatch.ID)
		}

		ck.setGateway(ctx, chain.Gateway)

		for _, token := range chain.Tokens {
			ck.setTokenMetadata(ctx, token)
		}

		for _, event := range chain.Events {
			ck.setEvent(ctx, event)
		}

		if err := ck.validateConfirmedEventQueueState(chain.ConfirmedEventQueue, confirmedEventQueueName); err != nil {
			panic(err)
		}
		ck.GetConfirmedEventQueue(ctx).(utils.GeneralKVQueue).ImportState(chain.ConfirmedEventQueue)
	}
}

// ExportGenesis generates a genesis file from the state
func (k BaseKeeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	return types.NewGenesisState(k.getChains(ctx))
}

func (k BaseKeeper) getChains(ctx sdk.Context) []types.GenesisState_Chain {
	iter := k.getBaseStore(ctx).Iterator(utils.KeyFromStr(subspacePrefix))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	var chains []types.GenesisState_Chain
	for ; iter.Valid(); iter.Next() {
		ck := k.ForChain(nexus.ChainName(iter.Value())).(chainKeeper)

		chain := types.GenesisState_Chain{
			Params:              ck.GetParams(ctx),
			BurnerInfos:         ck.getBurnerInfos(ctx),
			CommandQueue:        ck.getCommandQueue(ctx).ExportState(),
			ConfirmedDeposits:   ck.getConfirmedDeposits(ctx),
			BurnedDeposits:      ck.getBurnedDeposits(ctx),
			CommandBatches:      ck.getCommandBatchesMetadata(ctx),
			Gateway:             ck.getGateway(ctx),
			Tokens:              ck.getTokensMetadata(ctx),
			Events:              ck.getEvents(ctx),
			ConfirmedEventQueue: ck.GetConfirmedEventQueue(ctx).(utils.GeneralKVQueue).ExportState(),
		}
		chains = append(chains, chain)
	}

	return chains
}
