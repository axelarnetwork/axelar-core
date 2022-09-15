package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/utils/slices"
)

// Migrate5To6 returns the handler that performs in-place store migrations
func Migrate5To6(k BaseKeeper, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		return nil
	}
}

// AlwaysMigrateBytecode migrates contracts bytecode for all evm chains (CRUCIAL, DO NOT DELETE AND ALWAYS REGISTER)
func AlwaysMigrateBytecode(k BaseKeeper, n types.Nexus, otherMigrations func(ctx sdk.Context) error) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		// migrate contracts bytecode (CRUCIAL AND DO NOT DELETE) for all evm chains
		for _, chain := range slices.Filter(n.GetChains(ctx), types.IsEVMChain) {
			ck := k.ForChain(chain.Name).(chainKeeper)
			if err := migrateContractsBytecode(ctx, ck); err != nil {
				return err
			}
		}

		return otherMigrations(ctx)
	}
}

func addTransferLimitParam(ctx sdk.Context, ck chainKeeper) error {
	subspace, ok := ck.getSubspace(ctx)
	if !ok {
		return fmt.Errorf("param subspace for chain %s should exist", ck.GetName())
	}

	subspace.Set(ctx, types.KeyTransferLimit, types.DefaultParams()[0].TransferLimit)

	return nil
}

// this function migrates the contracts bytecode to the latest for every existing
// EVM chain. It's crucial whenever contracts are changed between versions.
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
