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

// Migrate8to9 returns the handler that performs in-place store migrations
func Migrate8to9(bk *BaseKeeper, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		for _, chain := range slices.Filter(n.GetChains(ctx), types.IsEVMChain) {
			ck := funcs.Must(bk.ForChain(ctx, chain.Name)).(chainKeeper)

			migrateDeposits(ctx, ck, types.DepositStatus_Confirmed)
			migrateDeposits(ctx, ck, types.DepositStatus_Burned)
		}

		return nil
	}
}

func migrateDeposits(ctx sdk.Context, ck chainKeeper, status types.DepositStatus) {
	var iteratedDepositCount, ignoredDepositCount uint64
	store := ck.getStore(ctx)
	var toDelete [][]byte

	var prefix key.Key
	switch status {
	case types.DepositStatus_Confirmed:
		prefix = key.FromStr(confirmedDepositPrefixDeprecated)
	case types.DepositStatus_Burned:
		prefix = key.FromStr(burnedDepositPrefixDeprecated)
	}

	iter := store.IteratorNew(prefix)
	defer utils.CloseLogError(iter, ck.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		iteratedDepositCount++

		var deposit types.ERC20Deposit
		iter.UnmarshalValue(&deposit)

		transferEvents := getTransferEventsByTxIDAndAddress(ctx, ck, deposit.TxID, deposit.BurnerAddress)
		if len(transferEvents) == 0 {
			// Deposits from the time when we had not started doing event processing.
			// Their log indexes cannot be retrieved anymore and are therefore ignored.
			ignoredDepositCount++
			continue
		}
		toDelete = append(toDelete, iter.Key())

		for _, event := range transferEvents {
			existingDeposit, existingDepositStatus, ok := ck.GetDeposit(ctx, event.TxID, event.Index)
			if ok && existingDepositStatus == status {
				continue
			}

			newDeposit := types.ERC20Deposit{
				TxID:             event.TxID,
				LogIndex:         event.Index,
				Amount:           event.GetTransfer().Amount,
				Asset:            deposit.Asset,
				DestinationChain: deposit.DestinationChain,
				BurnerAddress:    deposit.BurnerAddress,
			}

			if ok && existingDepositStatus != status {
				ck.Logger(ctx).Debug(fmt.Sprintf("deposit status changes from %s to %s", existingDepositStatus.String(), status.String()),
					"chain", ck.GetName(),
					"tx_id", event.TxID.Hex(),
					"log_index", event.Index,
					"burner_address", existingDeposit.BurnerAddress.Hex(),
				)
				ck.DeleteDeposit(ctx, newDeposit)
			}

			ck.SetDeposit(ctx, newDeposit, status)
		}
	}

	slices.ForEach(toDelete, store.DeleteRaw)

	ck.Logger(ctx).Debug(fmt.Sprintf("migrated %s deposits", status.String()),
		"chain", ck.GetName(),
		"iterated_deposit_count", iteratedDepositCount,
		"ignored_deposit_count", ignoredDepositCount,
	)
}

func getTransferEventsByTxIDAndAddress(ctx sdk.Context, ck chainKeeper, txID types.Hash, address types.Address) (events []types.Event) {
	iter := sdk.KVStorePrefixIterator(ck.getStore(ctx).KVStore, eventPrefix.Append(utils.LowerCaseKey(fmt.Sprintf("%s-", txID.Hex()))).AsKey())
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
