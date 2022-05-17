package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

const uaxlAsset = "uaxl"

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.17 to v0.18. The
// migration includes:
// - migrate contracts bytecode (CRUCIAL AND DO NOT DELETE) for all evm chains
// - delete pending chains from the base keeper
// - add VotingGracePeriod param for all evm chains
// - delete uaxl token for all evm chains
// - delete uaxl token's burners for all evm chains
// - migrate uaxl token's confirmed deposits to burnt for all evm chains
// - delete uaxl token's deployment commands for all evm chains
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

		deleteAllPendingChains(ctx, k)

		for _, chain := range n.GetChains(ctx) {
			if chain.Module != types.ModuleName {
				continue
			}
			ck := k.ForChain(chain.Name).(chainKeeper)

			if err := setVotingGracePeriod(ctx, ck); err != nil {
				return err
			}

			token, ok := ck.getTokenMetadataByAsset(ctx, uaxlAsset)
			if !ok {
				continue
			}

			deleteToken(ctx, ck, token)
			deleteTokenBurners(ctx, ck)
			migrateConfirmedDepositsToBurnt(ctx, ck)

			if err := deleteTokenDeploymentCommand(ctx, ck, token); err != nil {
				return err
			}
		}

		return nil
	}
}

func deleteAllPendingChains(ctx sdk.Context, k BaseKeeper) {
	iter := k.getBaseStore(ctx).Iterator(pendingChainKey)
	defer utils.CloseLogError(iter, k.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		k.getBaseStore(ctx).Delete(iter.GetKey())
	}
}

func setVotingGracePeriod(ctx sdk.Context, ck chainKeeper) error {
	subspace, ok := ck.getSubspace(ctx)
	if !ok {
		return fmt.Errorf("param subspace for chain %s should exist", ck.GetName())
	}

	subspace.Set(ctx, types.KeyVotingGracePeriod, types.DefaultParams()[0].VotingGracePeriod)
	return nil
}

func deleteToken(ctx sdk.Context, ck chainKeeper, token types.ERC20TokenMetadata) {
	// delete lookup by asset
	ck.getStore(ctx, ck.chainLowerKey).Delete(tokenMetadataByAssetPrefix.Append(utils.LowerCaseKey(token.Asset)))
	// delete lookup by symbol
	ck.getStore(ctx, ck.chainLowerKey).Delete(tokenMetadataBySymbolPrefix.Append(utils.LowerCaseKey(token.Details.Symbol)))
}

func deleteTokenDeploymentCommand(ctx sdk.Context, ck chainKeeper, token types.ERC20TokenMetadata) error {
	chainID, ok := ck.GetChainID(ctx)
	if !ok {
		return fmt.Errorf("chain ID not found for chain %s", ck.GetName())
	}

	tokenDeploymentCommandID := types.NewCommandID([]byte(token.Details.Symbol), chainID)
	if _, ok := ck.GetCommand(ctx, tokenDeploymentCommandID); !ok {
		return fmt.Errorf("token deployment command not found for token %s and chain %s", token.Details.Symbol, ck.GetName())
	}

	ck.getStore(ctx, ck.chainLowerKey).Delete(commandPrefix.AppendStr(tokenDeploymentCommandID.Hex()))

	return nil
}

func deleteTokenBurners(ctx sdk.Context, ck chainKeeper) {
	iter := ck.getStore(ctx, ck.chainLowerKey).Iterator(burnerAddrPrefix)
	defer utils.CloseLogError(iter, ck.Logger(ctx))

	for ; iter.Valid(); iter.Next() {
		var burner types.BurnerInfo
		iter.UnmarshalValue(&burner)

		if burner.Asset != uaxlAsset {
			continue
		}

		ck.getStore(ctx, ck.chainLowerKey).Delete(iter.GetKey())
	}
}

func migrateConfirmedDepositsToBurnt(ctx sdk.Context, ck chainKeeper) {
	for _, deposit := range ck.GetConfirmedDeposits(ctx) {
		if deposit.Asset != uaxlAsset {
			continue
		}

		ck.DeleteDeposit(ctx, deposit)
		ck.SetDeposit(ctx, deposit, types.DepositStatus_Burned)
	}
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
