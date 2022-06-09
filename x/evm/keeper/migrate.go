package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.19 to v0.20. The
// migration includes:
// - migrate contracts bytecode (CRUCIAL AND DO NOT DELETE) for all evm chains
// - set BurnerCode for external token to nil
func GetMigrationHandler(k BaseKeeper, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		// migrate contracts bytecode (CRUCIAL AND DO NOT DELETE) for all evm chains
		for _, chain := range n.GetChains(ctx) {
			if chain.Module != types.ModuleName {
				continue
			}

			ck := k.ForChain(chain.Name).(chainKeeper)
			if err := migrateContractsBytecode(ctx, ck); err != nil {
				return err
			}
		}

		// set external token burner token to nil
		for _, chain := range n.GetChains(ctx) {
			if chain.Module != types.ModuleName {
				continue
			}

			ck := k.ForChain(chain.Name).(chainKeeper)
			if err := removeExternalTokenBurnerCode(ctx, ck); err != nil {
				return err
			}
		}

		return nil
	}
}

func removeExternalTokenBurnerCode(ctx sdk.Context, ck chainKeeper) error {
	for _, meta := range ck.getTokensMetadata(ctx) {
		if !meta.IsExternal {
			continue
		}

		meta.BurnerCode = nil
		ck.setTokenMetadata(ctx, meta)
	}

	return nil
}

// this function migrates the contracts bytecode to the latest for every existing
// EVM chain. It's crucial whenever contracts are changed between versions and
// DO NOT DELETE
func migrateContractsBytecode(ctx sdk.Context, ck chainKeeper) error {
	bzToken, err := hex.DecodeString(types.Token)
	if err != nil {
		return err
	}

	bzBurnable, err := hex.DecodeString(types.Burnable)
	if err != nil {
		return err
	}

	subspace, ok := ck.getSubspace(ctx)
	if !ok {
		return fmt.Errorf("param subspace for chain %s should exist", ck.GetName())
	}

	subspace.Set(ctx, types.KeyToken, bzToken)
	subspace.Set(ctx, types.KeyBurnable, bzBurnable)

	return nil
}
