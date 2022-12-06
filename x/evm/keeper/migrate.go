package keeper

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/key"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/slices"
)

// Migrate8to9 returns the handler that performs in-place store migrations
func Migrate8to9(bk *BaseKeeper, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		for _, chain := range slices.Filter(n.GetChains(ctx), types.IsEVMChain) {
			ck := funcs.Must(bk.ForChain(ctx, chain.Name)).(chainKeeper)

			if err := migrateDeposits(ctx, ck, types.DepositStatus_Confirmed); err != nil {
				return err
			}

			if err := migrateDeposits(ctx, ck, types.DepositStatus_Burned); err != nil {
				return err
			}
		}

		return nil
	}
}

func migrateDeposits(ctx sdk.Context, ck chainKeeper, status types.DepositStatus) error {
	var prefix key.Key
	switch status {
	case types.DepositStatus_Confirmed:
		prefix = key.FromStr(confirmedDepositPrefixDeprecated)
	case types.DepositStatus_Burned:
		prefix = key.FromStr(burnedDepositPrefixDeprecated)
	}

	iter := ck.getStore(ctx).IteratorNew(prefix)
	defer utils.CloseLogError(iter, ck.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var deposit types.ERC20Deposit
		iter.UnmarshalValue(&deposit)

		transferEvents := getTransferEventsByTxIDAndAddress(ctx, ck, deposit.TxID, deposit.BurnerAddress)
		if len(transferEvents) == 0 {
			// Deposits from the time when we have not started doing event processing.
			// Their log indexes are no way to be retrieved anymore and therefore ignore.
			continue
		}

		rawKey := iter.Key()
		defer ck.getStore(ctx).DeleteRaw(rawKey)

		for _, event := range transferEvents {
			newDeposit := types.ERC20Deposit{
				TxID:             event.TxID,
				LogIndex:         event.Index,
				Amount:           event.GetTransfer().Amount,
				Asset:            deposit.Asset,
				DestinationChain: deposit.DestinationChain,
				BurnerAddress:    deposit.BurnerAddress,
			}

			ck.DeleteDeposit(ctx, newDeposit)
			ck.SetDeposit(ctx, newDeposit, status)
		}
	}

	return nil
}

func getTransferEventsByTxIDAndAddress(ctx sdk.Context, ck chainKeeper, txID types.Hash, address types.Address) (events []types.Event) {
	iter := sdk.KVStorePrefixIterator(ck.getStore(ctx).KVStore, eventPrefix.Append(utils.LowerCaseKey(txID.Hex())).AsKey())
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var event types.Event
		ck.cdc.MustUnmarshalLengthPrefixed(iter.Value(), &event)

		if event.GetTransfer() == nil {
			continue
		}

		if event.GetTransfer().To != address {
			continue
		}

		events = append(events, event)
	}

	return events
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
