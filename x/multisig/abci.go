package multisig

import (
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/keeper"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/utils/funcs"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// BeginBlocker is called at the beginning of every block
func BeginBlocker(sdk.Context, abci.RequestBeginBlock) {}

// EndBlocker is called at the end of every block, process external chain voting inflation
func EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock, k keeper.Keeper) []abci.ValidatorUpdate {
	for _, keygen := range k.GetKeygenSessionsByExpiry(ctx, ctx.BlockHeight()) {
		k.DeleteKeygenSession(ctx, keygen.GetKeyID())

		if keygen.State != exported.Completed {
			funcs.MustNoErr(ctx.EventManager().EmitTypedEvent(types.NewKeygenExpired(keygen.GetKeyID())))
			k.Logger(ctx).Info("keygen session expired",
				"key_id", keygen.GetKeyID(),
			)

			continue
		}

		key := funcs.Must(keygen.Result())
		k.SetKey(ctx, key)
	}

	return nil
}
