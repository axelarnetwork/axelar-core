package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
)

// RateLimitTransfer applies a rate limit to transfers, and returns an error if the rate limit is exceeded
func (k Keeper) RateLimitTransfer(ctx sdk.Context, chain exported.ChainName, asset sdk.Coin, direction exported.TransferDirection) error {
	rateLimit, found := k.getRateLimit(ctx, chain, asset.Denom)
	// If a rate limit is not set, it is treated as unbounded
	if !found {
		return nil
	}

	transferEpoch := k.getCurrentTransferEpoch(ctx, chain, asset.Denom, direction, rateLimit.Window)
	transferEpoch.Amount = transferEpoch.Amount.Add(asset)

	if transferEpoch.Amount.Amount.GT(rateLimit.Limit.Amount) {
		directionLog := "from"
		if direction == exported.TransferDirectionTo {
			directionLog = "to"
		}

		err := fmt.Errorf("transfer of %s %s chain %s exceeded rate limit %s with transfer rate %s", asset, directionLog, transferEpoch.Chain, rateLimit.Limit, transferEpoch.Amount)
		k.Logger(ctx).Error(err.Error(),
			types.AttributeKeyChain, transferEpoch.Chain,
			types.AttributeKeyAsset, asset,
			types.AttributeKeyLimit, rateLimit.Limit,
			types.AttributeKeyTransferEpoch, transferEpoch.Amount,
			types.AttributeKeyBlock, ctx.BlockHeight(),
		)
		return sdkerrors.Wrap(types.ErrRateLimitExceeded, err.Error())
	}

	k.setTransferEpoch(ctx, transferEpoch)

	return nil
}

// SetRateLimit sets a rate limit for the given chain and asset.
// If max uint256 is provided as a limit, it's treated as a rate limit being infinite/not being set.
func (k Keeper) SetRateLimit(ctx sdk.Context, chainName exported.ChainName, limit sdk.Coin, window time.Duration) error {
	chain, ok := k.GetChain(ctx, chainName)
	if !ok {
		return fmt.Errorf("%s is not a registered chain", chainName)
	}

	// NOTE: We could potentially skip the Asset registered check.
	// There can be benefit of rate limiting denoms that are not registered as cross-chain assets, due to IBC
	if !k.IsAssetRegistered(ctx, chain, limit.Denom) {
		return fmt.Errorf("%s is not a registered asset for chain %s", limit.Denom, chain.Name)
	}

	events.Emit(ctx, &types.RateLimitUpdated{
		Chain:  chain.Name,
		Limit:  limit,
		Window: window,
	})

	// delete any rate limit info if provided limit is max uint256
	if limit.Amount.Equal(sdk.NewIntFromBigInt(utils.MaxUint.BigInt())) {
		k.getStore(ctx).DeleteNew(getRateLimitKey(chain.Name, limit.Denom))
		k.deleteTransferEpoch(ctx, chain.Name, limit.Denom, exported.TransferDirectionFrom)
		k.deleteTransferEpoch(ctx, chain.Name, limit.Denom, exported.TransferDirectionTo)
		return nil
	}

	if err := k.getStore(ctx).SetNewValidated(getRateLimitKey(chain.Name, limit.Denom), &types.RateLimit{
		Chain:  chain.Name,
		Limit:  limit,
		Window: window,
	}); err != nil {
		return err
	}

	epoch := computeEpoch(ctx, window)

	k.setTransferEpoch(ctx, types.NewTransferEpoch(chain.Name, limit.Denom, epoch, exported.TransferDirectionFrom))
	k.setTransferEpoch(ctx, types.NewTransferEpoch(chain.Name, limit.Denom, epoch, exported.TransferDirectionTo))

	k.Logger(ctx).Info(fmt.Sprintf("transfer rate limit %s set for chain %s with window %s", chain.Name, limit, window))

	return nil
}

func getRateLimitKey(chain exported.ChainName, asset string) key.Key {
	return rateLimitPrefix.
		Append(key.From(chain)).
		Append(key.FromStr(asset))
}

func (k Keeper) getRateLimit(ctx sdk.Context, chain exported.ChainName, asset string) (rateLimit types.RateLimit, found bool) {
	return rateLimit, k.getStore(ctx).GetNew(getRateLimitKey(chain, asset), &rateLimit)
}

func (k Keeper) getRateLimits(ctx sdk.Context) (rateLimits []types.RateLimit) {
	iter := k.getStore(ctx).IteratorNew(rateLimitPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var rateLimit types.RateLimit
		iter.UnmarshalValue(&rateLimit)

		rateLimits = append(rateLimits, rateLimit)
	}

	return rateLimits
}

func getTransferEpochKey(chain exported.ChainName, asset string, direction exported.TransferDirection) key.Key {
	return transferEpochPrefix.
		Append(key.From(chain)).
		Append(key.FromStr(asset)).
		Append(key.FromUInt(uint(direction)))
}

func (k Keeper) getTransferEpoch(ctx sdk.Context, chain exported.ChainName, asset string, direction exported.TransferDirection) (transferEpoch types.TransferEpoch, found bool) {
	return transferEpoch, k.getStore(ctx).GetNew(getTransferEpochKey(chain, asset, direction), &transferEpoch)
}

func (k Keeper) getCurrentTransferEpoch(ctx sdk.Context, chain exported.ChainName, asset string, direction exported.TransferDirection, window time.Duration) types.TransferEpoch {
	// use a new transfer epoch if there was none or if the epoch is outdated
	epoch := computeEpoch(ctx, window)
	if transferEpoch, found := k.getTransferEpoch(ctx, chain, asset, direction); found && transferEpoch.Epoch == epoch {
		return transferEpoch
	}

	return types.NewTransferEpoch(chain, asset, epoch, direction)
}

func computeEpoch(ctx sdk.Context, window time.Duration) uint64 {
	return uint64(ctx.BlockTime().UnixNano() / window.Nanoseconds())
}

func (k Keeper) setTransferEpoch(ctx sdk.Context, transferEpoch types.TransferEpoch) {
	funcs.MustNoErr(k.getStore(ctx).SetNewValidated(getTransferEpochKey(transferEpoch.Chain, transferEpoch.Amount.Denom, transferEpoch.Direction), &transferEpoch))
}

func (k Keeper) deleteTransferEpoch(ctx sdk.Context, chain exported.ChainName, asset string, direction exported.TransferDirection) {
	k.getStore(ctx).DeleteNew(getTransferEpochKey(chain, asset, direction))
}

func (k Keeper) getTransferEpochs(ctx sdk.Context) (transferEpochs []types.TransferEpoch) {
	iter := k.getStore(ctx).IteratorNew(transferEpochPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var transferEpoch types.TransferEpoch
		iter.UnmarshalValue(&transferEpoch)

		transferEpochs = append(transferEpochs, transferEpoch)
	}

	return transferEpochs
}
