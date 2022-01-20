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

// EnqueueForTransfer appoints the amount of tokens to be transferred/minted to the recipient previously linked to the specified sender
func (k Keeper) EnqueueForTransfer(ctx sdk.Context, sender exported.CrossChainAddress, asset sdk.Coin, feeRate sdk.Dec) (exported.TransferID, error) {
	if !sender.Chain.SupportsForeignAssets && sender.Chain.NativeAsset != asset.Denom {
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

	if !recipient.Chain.SupportsForeignAssets && recipient.Chain.NativeAsset != asset.Denom {
		return 0, fmt.Errorf("recipient's chain %s does not support foreign assets", recipient.Chain.Name)
	}

	// collect fee
	if feeDue := sdk.NewDecFromInt(asset.Amount).Mul(feeRate).TruncateInt(); feeDue.IsPositive() {
		k.addTransferFee(ctx, sdk.NewCoin(asset.Denom, feeDue))
		asset = asset.SubAmount(feeDue)
	}

	// merging transfers for the specified recipient
	previousTransfer, found := k.getPendingTransferForRecipientAndAsset(ctx, recipient, asset.Denom)
	if found {
		asset = asset.Add(previousTransfer.Asset)
		k.deleteTransfer(ctx, previousTransfer)
	}

	k.Logger(ctx).Info(fmt.Sprintf("Transfer of %s to cross chain address %s in %s successfully prepared",
		asset.String(), recipient.Address, recipient.Chain.Name))

	return k.setNewPendingTransfer(ctx, recipient, asset), nil
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
}

// GetTransfersForChain returns the current set of transfers with the given state for the given chain
func (k Keeper) GetTransfersForChain(ctx sdk.Context, chain exported.Chain, state exported.TransferState) (transfers []exported.CrossChainTransfer) {
	if !k.IsChainActivated(ctx, chain) {
		return transfers
	}

	iter := k.getStore(ctx).Iterator(getTransferPrefix(chain.Name, state))
	defer utils.CloseLogError(iter, k.Logger(ctx))
	minAmountCache := map[string]sdk.Int{}

	for ; iter.Valid(); iter.Next() {
		var transfer exported.CrossChainTransfer
		iter.UnmarshalValue(&transfer)

		asset := transfer.Asset.Denom
		if minAmountCache[asset].IsNil() {
			minAmountCache[asset] = k.GetMinAmount(ctx, chain, asset)
		}

		if transfer.Asset.Amount.LT(minAmountCache[asset]) {
			k.Logger(ctx).Debug(fmt.Sprintf("skipping deposit for chain %s from recipient %s due to deposited amount being below "+
				"minimum amount for asset %s", chain.Name, transfer.Recipient.Address, asset))
			continue
		}

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
