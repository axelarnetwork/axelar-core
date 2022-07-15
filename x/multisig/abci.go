package multisig

import (
	"github.com/axelarnetwork/axelar-core/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

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
	for _, keygen := range k.GetKeygenSessionsByExpiry(ctx, ctx.BlockHeight()) {
		k.DeleteKeygenSession(ctx, keygen.GetKeyID())

		if keygen.State != exported.Completed {
			funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(types.NewKeygenExpired(keygen.GetKeyID())))
			k.Logger(ctx).Info("keygen session expired",
				"key_id", keygen.GetKeyID(),
			)

			slices.ForEach(keygen.GetMissingParticipants(), rewarder.GetPool(ctx, types.ModuleName).ClearRewards)

			continue
		}

		k.SetKey(ctx, funcs.Must(keygen.Result()))
	}
}

func handleSignings(ctx sdk.Context, k types.Keeper, rewarder types.Rewarder) {
	for _, signing := range k.GetSigningSessionsByExpiry(ctx, ctx.BlockHeight()) {

		_ = utils.RunCached(ctx, k, func(cachedCtx sdk.Context) ([]abci.ValidatorUpdate, error) {
			k.DeleteSigningSession(cachedCtx, signing.GetID())
			module := signing.GetModule()

			if signing.State != exported.Completed {
				funcs.MustNoErr(cachedCtx.EventManager().EmitTypedEvent(types.NewSigningExpired(signing.GetID())))
				k.Logger(cachedCtx).Info("signing session expired",
					"sig_id", signing.GetID(),
				)

				funcs.MustNoErr(k.GetSigRouter().GetHandler(module).HandleFailed(cachedCtx, signing.GetMetadata()))
				slices.ForEach(signing.GetMissingParticipants(), rewarder.GetPool(cachedCtx, types.ModuleName).ClearRewards)
				return nil, nil
			}

			sig := funcs.Must(signing.Result())

			funcs.MustNoErr(k.GetSigRouter().GetHandler(module).HandleCompleted(cachedCtx, &sig, signing.GetMetadata()))

			funcs.MustNoErr(cachedCtx.EventManager().EmitTypedEvent(types.NewSigningCompleted(signing.GetID())))
			k.Logger(cachedCtx).Info("signing session completed",
				"sig_id", signing.GetID(),
				"key_id", sig.GetKeyID(),
				"module", module,
			)

			return nil, nil
		})
	}
}
