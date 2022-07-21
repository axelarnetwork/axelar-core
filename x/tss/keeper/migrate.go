package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	multisigExported "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	snapshotexported "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
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
			return err
		}
		if len(keys) == 0 {
			tss.Logger(ctx).Debug(fmt.Sprintf("no old active %s keys found for chain %s", exported.SecondaryKey.SimpleString(), chain.Name))
		} else {
			for _, key := range keys {
				newKey, err := migrateKeyType(ctx, tss, snapshotter, chain.Name, key)
				if err != nil {
					return err
				}

				if _, ok := multisig.GetKey(ctx, newKey.ID); ok {
					return fmt.Errorf("key %s already set", newKey.ID)
				}
				multisig.SetKey(ctx, newKey)

				if err := multisig.AssignKey(ctx, chain.Name, multisigExported.KeyID(key.ID)); err != nil {
					return err
				}

				if err := multisig.RotateKey(ctx, chain.Name); err != nil {
					return err
				}
			}
		}
		key, ok := tss.GetCurrentKey(ctx, chain, exported.SecondaryKey)
		if !ok {
			tss.Logger(ctx).Debug(fmt.Sprintf("no active %s key found for chain %s", exported.SecondaryKey.SimpleString(), chain.Name))
		} else {
			newKey, err := migrateKeyType(ctx, tss, snapshotter, chain.Name, key)
			if err != nil {
				return err
			}

			if _, ok := multisig.GetKey(ctx, newKey.ID); ok {
				return fmt.Errorf("key %s already set", newKey.ID)
			}
			multisig.SetKey(ctx, newKey)

			if err := multisig.AssignKey(ctx, chain.Name, multisigExported.KeyID(key.ID)); err != nil {
				return err
			}
			if err := multisig.RotateKey(ctx, chain.Name); err != nil {
				return err
			}
		}

		key, ok = tss.GetNextKey(ctx, chain, exported.SecondaryKey)
		if !ok {
			tss.Logger(ctx).Debug(fmt.Sprintf("no next %s key found for chain %s", exported.SecondaryKey.SimpleString(), chain.Name))
		} else {
			newKey, err := migrateKeyType(ctx, tss, snapshotter, chain.Name, key)
			if err != nil {
				return err
			}

			if _, ok := multisig.GetKey(ctx, newKey.ID); ok {
				return fmt.Errorf("key %s already set", newKey.ID)
			}
			multisig.SetKey(ctx, newKey)

			if err := multisig.AssignKey(ctx, chain.Name, multisigExported.KeyID(key.ID)); err != nil {
				return err
			}

		}
	}
	return nil
}

func migrateKeyType(ctx sdk.Context, tss Keeper, snapshotter types.Snapshotter, chain nexus.ChainName, key exported.Key) (multisigTypes.Key, error) {
	s, ok := snapshotter.GetSnapshot(ctx, key.SnapshotCounter)
	if !ok {
		return multisigTypes.Key{}, fmt.Errorf("failed to migrateKeyType key %s for chain %s, no snapshot found", key.ID, chain)
	}
	if s.Participants == nil {
		s.Participants = make(map[string]snapshotexported.Participant)
	}
	for _, validator := range s.Validators {
		valAddress := validator.GetSDKValidator().GetOperator()
		s.Participants[valAddress.String()] = snapshotexported.NewParticipant(valAddress, sdk.NewUint(uint64(validator.ShareCount)))
	}
	s.BondedWeight = sdk.NewUintFromBigInt(s.TotalShareCount.BigInt())

	info, found := tss.GetMultisigKeygenInfo(ctx, key.ID)
	keyInfo, ok := info.(*types.MultisigInfo)
	if !found || !ok {
		return multisigTypes.Key{}, fmt.Errorf("failed to migrateKeyType key %s for chain %s, key info not found", key.ID, chain)
	}
	pubkeys := make(map[string]multisigExported.PublicKey)
	for _, infos := range keyInfo.Infos {

		val := infos.Participant.String()
		keys := infos.Data
		if len(keys) < 1 {
			return multisigTypes.Key{}, fmt.Errorf("failed to migrateKeyType key %s for chain %s, validator %s is participant without pubkey", key.ID, chain, infos.Participant.String())
		}
		pubkeys[val] = keys[0]
	}
	newKey := multisigTypes.Key{
		ID:               multisigExported.KeyID(key.ID),
		Snapshot:         s,
		PubKeys:          pubkeys,
		SigningThreshold: utils.NewThreshold(keyInfo.TargetNum, keyInfo.Count()),
	}

	if err := newKey.ValidateBasic(); err != nil {
		return multisigTypes.Key{}, err
	}

	tss.Logger(ctx).Debug(fmt.Sprintf("successfully migrated %s key %s for chain %s", exported.SecondaryKey, key.ID, chain))
	return newKey, nil
}
