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
	if !found {
		return nil
	}

	epoch := uint64(ctx.BlockTime().UnixNano() / rateLimit.Window.Nanoseconds())

	transferRate, found := k.getTransferRate(ctx, chain, asset.Denom, direction)
	if !found || transferRate.Epoch != epoch {
		transferRate = types.TransferRate{
			Chain:     chain,
			Amount:    sdk.NewCoin(asset.Denom, sdk.ZeroInt()),
			Epoch:     epoch,
			Direction: direction,
		}
	}

	transferRate.Amount = transferRate.Amount.Add(asset)

	if transferRate.Amount.Amount.GT(rateLimit.Limit.Amount) {
		err := fmt.Errorf("transfer %s for chain %s (%s) exceeded rate limit %s with transfer rate %s", asset, transferRate.Chain, transferRate.Direction, rateLimit.Limit, transferRate.Amount)
		k.Logger(ctx).Error(err.Error(), types.AttributeKeyChain, transferRate.Chain, types.AttributeKeyAsset, asset, types.AttributeKeyLimit, rateLimit.Limit, types.AttributeKeyTransferRate, transferRate.Amount)
		return sdkerrors.Wrap(types.ErrRateLimitExceeded, err.Error())
	}

	k.setTransferRate(ctx, transferRate)

	return nil
}

// SetRateLimit sets a rate limit for the given chain and asset
func (k Keeper) SetRateLimit(ctx sdk.Context, chainName exported.ChainName, limit sdk.Coin, window time.Duration) error {
	chain, ok := k.GetChain(ctx, chainName)
	if !ok {
		return fmt.Errorf("%s is not a registered chain", chainName)
	}

	// NOTE: We could potentially skip the Asset registered check.
	// There can be benefit of rate limiting denoms that are not registered as cross-chain assets, due to IBC
	if !k.IsAssetRegistered(ctx, chain, limit.Denom) {
		return fmt.Errorf("%s is not a registered for chain %s", limit.Denom, chain.Name)
	}

	k.deleteTransferRate(ctx, chain.Name, limit.Denom, exported.Incoming)
	k.deleteTransferRate(ctx, chain.Name, limit.Denom, exported.Outgoing)

	funcs.MustNoErr(k.getStore(ctx).SetNewValidated(getRateLimitKey(chain.Name, limit.Denom), &types.RateLimit{
		Chain:  chain.Name,
		Limit:  limit,
		Window: window,
	}))

	k.Logger(ctx).Info(fmt.Sprintf("transfer rate limit %s set for chain %s with window %s", chain.Name, limit, window))

	events.Emit(ctx, &types.RateLimitUpdated{
		Chain:  chain.Name,
		Limit:  limit,
		Window: window,
	})

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

func getTransferRateKey(chain exported.ChainName, asset string, direction exported.TransferDirection) key.Key {
	return transferRatePrefix.
		Append(key.From(chain)).
		Append(key.FromStr(asset)).
		Append(key.FromUInt(uint(direction)))
}

func (k Keeper) getTransferRate(ctx sdk.Context, chain exported.ChainName, asset string, direction exported.TransferDirection) (transferRate types.TransferRate, found bool) {
	return transferRate, k.getStore(ctx).GetNew(getTransferRateKey(chain, asset, direction), &transferRate)
}

func (k Keeper) setTransferRate(ctx sdk.Context, transferRate types.TransferRate) {
	funcs.MustNoErr(k.getStore(ctx).SetNewValidated(getTransferRateKey(transferRate.Chain, transferRate.Amount.Denom, transferRate.Direction), &transferRate))
}

func (k Keeper) deleteTransferRate(ctx sdk.Context, chain exported.ChainName, asset string, direction exported.TransferDirection) {
	k.getStore(ctx).DeleteNew(getTransferRateKey(exported.ChainName(chain), asset, direction))
}

func (k Keeper) getTransferRates(ctx sdk.Context) (transferRates []types.TransferRate) {
	iter := k.getStore(ctx).IteratorNew(transferRatePrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var transferRate types.TransferRate
		iter.UnmarshalValue(&transferRate)

		transferRates = append(transferRates, transferRate)
	}

	return transferRates
}
