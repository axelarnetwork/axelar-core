package keeper

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/utils/funcs"
)

func getRateLimitKey(chain exported.ChainName, asset string) key.Key {
	return rateLimitPrefix.
		Append(key.From(chain)).
		Append(key.FromStr(asset))
}

func (k Keeper) getRateLimit(ctx sdk.Context, chain exported.ChainName, asset string) (rateLimit types.RateLimit, found bool) {
	return rateLimit, k.getStore(ctx).GetNew(getRateLimitKey(chain, asset), &rateLimit)
}

func (k Keeper) SetRateLimitStore(ctx sdk.Context, chain exported.ChainName, limit sdk.Coin, window time.Duration) {
	funcs.MustNoErr(k.getStore(ctx).SetNewValidated(getRateLimitKey(chain, limit.Denom), &types.RateLimit{
		Chain:  chain,
		Limit:  limit,
		Window: window,
	}))
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

func getTransferAmountKey(chain exported.ChainName, asset string, outgoing bool) key.Key {
	return transferAmountPrefix.
		Append(key.From(chain)).
		Append(key.FromStr(asset)).
		Append(key.FromBool(outgoing))
}

func (k Keeper) getTransferAmount(ctx sdk.Context, chain exported.ChainName, asset string, outgoing bool) (transferAmount types.TransferAmount, found bool) {
	return transferAmount, k.getStore(ctx).GetNew(getTransferAmountKey(chain, asset, outgoing), &transferAmount)
}

func (k Keeper) setTransferAmount(ctx sdk.Context, transferAmount types.TransferAmount) {
	funcs.MustNoErr(k.getStore(ctx).SetNewValidated(getTransferAmountKey(transferAmount.Chain, transferAmount.Amount.Denom, transferAmount.Outgoing), &transferAmount))
}

func (k Keeper) getTransferAmounts(ctx sdk.Context) (transferAmounts []types.TransferAmount) {
	iter := k.getStore(ctx).IteratorNew(transferAmountPrefix)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var transferAmount types.TransferAmount
		iter.UnmarshalValue(&transferAmount)

		transferAmounts = append(transferAmounts, transferAmount)
	}

	return transferAmounts
}

// Get the epoch for the transfer
func computeEpoch(ctx sdk.Context, window time.Duration) uint64 {
	return uint64(ctx.BlockTime().UnixNano() / window.Nanoseconds())
}

func (k Keeper) RateLimitTransfer(ctx sdk.Context, chain exported.ChainName, asset sdk.Coin, outgoing bool) error {
	rateLimit, found := k.getRateLimit(ctx, chain, asset.Denom)
	if !found {
		return nil
	}

	epoch := computeEpoch(ctx, rateLimit.Window)

	transferAmount, found := k.getTransferAmount(ctx, chain, asset.Denom, outgoing)
	if !found || transferAmount.Epoch != epoch {
		transferAmount = types.TransferAmount{
			Chain:    chain,
			Amount:   sdk.NewCoin(asset.Denom, sdk.ZeroInt()),
			Epoch:    epoch,
			Outgoing: outgoing,
		}
	}

	transferAmount.Amount = transferAmount.Amount.Add(asset)

	if transferAmount.Amount.Amount.GT(rateLimit.Limit.Amount) {
		err := fmt.Errorf("transfer %s for chain %s (outgoing: %v) exceeded rate limit %s", transferAmount.Amount, transferAmount.Chain, transferAmount.Outgoing, rateLimit.Limit)
		k.Logger(ctx).Error(err.Error())
		return err
	}

	k.setTransferAmount(ctx, transferAmount)

	return nil
}
