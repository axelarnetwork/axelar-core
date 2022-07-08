package multisig

import (
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

	return nil, nil
}

func handleKeygens(ctx sdk.Context, k types.Keeper, rewarder types.Rewarder) {
	for _, keygen := range k.GetKeygenSessionsByExpiry(ctx, ctx.BlockHeight()) {
		k.DeleteKeygenSession(ctx, keygen.GetKeyID())

		if keygen.State == exported.Completed {
			k.SetKey(ctx, funcs.Must(keygen.Result()))

			continue
		}

		funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(types.NewKeygenExpired(keygen.GetKeyID())))
		k.Logger(ctx).Info("keygen session expired",
			"key_id", keygen.GetKeyID(),
		)

		slices.ForEach(keygen.GetMissingParticipants(), rewarder.GetPool(ctx, types.ModuleName).ClearRewards)
	}
}
