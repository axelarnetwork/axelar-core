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

func (k Keeper) setNewTransfer(ctx sdk.Context, recipient exported.CrossChainAddress, amount sdk.Coin, state exported.TransferState) exported.TransferID {
	id := k.getNonce(ctx)
	k.setTransfer(ctx, exported.NewCrossChainTransfer(id, recipient, amount, state))
	k.setNonce(ctx, id+1)
	return exported.TransferID(id)
}

func (k Keeper) setNewPendingTransfer(ctx sdk.Context, recipient exported.CrossChainAddress, amount sdk.Coin) exported.TransferID {
	return k.setNewTransfer(ctx, recipient, amount, exported.Pending)
}

func (k Keeper) setTransferFee(ctx sdk.Context, fee exported.TransferFee) {
	k.getStore(ctx).Set(transferFee, &fee)
}

func (k Keeper) getTransferFee(ctx sdk.Context) (fee exported.TransferFee) {
	k.getStore(ctx).Get(transferFee, &fee)
	return fee
}

// computeChainFee computes the fee for an asset transfer on a given chain
//
// chain_fee = min(max_fee, max(min_fee, fee_rate * amount))
func (k Keeper) computeChainFee(ctx sdk.Context, chain exported.Chain, asset sdk.Coin) sdk.Coin {
	feeInfo, _ := k.GetFeeInfo(ctx, chain, asset.Denom)

	fee := sdk.NewDecFromInt(asset.Amount).Mul(feeInfo.FeeRate).TruncateInt()
	fee = sdk.MaxInt(sdk.Int(feeInfo.MinFee), fee)
	fee = sdk.MinInt(sdk.Int(feeInfo.MaxFee), fee)

	return sdk.NewCoin(asset.Denom, fee)
}

// ComputeTransferFee computes the fee for a cross-chain transfer.
// If fee_info is not set for an asset on a chain, default of zero is used
//
// transfer_fee = min(total_max_fee, max(total_min_fee, (total_fee_rate) * amount))
//
// INVARIANT: source_chain.min_fee + destination_chain.min_fee <= transfer_fee <= source_chain.max_fee + destination_chain.max_fee
func (k Keeper) ComputeTransferFee(ctx sdk.Context, sourceChain exported.Chain, destinationChain exported.Chain, asset sdk.Coin) (sdk.Coin, error) {
	sourceChainFeeInfo, _ := k.GetFeeInfo(ctx, sourceChain, asset.Denom)
	destinationChainFeeInfo, _ := k.GetFeeInfo(ctx, destinationChain, asset.Denom)

	feeRate := sourceChainFeeInfo.FeeRate.Add(destinationChainFeeInfo.FeeRate)
	if feeRate.GT(sdk.OneDec()) {
		return sdk.Coin{}, fmt.Errorf("total fee rate should not be greater than 1")
	}

	minFee := sourceChainFeeInfo.MinFee.Add(destinationChainFeeInfo.MinFee)
	maxFee := sourceChainFeeInfo.MaxFee.Add(destinationChainFeeInfo.MaxFee)

	fee := sdk.NewDecFromInt(asset.Amount).Mul(feeRate).TruncateInt()
	fee = sdk.MaxInt(sdk.Int(minFee), fee)
	fee = sdk.MinInt(sdk.Int(maxFee), fee)

	return sdk.NewCoin(asset.Denom, fee), nil
}

// EnqueueForTransfer appoints the amount of tokens to be transferred/minted to the recipient previously linked to the specified sender
func (k Keeper) EnqueueForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, asset sdk.Coin) (exported.TransferID, error) {
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

	// merging transfers below minimum for the specified recipient
	insufficientAmountTransfer, found := k.getTransfer(ctx, recipient, asset.Denom, exported.InsufficientAmount)
	if found {
		asset = asset.Add(insufficientAmountTransfer.Asset)
		k.deleteTransfer(ctx, insufficientAmountTransfer)
	}

	// collect fee
	fee, err := k.ComputeTransferFee(ctx, sender.Chain, recipient.Chain, asset)
	if err != nil {
		return 0, err
	}

	if fee.Amount.GTE(asset.Amount) {
		k.Logger(ctx).Debug(fmt.Sprintf("skipping deposit for chain %s at %s from recipient %s due to deposited amount being below fees %s for asset %s",
			sender.Chain.Name, sender.Address, recipient.Address, fee.String(), asset.String()))

		return k.setNewTransfer(ctx, recipient, asset, exported.InsufficientAmount), nil
	}

	if fee.IsPositive() {
		k.addTransferFee(ctx, fee)
		asset = asset.Sub(fee)
	}

	// merging transfers for the specified recipient
	previousTransfer, found := k.getTransfer(ctx, recipient, asset.Denom, exported.Pending)
	if found {
		asset = asset.Add(previousTransfer.Asset)
		k.deleteTransfer(ctx, previousTransfer)
	}

	k.Logger(ctx).Info(fmt.Sprintf("Transfer of %s from %s in %s to cross chain address %s in %s successfully prepared",
		asset.String(), sender.Address, sender.Chain.Name, recipient.Address, recipient.Chain.Name))

	return k.setNewPendingTransfer(ctx, recipient, asset), nil
}

func (k Keeper) getTransfer(ctx sdk.Context, recipient exported.CrossChainAddress, denom string, state exported.TransferState) (exported.CrossChainTransfer, bool) {
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
