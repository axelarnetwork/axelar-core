package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/permission/exported"
)

// GetMigrationHandler returns a migration handler from v1 to v2
// The order of role enums has changed, so we have to migrate all government accounts.
// Enums are int32 underneath, so without this migration, the meaning of the value would change and open up the system to attacks
func GetMigrationHandler(k Keeper) func(sdk.Context) error {
	roleMapping := map[exported.Role]exported.Role{
		0: exported.ROLE_UNRESTRICTED,
		1: exported.ROLE_ACCESS_CONTROL,
		2: exported.ROLE_CHAIN_MANAGEMENT,
	}

	return func(ctx sdk.Context) error {
		accs := k.getGovAccounts(ctx)

		for _, acc := range accs {
			newRole, ok := roleMapping[acc.Role]
			if !ok {
				return fmt.Errorf("unrecognized role %d", acc.Role)
			}
			acc.Role = newRole
			k.setGovAccount(ctx, acc)
		}

		return nil
	}
}
