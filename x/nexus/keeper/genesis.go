package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
)

// InitGenesis initializes the reward module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	k.setNonce(ctx, genState.Nonce)

	for _, chain := range genState.Chains {
		if _, ok := k.GetChain(ctx, chain.Name); ok {
			panic(fmt.Errorf("chain %s already set", chain.Name))
		}

		k.SetChain(ctx, chain)
	}

	for _, chainState := range genState.ChainStates {
		if _, ok := k.getChainState(ctx, chainState.Chain); ok {
			panic(fmt.Errorf("chain state %s already set", chainState.Chain.Name))
		}
		for _, asset := range chainState.Assets {
			if asset.IsNativeAsset {
				k.setChainByNativeAsset(ctx, asset.Denom, chainState.Chain)
			}
		}
		k.setChainState(ctx, chainState)
	}

	for _, linkedAddresses := range genState.LinkedAddresses {
		if _, ok := k.getLinkedAddresses(ctx, linkedAddresses.DepositAddress); ok {
			panic(fmt.Errorf("linked addresses for deposit address %s on chain %s already set", linkedAddresses.DepositAddress.Address, linkedAddresses.DepositAddress.Chain.Name))
		}

		k.setLinkedAddresses(ctx, linkedAddresses)
	}

	transferSeen := make(map[exported.TransferID]bool)
	for _, transfer := range genState.Transfers {
		if transferSeen[transfer.ID] {
			panic(fmt.Errorf("transfer %d already set", transfer.ID))
		}

		k.setTransfer(ctx, transfer)
		transferSeen[transfer.ID] = true
	}

	k.setTransferFee(ctx, genState.Fee)

	for _, feeInfo := range genState.FeeInfos {
		chain, ok := k.GetChain(ctx, feeInfo.Chain)
		if !ok {
			panic(fmt.Errorf("chain %s not found", feeInfo.Chain))
		}

		if _, found := k.getFeeInfo(ctx, chain, feeInfo.Asset); found {
			panic(fmt.Errorf("fee info for chain %s and asset %s already registered", chain.Name, feeInfo.Asset))
		}

		if err := k.RegisterFee(ctx, chain, feeInfo); err != nil {
			panic(err)
		}
	}

	for _, rateLimit := range genState.RateLimits {
		if _, found := k.getRateLimit(ctx, rateLimit.Chain, rateLimit.Limit.Denom); found {
			panic(fmt.Errorf("rate limit for chain %s and asset %s already registered", rateLimit.Chain, rateLimit.Limit.Denom))
		}

		funcs.MustNoErr(k.SetRateLimit(ctx, rateLimit.Chain, rateLimit.Limit, rateLimit.Window))
	}

	for _, transferEpoch := range genState.TransferEpochs {
		if _, ok := k.GetChain(ctx, transferEpoch.Chain); !ok {
			panic(fmt.Errorf("chain %s not found", transferEpoch.Chain))
		}

		if _, found := k.getTransferEpoch(ctx, transferEpoch.Chain, transferEpoch.Amount.Denom, transferEpoch.Direction); found {
			panic(fmt.Errorf("transfer rate for chain %s (%s) and asset %s already registered", transferEpoch.Chain, transferEpoch.Direction, transferEpoch.Amount.Denom))
		}

		k.setTransferEpoch(ctx, transferEpoch)
	}

	for _, msg := range genState.Messages {
		funcs.MustNoErr(k.setMessage(ctx, msg))
	}

	utils.NewCounter[uint64](messageNonceKey, k.getStore(ctx)).Set(ctx, genState.MessageNonce)
}

// ExportGenesis returns the reward module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	return types.NewGenesisState(
		k.GetParams(ctx),
		k.getNonce(ctx),
		k.GetChains(ctx),
		k.getChainStates(ctx),
		k.getAllLinkedAddresses(ctx),
		k.getTransfers(ctx),
		k.getTransferFee(ctx),
		k.getFeeInfos(ctx),
		k.getRateLimits(ctx),
		k.getTransferEpochs(ctx),
		k.getMessages(ctx),
		utils.NewCounter[uint64](messageNonceKey, k.getStore(ctx)).Curr(ctx),
	)
}
