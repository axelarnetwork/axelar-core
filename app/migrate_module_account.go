package app

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// MigratePreInitializedModuleAccounts migrates module accounts that were pre-initialized as BaseAccounts to ModuleAccounts,
// or creates new module accounts if they don't exist.
func MigratePreInitializedModuleAccounts(
	ctx sdk.Context,
	ak authkeeper.AccountKeeper,
	moduleAccountsToInitialize []string,
) error {
	for _, module := range moduleAccountsToInitialize {
		addr, perms := ak.GetModuleAddressAndPermissions(module)
		if addr == nil {
			return fmt.Errorf(
				"failed to get module address and permissions for module %s",
				module,
			)
		}

		acc := ak.GetAccount(ctx, addr)
		// The account has not been initialized yet
		if acc == nil {
			initModuleAccount(ctx, ak, module, perms...)
			continue
		}

		_, isModuleAccount := acc.(authtypes.ModuleAccountI)
		if isModuleAccount {
			ctx.Logger().Info(fmt.Sprintf(
				"account for module %s was correctly initialized, skipping",
				module,
			))
			continue
		}

		// Migrate from base account to module account
		baseAccount, ok := acc.(*authtypes.BaseAccount)
		if !ok {
			panic(fmt.Sprintf("account %s must be a base account", acc.GetAddress()))
		}

		newModuleAccount := authtypes.NewModuleAccount(
			baseAccount,
			module,
			perms...,
		)
		ak.SetModuleAccount(ctx, newModuleAccount)

		ctx.Logger().Info(fmt.Sprintf(
			"migrated %s module from base account %+v to module account %+v",
			module,
			baseAccount,
			newModuleAccount,
		))
	}

	return nil
}

// create a new module account
func initModuleAccount(ctx sdk.Context, ak authkeeper.AccountKeeper, moduleName string, perms ...string) {
	newModuleAccount := authtypes.NewEmptyModuleAccount(moduleName, perms...)
	maccI := (ak.NewAccount(ctx, newModuleAccount)).(authtypes.ModuleAccountI) // set the account number
	ak.SetModuleAccount(ctx, maccI)

	ctx.Logger().Info(fmt.Sprintf(
		"initialized %s module account %+v",
		moduleName,
		newModuleAccount,
	))
}
