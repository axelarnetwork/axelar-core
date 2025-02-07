package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// MigrateDistributionAccountPermission migrates the distribution module account to have the burner permission.
func MigrateDistributionAccountPermission(ctx sdk.Context, ak authkeeper.AccountKeeper) {
	acc, _ := ak.GetModuleAccountAndPermissions(ctx, distrtypes.ModuleName)
	baseAcc := authtypes.NewBaseAccount(acc.GetAddress(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())

	moduleAcc := authtypes.NewModuleAccount(
		baseAcc,
		distrtypes.ModuleName,
		InitModuleAccountPermissions()[distrtypes.ModuleName]...,
	)
	ak.SetModuleAccount(ctx, moduleAcc)
}
