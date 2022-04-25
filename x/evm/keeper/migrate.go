package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

const uaxlAsset = "uaxl"

// GetMigrationHandler returns the handler that performs in-place store migrations from v0.15 to v0.16. The
// migration includes:
// - delete uaxl token for all evm chains
// - delete uaxl token's burners for all evm chains
func GetMigrationHandler(k types.BaseKeeper, n types.Nexus) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		for _, chain := range n.GetChains(ctx) {
			if chain.Module != types.ModuleName {
				continue
			}

			ck := k.ForChain(chain.Name).(chainKeeper)
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
