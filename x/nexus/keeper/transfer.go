package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func getTransferPrefix(chain string, state exported.TransferState) utils.Key {
	return transferPrefix.
		AppendStr(state.String()).
		AppendStr(chain)
}

func getTransferKey(transfer exported.CrossChainTransfer) utils.Key {
	return getTransferPrefix(transfer.Recipient.Chain.Name, transfer.State).
		Append(utils.KeyFromStr(strconv.FormatUint(transfer.ID, 10)))
}

func (k Keeper) getTransfers(ctx sdk.Context) (transfers []exported.CrossChainTransfer) {
	iter := k.getStore(ctx).Iterator(transferPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var transfer exported.CrossChainTransfer
		iter.UnmarshalValue(&transfer)

		transfers = append(transfers, transfer)
	}

	return transfers
}

func (k Keeper) setTransfer(ctx sdk.Context, transfer exported.CrossChainTransfer) {
	k.getStore(ctx).Set(getTransferKey(transfer), &transfer)
}

func (k Keeper) deleteTransfer(ctx sdk.Context, transfer exported.CrossChainTransfer) {
	k.getStore(ctx).Delete(getTransferKey(transfer))
}

func (k Keeper) setNewPendingTransfer(ctx sdk.Context, recipient exported.CrossChainAddress, amount sdk.Coin) {
	id := k.getNonce(ctx)
	k.setTransfer(ctx, exported.NewPendingCrossChainTransfer(id, recipient, amount))
	k.setNonce(ctx, id+1)
}

// EnqueueForTransfer appoints the amount of tokens to be transfered/minted to the recipient previously linked to the specified sender
func (k Keeper) EnqueueForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, asset sdk.Coin, feeRate sdk.Dec) error {
	if !sender.Chain.SupportsForeignAssets && sender.Chain.NativeAsset != asset.Denom {
		return fmt.Errorf("sender's chain %s does not support foreign assets", sender.Chain.Name)
	}

	if sender.Chain.NativeAsset != asset.Denom && k.getChainTotal(ctx, sender.Chain, asset.Denom).IsLT(asset) {
		return fmt.Errorf("not enough funds available for asset '%s' in chain %s", asset.Denom, sender.Chain.Name)
	}

	recipient, ok := k.GetRecipient(ctx, sender)
	if !ok {
		return fmt.Errorf("no recipient linked to sender %s", sender.String())
	}

	if !recipient.Chain.SupportsForeignAssets && recipient.Chain.NativeAsset != asset.Denom {
		return fmt.Errorf("recipient's chain %s does not support foreign assets", recipient.Chain.Name)
	}

	// collect fee
	// TODO: this should be now done upon mint/withdrawl rather than per individual transfer
	feeCollector, ok := k.axelarnetKeeper.GetFeeCollector(ctx)
	feeDue := sdk.NewDecFromInt(asset.Amount).Mul(feeRate).TruncateInt()
	if ok && feeDue.IsPositive() {
		asset.Amount = asset.Amount.Sub(feeDue)
		fee := sdk.NewCoin(asset.Denom, feeDue)
		feeRecipient := exported.CrossChainAddress{Chain: axelarnet.Axelarnet, Address: feeCollector.String()}
		k.setNewPendingTransfer(ctx, feeRecipient, fee)
	}

	if sender.Chain.NativeAsset != asset.Denom {
		k.subtractFromChainTotal(ctx, sender.Chain, asset)
	}

	// merging transfers for the specified recipient
	previousTransfer, found := k.getPendingTransferForRecipientAndAsset(ctx, recipient, asset.Denom)
	if found {
		asset = asset.Add(previousTransfer.Asset)
		k.deleteTransfer(ctx, previousTransfer)
	}

	k.setNewPendingTransfer(ctx, recipient, asset)
	k.Logger(ctx).Info(fmt.Sprintf("Transfer of %s to cross chain address %s in %s successfully prepared",
		asset.String(), recipient.Address, recipient.Chain.Name))

	return nil
}

func (k Keeper) getPendingTransferForRecipientAndAsset(ctx sdk.Context, recipient exported.CrossChainAddress, denom string) (exported.CrossChainTransfer, bool) {
	iter := k.getStore(ctx).Iterator(getTransferPrefix(recipient.Chain.Name, exported.Pending))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var transfer exported.CrossChainTransfer
		iter.UnmarshalValue(&transfer)

		if recipient == transfer.Recipient && denom == transfer.Asset.Denom {
			return transfer, true
		}
	}

	return exported.CrossChainTransfer{}, false
}

// ArchivePendingTransfer marks the transfer for the given recipient as concluded and archived
func (k Keeper) ArchivePendingTransfer(ctx sdk.Context, transfer exported.CrossChainTransfer) {
	k.deleteTransfer(ctx, transfer)

	transfer.State = exported.Archived
	k.setTransfer(ctx, transfer)

	// Update the total nexus for the chain if it is a foreign asset
	info, _ := k.GetChain(ctx, transfer.Recipient.Chain.Name)
	if info.NativeAsset != transfer.Asset.Denom {
		k.AddToChainTotal(ctx, transfer.Recipient.Chain, transfer.Asset)
	}
}

// GetTransfersForChain returns the current set of transfers with the given state for the given chain
func (k Keeper) GetTransfersForChain(ctx sdk.Context, chain exported.Chain, state exported.TransferState) (transfers []exported.CrossChainTransfer) {
	iter := k.getStore(ctx).Iterator(getTransferPrefix(chain.Name, state))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var transfer exported.CrossChainTransfer
		iter.UnmarshalValue(&transfer)

		transfers = append(transfers, transfer)
	}

	return transfers
}
