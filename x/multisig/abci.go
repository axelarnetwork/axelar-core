package multisig

import (
	"cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// EndBlocker is called at the end of every block, process external chain voting inflation
func EndBlocker(ctx sdk.Context, k types.Keeper, rewarder types.Rewarder) ([]abci.ValidatorUpdate, error) {
	handleKeygens(ctx, k, rewarder)
	handleSignings(ctx, k, rewarder)

	return nil, nil
}

func handleKeygens(ctx sdk.Context, k types.Keeper, rewarder types.Rewarder) {
	// we handle sessions that'll expire on the next block,
	// to avoid waiting for an additional block
	for _, keygen := range k.GetKeygenSessionsByExpiry(ctx, ctx.BlockHeight()+1) {
		k.DeleteKeygenSession(ctx, keygen.GetKeyID())

		pool := rewarder.GetPool(ctx, types.ModuleName)
		slices.ForEach(keygen.GetMissingParticipants(), pool.ClearRewards)

		if keygen.State != exported.Completed {
			events.Emit(ctx, types.NewKeygenExpired(keygen.GetKeyID()))
			k.Logger(ctx).Info("keygen session expired",
				"key_id", keygen.GetKeyID(),
			)

			continue
		}

		completed := utils.RunCached(ctx, k, func(cachedCtx sdk.Context) (bool, error) {
			key := funcs.Must(keygen.Result())

			pool := rewarder.GetPool(cachedCtx, types.ModuleName)
			slices.ForEach(key.GetParticipants(), func(p sdk.ValAddress) { funcs.MustNoErr(pool.ReleaseRewards(p)) })
			k.SetKey(cachedCtx, key)

			return true, nil
		})

		if !completed {
			events.Emit(ctx, types.NewKeygenExpired(keygen.GetKeyID()))
			k.Logger(ctx).Error("failed to handle completed keygen session",
				"key_id", keygen.GetKeyID(),
			)
		}
	}
}

func handleSignings(ctx sdk.Context, k types.Keeper, rewarder types.Rewarder) {
	// we handle sessions that'll expire on the next block,
	// to avoid waiting for an additional block
	for _, signing := range k.GetSigningSessionsByExpiry(ctx, ctx.BlockHeight()+1) {
		k.DeleteSigningSession(ctx, signing.GetID())
		module := signing.GetModule()

		pool := rewarder.GetPool(ctx, types.ModuleName)
		slices.ForEach(signing.GetMissingParticipants(), pool.ClearRewards)

		if signing.State != exported.Completed {
			events.Emit(ctx, types.NewSigningExpired(signing.GetID()))
			k.Logger(ctx).Info("signing session expired",
				"sig_id", signing.GetID(),
			)

			abortSigning(ctx, k, signing)
			continue
		}

		completed := utils.RunCached(ctx, k, func(cachedCtx sdk.Context) (bool, error) {
			sig := funcs.Must(signing.Result())

			cachedPool := rewarder.GetPool(cachedCtx, types.ModuleName)
			slices.ForEach(sig.GetParticipants(), func(p sdk.ValAddress) { funcs.MustNoErr(cachedPool.ReleaseRewards(p)) })

			if err := k.GetSigRouter().GetHandler(module).HandleCompleted(cachedCtx, &sig, signing.GetMetadata()); err != nil {
				return false, errors.Wrap(err, "failed to handle completed signature")
			}

			events.Emit(cachedCtx, types.NewSigningCompleted(signing.GetID()))
			k.Logger(cachedCtx).Info("signing session completed",
				"sig_id", signing.GetID(),
				"key_id", sig.GetKeyID(),
				"module", module,
			)

			return true, nil
		})

		if !completed {
			k.Logger(ctx).Error("failed to handle completed signing session, aborting signing",
				"sig_id", signing.GetID(),
				"module", module,
			)
			abortSigning(ctx, k, signing)
		}
	}
}

func abortSigning(ctx sdk.Context, k types.Keeper, signing types.SigningSession) {
	_ = utils.RunCached(ctx, k, func(cachedCtx sdk.Context) (struct{}, error) {
		return struct{}{}, k.GetSigRouter().GetHandler(signing.GetModule()).HandleFailed(cachedCtx, signing.GetMetadata())
	})
}
