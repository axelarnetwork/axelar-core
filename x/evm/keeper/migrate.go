package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/utils/slices"
)

// NoOpMigration performs no store migration. Replace it in the AlwaysMigrateBytecode
// registration (and bump ConsensusVersion) when a store migration is needed.
func NoOpMigration(_ sdk.Context) error {
	return nil
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
