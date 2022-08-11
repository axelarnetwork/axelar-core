package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	tss "github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.21 to v0.22. The
// migration includes:
// - fix the threshold of current and next keys of all chain that were migrated from the tss module
func GetMigrationHandler(k Keeper, t types.Tss, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		// suppress all events during migration
		mgr := ctx.EventManager()
		ctx.WithEventManager(sdk.NewEventManager())
		defer ctx.WithEventManager(mgr)

		for _, chain := range n.GetChains(ctx) {
			keyID, ok := k.GetCurrentKeyID(ctx, chain.Name)
			if !ok {
				k.Logger(ctx).Debug(fmt.Sprintf("current key is not set for chain %s, skip migration", chain.Name))
				continue
			}
			fixKeyThresholdIfFromTss(ctx, k, t, n, chain.Name, keyID)

			keyID, ok = k.GetNextKeyID(ctx, chain.Name)
			if !ok {
				k.Logger(ctx).Debug(fmt.Sprintf("next key is not set for chain %s, skip migration", chain.Name))
				continue
			}
			fixKeyThresholdIfFromTss(ctx, k, t, n, chain.Name, keyID)
		}

		return nil
	}
}

func fixKeyThresholdIfFromTss(ctx sdk.Context, k Keeper, t types.Tss, n types.Nexus, chainName nexus.ChainName, keyID exported.KeyID) {
	tssKey, ok := t.GetKey(ctx, tss.KeyID(keyID))
	if !ok {
		k.Logger(ctx).Debug(fmt.Sprintf("chain %s's key %s is not set in tss, skip migration", chainName, keyID))
		return
	}

	tssMultisigKey := tssKey.GetPublicKey().(*tss.Key_MultisigKey_).MultisigKey

	key := funcs.MustOk(k.getKey(ctx, keyID))
	key.SigningThreshold = utils.NewThreshold(tssMultisigKey.Threshold, int64(len(tssMultisigKey.Values)))
	funcs.MustNoErr(key.ValidateBasic())

	k.setKey(ctx, key)
	k.Logger(ctx).Debug("migrated signing threshold of key",
		"keyID", keyID,
		"chain", chainName,
		"signingThreshold", key.SigningThreshold.String(),
		"minPassingWeight", key.GetMinPassingWeight().String(),
		"participantsWeight", key.GetParticipantsWeight().String(),
	)
}
