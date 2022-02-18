package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
)

func getTransferPrefix(chain string, state exported.TransferState) utils.Key {
	return transferPrefix.
		AppendStr(state.String()).
		AppendStr(chain)
}

func getTransferKey(transfer exported.CrossChainTransfer) utils.Key {
	return getTransferPrefix(transfer.Recipient.Chain.Name, transfer.State).
		Append(utils.KeyFromStr(transfer.ID.String()))
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

func (k Keeper) setNewIncompleteTransfer(ctx sdk.Context, recipient exported.CrossChainAddress, amount sdk.Coin) exported.TransferID {
	id := k.getNonce(ctx)
	k.setTransfer(ctx, exported.NewIncompleteCrossChainTransfer(id, recipient, amount))
	k.setNonce(ctx, id+1)
	return exported.TransferID(id)
}

func (k Keeper) setNewPendingTransfer(ctx sdk.Context, recipient exported.CrossChainAddress, amount sdk.Coin) exported.TransferID {
	id := k.getNonce(ctx)
	k.setTransfer(ctx, exported.NewPendingCrossChainTransfer(id, recipient, amount))
	k.setNonce(ctx, id+1)
	return exported.TransferID(id)
}

func (k Keeper) setTransferFee(ctx sdk.Context, fee exported.TransferFee) {
	k.getStore(ctx).Set(transferFee, &fee)
}

func (k Keeper) getTransferFee(ctx sdk.Context) (fee exported.TransferFee) {
	k.getStore(ctx).Get(transferFee, &fee)
	return fee
}

func (k Keeper) computeChainFee(ctx sdk.Context, chain exported.Chain, asset sdk.Coin) sdk.Coin {
	feeInfo, ok := k.getFeeInfo(ctx, chain, asset.Denom)
	if !ok {
		feeInfo = exported.NewFeeInfo(sdk.ZeroDec(), sdk.ZeroUint(), sdk.ZeroUint())
	}

	amount := asset.Amount
	fee := sdk.Int(feeInfo.MinFee)

	remaining := sdk.ZeroInt()
	if fee.LT(amount) {
		remaining = amount.Sub(fee)
	}

	fee = fee.Add(sdk.NewDecFromInt(remaining).Mul(feeInfo.FeeRate).TruncateInt())

	if feeInfo.MaxFee.LT(sdk.Uint(fee)) {
		return sdk.NewCoin(asset.Denom, sdk.Int(feeInfo.MaxFee))
	}

	return sdk.NewCoin(asset.Denom, fee)
}

// computeTransferFee computes the fee for a cross-chain transfer
// If fee_info is not set for an asset on a chain, default of zero is used
// chain_fee = min(chain.max_fee, chain.min_fee + chain.fee_rate * max(0, amount - chain.min_fee))
// transfer_fee = deposit_chain_fee + recipient_chain_fee
func (k Keeper) computeTransferFee(ctx sdk.Context, depositChain exported.Chain, recipientChain exported.Chain, asset sdk.Coin) sdk.Coin {
	depositChainFee := k.computeChainFee(ctx, depositChain, asset)
	recipientChainFee := k.computeChainFee(ctx, recipientChain, asset)

	fees := depositChainFee.Add(recipientChainFee)

	return fees
}

// EnqueueForTransfer appoints the amount of tokens to be transferred/minted to the recipient previously linked to the specified sender
func (k Keeper) EnqueueForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, asset sdk.Coin, feeRate sdk.Dec) (exported.TransferID, error) {
	chain, isNativeAsset := k.GetChainByNativeAsset(ctx, asset.Denom)
	if !sender.Chain.SupportsForeignAssets && !(isNativeAsset && sender.Chain.Name == chain.Name) {
		return 0, fmt.Errorf("sender's chain %s does not support foreign assets", sender.Chain.Name)
	}

	if !k.IsChainActivated(ctx, sender.Chain) {
		return 0, fmt.Errorf("source chain '%s' is not activated", sender.Chain.Name)
	}

	recipient, ok := k.GetRecipient(ctx, sender)
	if !ok {
		return 0, fmt.Errorf("no recipient linked to sender %s", sender.String())
	}

	if !k.IsChainActivated(ctx, recipient.Chain) {
		return 0, fmt.Errorf("recipient chain '%s' is not activated", recipient.Chain.Name)
	}

	if !recipient.Chain.SupportsForeignAssets && !(isNativeAsset && sender.Chain.Name == chain.Name) {
		return 0, fmt.Errorf("recipient's chain %s does not support foreign assets", recipient.Chain.Name)
	}

	// merging incomplete transfers for the specified recipient
	incompleteTransfer, found := k.getTransferForRecipientAndAsset(ctx, recipient, asset.Denom, exported.Incomplete)
	if found {
		asset = asset.Add(incompleteTransfer.Asset)
		k.deleteTransfer(ctx, incompleteTransfer)
	}

	// collect fee
	fees := k.computeTransferFee(ctx, sender.Chain, recipient.Chain, asset)
	if fees.Amount.GTE(asset.Amount) {
		k.Logger(ctx).Debug(fmt.Sprintf("skipping deposit for chain %s at %s from recipient %s due to deposited amount being below "+
			"fees %s for asset %s", sender.Chain.Name, sender.Address, recipient.Address, fees.String(), asset.String()))

		return k.setNewIncompleteTransfer(ctx, recipient, asset), nil
	}

	if fees.IsPositive() {
		k.addTransferFee(ctx, fees)
		asset = asset.Sub(fees)
	}

	// merging transfers for the specified recipient
	previousTransfer, found := k.getTransferForRecipientAndAsset(ctx, recipient, asset.Denom, exported.Pending)
	if found {
		asset = asset.Add(previousTransfer.Asset)
		k.deleteTransfer(ctx, previousTransfer)
	}

	k.Logger(ctx).Info(fmt.Sprintf("Transfer of %s to cross chain address %s in %s successfully prepared",
		asset.String(), recipient.Address, recipient.Chain.Name))

	return k.setNewPendingTransfer(ctx, recipient, asset), nil
}

