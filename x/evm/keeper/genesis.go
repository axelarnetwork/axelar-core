package keeper

import (
	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k baseKeeper) InitGenesis(ctx sdk.Context, state types.GenesisState) {
	for _, chain := range state.Chains {
		ck := k.ForChain(chain.Params.Chain).(chainKeeper)

		ck.SetParams(ctx, chain.Params)

		for _, burner := range chain.BurnerInfos {
			ck.SetBurnerInfo(ctx, burner)
		}

		ck.setCommandQueue(ctx, chain.CommandQueue)

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
			ck.setLatestSignedCommandBatchID(ctx, latestBatch.ID)
		}

		ck.setGateway(ctx, chain.Gateway)

		for _, token := range chain.Tokens {
			ck.setTokenMetadata(ctx, token)
		}
	}
}

func (k baseKeeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	return types.NewGenesisState(k.getChains(ctx))
}

func (k baseKeeper) getChains(ctx sdk.Context) []types.GenesisState_Chain {
	iter := k.getBaseStore(ctx).Iterator(subspacePrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	var chains []types.GenesisState_Chain
	for ; iter.Valid(); iter.Next() {
		ck := k.ForChain(string(iter.Value())).(chainKeeper)

		chain := types.GenesisState_Chain{
			Params:            ck.GetParams(ctx),
			BurnerInfos:       ck.getBurnerInfos(ctx),
			CommandQueue:      ck.serializeCommandQueue(ctx),
			ConfirmedDeposits: ck.GetConfirmedDeposits(ctx),
			BurnedDeposits:    ck.getBurnedDeposits(ctx),
			CommandBatches:    ck.getCommandBatchesMetadata(ctx),
			Gateway:           ck.getGateway(ctx),
			Tokens:            ck.getTokensMetadata(ctx),
		}
		chains = append(chains, chain)
	}

	return chains
}
