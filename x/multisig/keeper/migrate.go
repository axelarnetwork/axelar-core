package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.21 to v0.22. The
// migration includes:
// - fix the threshold of current keys of all chain that were migrated from the tss module
func GetMigrationHandler(k Keeper, t types.Tss, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		// suppress all events during migration
		mgr := ctx.EventManager()
		ctx.WithEventManager(sdk.NewEventManager())
		defer ctx.WithEventManager(mgr)

		return fixCurrentKeysThreshold(ctx, k, t, n)
	}
}

func fixCurrentKeysThreshold(ctx sdk.Context, k Keeper, t types.Tss, n types.Nexus) error {
	for _, chain := range n.GetChains(ctx) {
		keyID, ok := k.GetCurrentKeyID(ctx, chain.Name)
		if !ok {
			continue
		}

		tssKey, ok := t.GetKey(ctx, tss.KeyID(keyID))
		if !ok {
			continue
		}

		threshold := tssKey.GetPublicKey().(*tss.Key_MultisigKey_).MultisigKey.Threshold

		key := funcs.MustOk(k.getKey(ctx, keyID))
		key.SigningThreshold = utils.NewThreshold(threshold, int64(len(key.GetParticipants())))
		funcs.MustNoErr(key.ValidateBasic())

		k.setKey(ctx, key)
	}

	return nil
}
