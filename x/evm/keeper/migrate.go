package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// Migrate10to11 removes deprecated link-deposit protocol state from all EVM chains.
// This cleans up burnerAddr, confirmedDeposit, and burnedDeposit entries.
func Migrate10to11(bk *BaseKeeper, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		for _, chain := range slices.Filter(n.GetChains(ctx), types.IsEVMChain) {
			ck := funcs.Must(bk.ForChain(ctx, chain.Name)).(chainKeeper)
			deleted := deleteDeprecatedState(ctx, ck)
			ctx.Logger().Info(fmt.Sprintf("deleted %d deprecated link-deposit keys from %s chain store", deleted, chain.Name))
		}

		return nil
	}
}

// deleteDeprecatedState removes all entries with deprecated key prefixes from the chain store.
// These prefixes were used by the link-deposit protocol which has been removed.
func deleteDeprecatedState(ctx sdk.Context, ck chainKeeper) int {
	store := ck.getStore(ctx)
	totalDeleted := 0

	// Delete entries for each deprecated prefix:
	// - Static key prefixes (1, 2, 3) used by the current key system
	// - Legacy string prefixes ("confirmed_deposit", "burned_deposit") from older code
	deprecatedPrefixes := []key.Key{
		burnerAddrPrefix,
		confirmedDepositPrefix,
		burnedDepositPrefix,
		key.FromStr("confirmed_deposit"),
		key.FromStr("burned_deposit"),
	}

	for _, prefix := range deprecatedPrefixes {
		iter := store.IteratorNew(prefix)
		defer utils.CloseLogError(iter, ck.Logger(ctx))

		var keysToDelete [][]byte
		for ; iter.Valid(); iter.Next() {
			keysToDelete = append(keysToDelete, iter.Key())
		}

		for _, k := range keysToDelete {
			store.DeleteRaw(k)
		}

		totalDeleted += len(keysToDelete)
	}

	return totalDeleted
}

// AlwaysMigrateBytecode migrates contracts bytecode for all evm chains (CRUCIAL, DO NOT DELETE AND ALWAYS REGISTER)
func AlwaysMigrateBytecode(k *BaseKeeper, n types.Nexus, otherMigrations func(ctx sdk.Context) error) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		// migrate contracts bytecode (CRUCIAL AND DO NOT DELETE) for all evm chains
		for _, chain := range slices.Filter(n.GetChains(ctx), types.IsEVMChain) {
			ck, err := k.ForChain(ctx, chain.Name)
			if err != nil {
				return err
			}
			if err := migrateContractsBytecode(ctx, ck.(chainKeeper)); err != nil {
				return err
			}
		}

		return otherMigrations(ctx)
	}
}

// this function migrates the contracts bytecode to the latest for every existing
// EVM chain. It's crucial whenever contracts are changed between versions.
// DO NOT DELETE
func migrateContractsBytecode(ctx sdk.Context, ck chainKeeper) error {
	bzToken, err := utils.HexDecode(types.Token)
	if err != nil {
		return err
	}

	bzBurnable, err := utils.HexDecode(types.Burnable)
	if err != nil {
		return err
	}

	subspace := ck.getSubspace()
	subspace.Set(ctx, types.KeyToken, bzToken)
	subspace.Set(ctx, types.KeyBurnable, bzBurnable)

	return nil
}
