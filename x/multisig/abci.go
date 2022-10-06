package multisig

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// BeginBlocker is called at the beginning of every block
func BeginBlocker(sdk.Context, abci.RequestBeginBlock) {}

// EndBlocker is called at the end of every block, process external chain voting inflation
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, k types.Keeper, rewarder types.Rewarder) ([]abci.ValidatorUpdate, error) {
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

		key := funcs.Must(keygen.Result())

		slices.ForEach(key.GetParticipants(), func(p sdk.ValAddress) { funcs.MustNoErr(pool.ReleaseRewards(p)) })
		k.SetKey(ctx, key)
	}
}

func handleSignings(ctx sdk.Context, k types.Keeper, rewarder types.Rewarder) {
	// we handle sessions that'll expire on the next block,
	// to avoid waiting for an additional block
	for _, signing := range k.GetSigningSessionsByExpiry(ctx, ctx.BlockHeight()+1) {
		_ = utils.RunCached(ctx, k, func(cachedCtx sdk.Context) ([]abci.ValidatorUpdate, error) {
			k.DeleteSigningSession(cachedCtx, signing.GetID())
			module := signing.GetModule()

			pool := rewarder.GetPool(cachedCtx, types.ModuleName)
			slices.ForEach(signing.GetMissingParticipants(), pool.ClearRewards)

			if signing.State != exported.Completed {
				events.Emit(cachedCtx, types.NewSigningExpired(signing.GetID()))
				k.Logger(cachedCtx).Info("signing session expired",
					"sig_id", signing.GetID(),
				)

				funcs.MustNoErr(k.GetSigRouter().GetHandler(module).HandleFailed(cachedCtx, signing.GetMetadata()))
				return nil, nil
			}

			sig := funcs.Must(signing.Result())

			slices.ForEach(sig.GetParticipants(), func(p sdk.ValAddress) { funcs.MustNoErr(pool.ReleaseRewards(p)) })
			if err := k.GetSigRouter().GetHandler(module).HandleCompleted(cachedCtx, &sig, signing.GetMetadata()); err != nil {
				return nil, sdkerrors.Wrap(err, "failed to handle completed signature")
			}

			events.Emit(cachedCtx, types.NewSigningCompleted(signing.GetID()))
			k.Logger(cachedCtx).Info("signing session completed",
				"sig_id", signing.GetID(),
				"key_id", sig.GetKeyID(),
				"module", module,
			)

			return nil, nil
		})
	}
}
