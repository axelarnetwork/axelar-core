package keeper

import (
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/slices"
)

// Migrate7To8 returns the handler that performs in-place store migrations
func Migrate7To8(k *BaseKeeper, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		chains := slices.Filter(n.GetChains(ctx), func(chain exported.Chain) bool { return chain.Module == types.ModuleName })
		for _, chain := range chains {
			if err := migrateBurnerInfoForChain(ctx, k, chain, burnerAddrPrefixDeprecated); err != nil {
				return err
			}
			if err := migrateBurnerInfoForChain(ctx, k, chain, strings.ToLower(burnerAddrPrefixDeprecated)); err != nil {
				return err
			}

			k.Logger(ctx).Info(fmt.Sprintf("migrated all burner info keys for chain %s", chain.Name))

		}
		k.Logger(ctx).Info("burner info keys migration complete")
		return nil
	}
}

func migrateBurnerInfoForChain(ctx sdk.Context, k *BaseKeeper, chain exported.Chain, oldKey string) error {
	ck, err := k.forChain(ctx, chain.Name)
	if err != nil {
		return err
	}

	// migrate in batches so memory pressure doesn't become too large
	keysToDelete := make([][]byte, 0, 1000)
	for {
		iterBurnerAddr := ck.getStore(ctx).Iterator(utils.KeyFromStr(oldKey))

		if !iterBurnerAddr.Valid() {
			break
		}
		var burnerInfo types.BurnerInfo
		for ; iterBurnerAddr.Valid() && len(keysToDelete) < 1000; iterBurnerAddr.Next() {
			iterBurnerAddr.UnmarshalValue(&burnerInfo)
			ck.SetBurnerInfo(ctx, burnerInfo)
			keysToDelete = append(keysToDelete, iterBurnerAddr.Key())
		}

		if err := iterBurnerAddr.Close(); err != nil {
			return err
		}

		for _, burnerKey := range keysToDelete {
			ck.getStore(ctx).DeleteRaw(burnerKey)
		}

		keysToDelete = keysToDelete[:0]

		ck.Logger(ctx).Debug(fmt.Sprintf("migrated %d burner info keys for chain %s", len(keysToDelete), chain.String()))
	}
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
	bzToken, err := hex.DecodeString(types.Token)
	if err != nil {
		return err
	}

	bzBurnable, err := hex.DecodeString(types.Burnable)
	if err != nil {
		return err
	}

	subspace := ck.getSubspace()
	subspace.Set(ctx, types.KeyToken, bzToken)
	subspace.Set(ctx, types.KeyBurnable, bzBurnable)

	return nil
}