func (k Keeper) getTransferForRecipientAndAsset(ctx sdk.Context, recipient exported.CrossChainAddress, denom string, state exported.TransferState) (exported.CrossChainTransfer, bool) {
	iter := k.getStore(ctx).Iterator(getTransferPrefix(recipient.Chain.Name, state))
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
}

// GetTransfersForChain returns the current set of transfers with the given state for the given chain
func (k Keeper) GetTransfersForChain(ctx sdk.Context, chain exported.Chain, state exported.TransferState) (transfers []exported.CrossChainTransfer) {
	if !k.IsChainActivated(ctx, chain) {
		return transfers
	}

	iter := k.getStore(ctx).Iterator(getTransferPrefix(chain.Name, state))
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var transfer exported.CrossChainTransfer
		iter.UnmarshalValue(&transfer)

		transfers = append(transfers, transfer)
	}

	return transfers
}

// GetTransfersForChainPaginated returns the current set of transfers with the given state for the given chain with the given pagination properties
func (k Keeper) GetTransfersForChainPaginated(ctx sdk.Context, chain exported.Chain, state exported.TransferState, pageRequest *query.PageRequest) ([]exported.CrossChainTransfer, *query.PageResponse, error) {
	var transfers []exported.CrossChainTransfer
	if !k.IsChainActivated(ctx, chain) {
		return transfers, &query.PageResponse{}, nil
	}

	resp, err := query.Paginate(prefix.NewStore(k.getStore(ctx).KVStore, getTransferPrefix(chain.Name, state).AsKey()), pageRequest, func(key []byte, value []byte) error {
		var transfer exported.CrossChainTransfer
		k.cdc.MustUnmarshalLengthPrefixed(value, &transfer)

		transfers = append(transfers, transfer)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return transfers, resp, nil
}

// addTransferFee adds transfer fee
func (k Keeper) addTransferFee(ctx sdk.Context, coin sdk.Coin) {
	fee := k.getTransferFee(ctx)
	fee.Coins = fee.Coins.Add(coin)
	k.setTransferFee(ctx, fee)
}

// GetTransferFees returns the accumulated transfer fees
func (k Keeper) GetTransferFees(ctx sdk.Context) sdk.Coins {
	return k.getTransferFee(ctx).Coins
}

// SubTransferFee subtracts coin from transfer fee
func (k Keeper) SubTransferFee(ctx sdk.Context, coin sdk.Coin) {
	fee := k.getTransferFee(ctx)
	fee.Coins = fee.Coins.Sub(sdk.NewCoins(coin))
	k.setTransferFee(ctx, fee)
}
