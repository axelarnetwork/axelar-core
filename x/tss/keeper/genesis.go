package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// InitGenesis initializes the tss module's state from a given genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, snapshotter types.Snapshotter, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	rotationCountMap := make(map[string]map[exported.KeyRole]int64)
	for _, key := range genState.Keys {
		if _, ok := k.GetKey(ctx, key.ID); ok {
			panic(fmt.Errorf("key %s already set", key.ID))
		}

		k.setKey(ctx, key)

		if key.RotationCount > 0 {
			if _, ok := k.getKeyID(ctx, nexus.ChainName(key.Chain), key.RotationCount, key.Role); ok {
				panic(fmt.Errorf("key ID for chain %s, rotation count %d and role %s already set", key.Chain, key.RotationCount, key.Role.SimpleString()))
			}

			k.setKeyID(ctx, nexus.ChainName(key.Chain), key.RotationCount, key.Role, key.ID)
		}

		if key.Role != exported.ExternalKey {
			if _, ok := snapshotter.GetSnapshot(ctx, key.SnapshotCounter); !ok {
				panic(fmt.Errorf("snapshot %d for key %s not found", key.SnapshotCounter, key.ID))
			}

			k.setSnapshotCounterForKeyID(ctx, key.ID, key.SnapshotCounter)
		}

		if _, ok := rotationCountMap[key.Chain]; !ok {
			rotationCountMap[key.Chain] = make(map[exported.KeyRole]int64)
		}

		rotationCount, ok := rotationCountMap[key.Chain][key.Role]
		if !ok || key.RotationCount > rotationCount {
			rotationCountMap[key.Chain][key.Role] = key.RotationCount
		}
	}

	for chain, keyRoleToRotationCount := range rotationCountMap {
		for keyRole, rotationCount := range keyRoleToRotationCount {
			k.setRotationCount(ctx, nexus.ChainName(chain), keyRole, rotationCount)
		}
	}

	for _, keyRecoveryInfo := range genState.KeyRecoveryInfos {
		if key, ok := k.GetKey(ctx, keyRecoveryInfo.KeyID); ok && key.Type != exported.Threshold {
			panic(fmt.Errorf("no threshold key %s found", keyRecoveryInfo.KeyID))
		}

		if _, ok := k.getKeyRecoveryInfo(ctx, keyRecoveryInfo.KeyID); ok {
			panic(fmt.Errorf("key recovery info %s already set", keyRecoveryInfo.KeyID))
		}

		k.setKeyRecoveryInfo(ctx, keyRecoveryInfo)
	}

	for _, multisigInfo := range genState.MultisigInfos {
		if _, ok := k.GetMultisigKeygenInfo(ctx, exported.KeyID(multisigInfo.ID)); ok {
			panic(fmt.Errorf("multisig keygen info %s already set", multisigInfo.ID))
		}

		k.SetMultisigKeygenInfo(ctx, multisigInfo)
	}

	for _, externalKeys := range genState.ExternalKeys {
		for _, externalKeyID := range externalKeys.KeyIDs {
			if _, ok := k.GetKey(ctx, externalKeyID); !ok {
				panic(fmt.Errorf("no key %s found", externalKeyID))
			}
		}

		if _, ok := k.getExternalKeyIDs(ctx, externalKeys.Chain); ok {
			panic(fmt.Errorf("external key IDs for chain %s already set", externalKeys.Chain))
		}

		k.setExternalKeys(ctx, externalKeys)
	}

	for _, signature := range genState.Signatures {
		if signature.SigStatus != exported.SigStatus_Signed {
			panic(fmt.Errorf("signature %s is not completed", signature.SigID))
		}

		if signature.GetSingleSig() == nil && signature.GetMultiSig() == nil {
			panic(fmt.Errorf("signature %s is not completed", signature.SigID))
		}

		if _, ok := k.getSig(ctx, signature.SigID); ok {
			panic(fmt.Errorf("signature %s already set", signature.SigID))
		}

		k.SetSig(ctx, signature)
	}

	for _, validatorStatus := range genState.ValidatorStatuses {
		if _, ok := k.getValidatorStatus(ctx, validatorStatus.Validator); ok {
			panic(fmt.Errorf("validator status %s already set", validatorStatus.Validator.String()))
		}

		k.setValidatorStatus(ctx, validatorStatus)
	}
}

// ExportGenesis returns the tss module's genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	keys := k.getKeys(ctx)

	return types.NewGenesisState(
		k.GetParams(ctx),
		k.getKeyRecoveryInfos(ctx),
		keys,
		k.getCompletedMultisigKeygenInfos(ctx),
		k.getAllExternalKeys(ctx),
		k.getSignedSigs(ctx),
		k.getValidatorStatuses(ctx),
	)
}
