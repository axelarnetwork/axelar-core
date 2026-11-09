package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
)

// Migrate2to3 returns the handler that performs in-place store migrations
func Migrate2to3(k Keeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		removeObsoleteAccessControlAccounts(ctx, k)
		return nil
	}
}

// removeObsoleteAccessControlAccounts deletes every account with the access control role other than the one of the
// current governance key. Updating the governance key through a governance proposal used to leave the role on the
// previous governance account.
func removeObsoleteAccessControlAccounts(ctx sdk.Context, k Keeper) {
	govKey, ok := k.GetGovernanceKey(ctx)
	if !ok {
		return
	}

	govAddr := govKey.Address().Bytes()
	for _, account := range k.getGovAccounts(ctx) {
		if account.Role != exported.ROLE_ACCESS_CONTROL || bytes.Equal(account.Address, govAddr) {
			continue
		}

		k.deleteGovAccount(ctx, account.Address)
		k.Logger(ctx).Info("removed obsolete access control account", "address", account.Address.String())
	}
}
