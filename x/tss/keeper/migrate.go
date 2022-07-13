package keeper

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/utils"
	exported1 "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.19 to v0.20. The
// migration includes:
// - migrating active old keys, the current active key and the next key in rotation for all chains to the multisig module
func GetMigrationHandler(tss Keeper, multisig types.MultiSigKeeper, nexus types.Nexus, snapshotter types.Snapshotter) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		// suppress all events during migration
		mgr := ctx.EventManager()
		ctx.WithEventManager(sdk.NewEventManager())

		migrateKeys(ctx, tss, multisig, nexus, snapshotter)

		ctx.WithEventManager(mgr)
		return nil
	}
}

func migrateKeys(ctx sdk.Context, tss Keeper, multisig types.MultiSigKeeper, nexus types.Nexus, snapshotter types.Snapshotter) {
	chains := nexus.GetChains(ctx)
	for _, chain := range chains {
		keys, err := tss.GetOldActiveKeys(ctx, chain, exported.SecondaryKey)
		if err != nil {
			tss.Logger(ctx).Error(fmt.Sprintf("failed to migrate old %s keys for chain %s", exported.SecondaryKey.SimpleString(), chain.Name))
		} else {
			migrate(ctx, tss, multisig, snapshotter, chain.Name, keys...)
		}

		key, ok := tss.GetCurrentKey(ctx, chain, exported.SecondaryKey)
		if !ok {
			tss.Logger(ctx).Error(fmt.Sprintf("failed to migrate active %s key for chain %s", exported.SecondaryKey.SimpleString(), chain.Name))
		} else {
			migrate(ctx, tss, multisig, snapshotter, chain.Name, key)
		}

		key, ok = tss.GetNextKey(ctx, chain, exported.SecondaryKey)
		if !ok {
			tss.Logger(ctx).Error(fmt.Sprintf("failed to migrate next %s key in rotation for chain %s", exported.SecondaryKey.SimpleString(), chain.Name))
		} else {
			migrate(ctx, tss, multisig, snapshotter, chain.Name, key)
		}
	}
}

func migrate(ctx sdk.Context, tss Keeper, multisig types.MultiSigKeeper, snapshotter types.Snapshotter, chain nexus.ChainName, keys ...exported.Key) {
keyLoop:
	for _, key := range keys {
		s, ok := snapshotter.GetSnapshot(ctx, key.SnapshotCounter)
		if !ok {
			tss.Logger(ctx).Error(fmt.Sprintf("failed to migrate key %s for chain %s, no snapshot found", key.ID, chain))
			continue
		}
		info, found := tss.GetMultisigKeygenInfo(ctx, key.ID)
		keyInfo, ok := info.(*types.MultisigInfo)
		if !found || !ok {
			tss.Logger(ctx).Error(fmt.Sprintf("failed to migrate key %s for chain %s, key info not found", key.ID, chain))
		}
		pubkeys := make(map[string]multisigTypes.PublicKey)
		for _, infos := range keyInfo.Infos {

			val := infos.Participant.String()
			keys := infos.Data
			if len(keys) < 1 {
				tss.Logger(ctx).Error(fmt.Sprintf("failed to migrate key %s for chain %s, validator %s is participant without pubkey", key.ID, chain, infos.Participant.String()))
				continue keyLoop
			}
			pubkeys[val] = keys[0]
		}
		multisig.SetKey(ctx, multisigTypes.Key{
			ID:               exported1.KeyID(key.ID),
			Snapshot:         s,
			PubKeys:          pubkeys,
			SigningThreshold: utils.NewThreshold(keyInfo.TargetNum, keyInfo.Count()),
		})

		tss.Logger(ctx).Debug(fmt.Sprintf("successfully migrated %s key %s for chain %s", exported.SecondaryKey, key.ID, chain))
	}
}
