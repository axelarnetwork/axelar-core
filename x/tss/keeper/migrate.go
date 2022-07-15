package keeper

import (
	"fmt"

	"github.com/axelarnetwork/axelar-core/utils"
	exported1 "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshotexported "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
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
		defer ctx.WithEventManager(mgr)

		return migrateKeys(ctx, tss, multisig, nexus, snapshotter)
	}
}

func migrateKeys(ctx sdk.Context, tss Keeper, multisig types.MultiSigKeeper, nexus types.Nexus, snapshotter types.Snapshotter) error {
	chains := nexus.GetChains(ctx)
	for _, chain := range chains {
		keys, err := tss.GetOldActiveKeys(ctx, chain, exported.SecondaryKey)
		if err != nil {
			return fmt.Errorf("failed to migrate old %s keys for chain %s", exported.SecondaryKey.SimpleString(), chain.Name)
		}

		if err := migrate(ctx, tss, multisig, snapshotter, chain.Name, keys...); err != nil {
			return err
		}

		key, ok := tss.GetCurrentKey(ctx, chain, exported.SecondaryKey)
		if !ok {
			return fmt.Errorf("failed to migrate active %s key for chain %s", exported.SecondaryKey.SimpleString(), chain.Name)
		}
		if err := migrate(ctx, tss, multisig, snapshotter, chain.Name, key); err != nil {
			return err
		}

		key, ok = tss.GetNextKey(ctx, chain, exported.SecondaryKey)
		if !ok {
			return fmt.Errorf("failed to migrate next %s key in rotation for chain %s", exported.SecondaryKey.SimpleString(), chain.Name)
		}
		return migrate(ctx, tss, multisig, snapshotter, chain.Name, key)
	}
	return nil
}

func migrate(ctx sdk.Context, tss Keeper, multisig types.MultiSigKeeper, snapshotter types.Snapshotter, chain nexus.ChainName, keys ...exported.Key) error {
	for _, key := range keys {
		s, ok := snapshotter.GetSnapshot(ctx, key.SnapshotCounter)
		for _, validator := range s.Validators {
			valAddress := validator.GetSDKValidator().GetOperator()
			s.Participants[valAddress.String()] = snapshotexported.NewParticipant(valAddress, sdk.NewUint(uint64(validator.ShareCount)))
			s.BondedWeight = sdk.NewUintFromBigInt(s.TotalShareCount.BigInt())
		}
		if !ok {
			return fmt.Errorf("failed to migrate key %s for chain %s, no snapshot found", key.ID, chain)
		}
		info, found := tss.GetMultisigKeygenInfo(ctx, key.ID)
		keyInfo, ok := info.(*types.MultisigInfo)
		if !found || !ok {
			return fmt.Errorf("failed to migrate key %s for chain %s, key info not found", key.ID, chain)
		}
		pubkeys := make(map[string]multisigTypes.PublicKey)
		for _, infos := range keyInfo.Infos {

			val := infos.Participant.String()
			keys := infos.Data
			if len(keys) < 1 {
				return fmt.Errorf("failed to migrate key %s for chain %s, validator %s is participant without pubkey", key.ID, chain, infos.Participant.String())
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
	return nil
}
