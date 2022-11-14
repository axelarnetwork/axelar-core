package keeper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stoewer/go-strcase"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// Migrate6To7 returns the handler that performs in-place store migrations
func Migrate6To7(k *BaseKeeper, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		chains := slices.Filter(n.GetChains(ctx), func(chain exported.Chain) bool { return chain.Module == types.ModuleName })
		for _, chain := range chains {
			ck, err := k.forChain(ctx, chain.Name)
			if err != nil {
				return err
			}
			iterCmd := ck.getStore(ctx).IteratorNew(key.FromStr(commandPrefix))

			totalCmds := 0
			invalidCmds := 0
			for ; iterCmd.Valid(); iterCmd.Next() {
				totalCmds++
				var cmd types.Command
				iterCmd.UnmarshalValue(&cmd)
				if err := migrateCmdType(ctx, ck, key.FromBz(iterCmd.Key()), cmd); err != nil {
					invalidCmds++
					ck.Logger(ctx).Debug(fmt.Sprintf("chain %s: failed to migrate command type for command %s", chain.String(), funcs.Must(json.Marshal(cmd))))
					continue
				}
			}

			ck.Logger(ctx).Info(fmt.Sprintf("command type migration complete. Total migrated: %d, failed: %d", totalCmds, invalidCmds))

		}
		return nil
	}
}

func migrateCmdType(ctx sdk.Context, ck chainKeeper, key key.Key, cmd types.Command) error {
	cmdType := strcase.UpperSnakeCase(fmt.Sprintf("COMMAND_TYPE_%s", cmd.Command))
	typeEnum, ok := types.CommandType_value[cmdType]
	if !ok {
		return fmt.Errorf("command type %s is invalid at key %s", cmdType, key.String())
	}
	cmd.Type = types.CommandType(typeEnum)

	// keep data as is, in a future release need to clean up command state
	return ck.getStore(ctx).SetNewValidated(key, utils.NoValidation(&cmd))
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
