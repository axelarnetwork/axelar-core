package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

func getTransferPrefix(chain exported.ChainName, state exported.TransferState) utils.Key {
	return transferPrefix.
		AppendStr(state.String()).
		AppendStr(chain.String())
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

func (k Keeper) setTransferFee(ctx sdk.Context, fee exported.TransferFee) {
	k.getStore(ctx).Set(transferFee, &fee)
}

func (k Keeper) getTransferFee(ctx sdk.Context) (fee exported.TransferFee) {
	k.getStore(ctx).Get(transferFee, &fee)
	return fee
}

// getCrossChainFees computes the fee info for a cross-chain transfer.
func (k Keeper) getCrossChainFees(ctx sdk.Context, sourceChain exported.Chain, destinationChain exported.Chain, asset string) (feeRate sdk.Dec, minFee sdk.Int, maxFee sdk.Int, err error) {
	sourceChainFeeInfo, _ := k.GetFeeInfo(ctx, sourceChain, asset)
	destinationChainFeeInfo, _ := k.GetFeeInfo(ctx, destinationChain, asset)

	feeRate = sourceChainFeeInfo.FeeRate.Add(destinationChainFeeInfo.FeeRate)
	if feeRate.GT(sdk.OneDec()) {
		return sdk.Dec{}, sdk.Int{}, sdk.Int{}, fmt.Errorf("total fee rate should not be greater than 1")
	}

	minFee = sourceChainFeeInfo.MinFee.Add(destinationChainFeeInfo.MinFee)
	maxFee = sourceChainFeeInfo.MaxFee.Add(destinationChainFeeInfo.MaxFee)

	return feeRate, minFee, maxFee, nil
}

// ComputeTransferFee computes the fee for a cross-chain transfer.
// If fee_info is not set for an asset on a chain, default of zero is used
//
// transfer_fee = min(total_max_fee, max(total_min_fee, (total_fee_rate) * amount))
//
// INVARIANT: source_chain.min_fee + destination_chain.min_fee <= transfer_fee <= source_chain.max_fee + destination_chain.max_fee
func (k Keeper) ComputeTransferFee(ctx sdk.Context, sourceChain exported.Chain, destinationChain exported.Chain, asset sdk.Coin) (sdk.Coin, error) {
	feeRate, minFee, maxFee, err := k.getCrossChainFees(ctx, sourceChain, destinationChain, asset.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	fee := sdk.NewDecFromInt(asset.Amount).Mul(feeRate).TruncateInt()
	fee = sdk.MaxInt(minFee, fee)
	fee = sdk.MinInt(maxFee, fee)

	return sdk.NewCoin(asset.Denom, fee), nil
}

// EnqueueTransfer enqueues an asset transfer to the given recipient address
func (k Keeper) EnqueueTransfer(ctx sdk.Context, senderChain exported.Chain, recipient exported.CrossChainAddress, asset sdk.Coin) (exported.TransferID, error) {
	if err := k.validateTransferAsset(ctx, senderChain, asset.Denom); err != nil {
		return 0, err
	}

	if err := k.validateTransferAsset(ctx, recipient.Chain, asset.Denom); err != nil {
		return 0, err
	}

	if validator := k.GetRouter().GetAddressValidator(recipient.Chain.Module); validator == nil {
		return 0, fmt.Errorf("unknown module for recipient chain %s", recipient.Chain.String())
	} else if err := validator(ctx, recipient); err != nil {
		return 0, err
	}

	// merging transfers below minimum for the specified recipient
	insufficientAmountTransfer, found := k.getTransfer(ctx, recipient, asset.Denom, exported.InsufficientAmount)
	if found {
		asset = asset.Add(insufficientAmountTransfer.Asset)
		k.deleteTransfer(ctx, insufficientAmountTransfer)
	}

	// collect fee
	fee, err := k.ComputeTransferFee(ctx, senderChain, recipient.Chain, asset)
	if err != nil {
		return 0, err
	}

	if fee.Amount.GTE(asset.Amount) {
		k.Logger(ctx).Debug(fmt.Sprintf("skipping deposit from chain %s to chain %s and recipient %s due to deposited amount being below fees %s for asset %s",
			senderChain.Name, recipient.Chain.Name, recipient.Address, fee.String(), asset.String()))

		transferID := k.setNewTransfer(ctx, recipient, asset, exported.InsufficientAmount)

		events.Emit(ctx, &types.InsufficientFee{
			TransferID:       transferID,
			RecipientChain:   recipient.Chain.Name,
			RecipientAddress: recipient.Address,
			Amount:           asset,
			Fee:              fee,
		})

		return transferID, nil
	}

	if fee.IsPositive() {
		k.AddTransferFee(ctx, fee)
		asset = asset.Sub(fee)
	}

	// merging transfers for the specified recipient
	previousTransfer, found := k.getTransfer(ctx, recipient, asset.Denom, exported.Pending)
	if found {
		asset = asset.Add(previousTransfer.Asset)
		k.deleteTransfer(ctx, previousTransfer)
	}

	k.Logger(ctx).Info(fmt.Sprintf("transfer %s from chain %s to chain %s and recipient %s is successfully prepared",
		asset.String(), senderChain.Name, recipient.Chain.Name, recipient.Address))

	transferID := k.setNewTransfer(ctx, recipient, asset, exported.Pending)

	events.Emit(ctx, &types.FeeDeducted{
		TransferID:       transferID,
		RecipientChain:   recipient.Chain.Name,
		RecipientAddress: recipient.Address,
		Amount:           asset,
		Fee:              fee,
	})

	return transferID, nil
}

func (k Keeper) validateTransferAsset(ctx sdk.Context, chain exported.Chain, asset string) error {
	if chain.SupportsForeignAssets {
		return nil
	}

	nativeChain, hasNativeChain := k.GetChainByNativeAsset(ctx, asset)
	if !hasNativeChain || nativeChain.Name != chain.Name {
		return fmt.Errorf("chain %s does not support foreign asset %s", chain.Name, asset)
	}

	return nil
}

// EnqueueForTransfer enqueues an asset transfer for the given deposit address
func (k Keeper) EnqueueForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, asset sdk.Coin) (exported.TransferID, error) {
	recipient, ok := k.GetRecipient(ctx, sender)
	if !ok {
		return 0, fmt.Errorf("no recipient linked to sender %s", sender.String())
	}

	return k.EnqueueTransfer(ctx, sender.Chain, recipient, asset)
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

// AddTransferFee adds transfer fee
func (k Keeper) AddTransferFee(ctx sdk.Context, coin sdk.Coin) {
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
