package app

import (
	"context"

	store "cosmossdk.io/store/types"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	v2 "github.com/CosmWasm/wasmd/x/wasm/migrations/v2"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/axelarnetwork/utils/funcs"
)

func (app *AxelarApp) setUpgradeBehaviour(configurator module.Configurator, keepers *KeeperCache) {
	setupLegacyKeyTables(GetKeeper[paramskeeper.Keeper](keepers))

	baseAppLegacySS := GetKeeper[paramskeeper.Keeper](keepers).Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())

	upgradeKeeper := GetKeeper[upgradekeeper.Keeper](keepers)
	upgradeKeeper.SetUpgradeHandler(
		upgradeName(app.Version()),
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			app.Logger().Info("Running upgrade handler", "version", app.Version())
			// Migrate Tendermint consensus parameters from x/params module to a
			// dedicated x/consensus module.
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			consensusParams := GetKeeper[consensusparamkeeper.Keeper](keepers)
			funcs.MustNoErr(baseapp.MigrateParams(sdkCtx, baseAppLegacySS, consensusParams.ParamsStore))

			updatedVM, err := app.mm.RunMigrations(ctx, configurator, fromVM)
			if err != nil {
				return updatedVM, err
			}

			return updatedVM, err
		},
	)

	app.Logger().Info("Reading upgrade info from disk", "version", app.Version())
	upgradeInfo, err := upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == upgradeName(app.Version()) && !upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := store.StoreUpgrades{
			Added: []string{
				crisistypes.ModuleName,
				consensustypes.ModuleName,
			},
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

func setupLegacyKeyTables(k *paramskeeper.Keeper) {
	for _, subspace := range k.GetSubspaces() {
		var keyTable paramstypes.KeyTable
		switch subspace.Name() {
		case authtypes.ModuleName:
			keyTable = authtypes.ParamKeyTable()
		case banktypes.ModuleName:
			keyTable = banktypes.ParamKeyTable()
		case stakingtypes.ModuleName:
			keyTable = stakingtypes.ParamKeyTable()
		case minttypes.ModuleName:
			keyTable = minttypes.ParamKeyTable()
		case distrtypes.ModuleName:
			keyTable = distrtypes.ParamKeyTable()
		case slashingtypes.ModuleName:
			keyTable = slashingtypes.ParamKeyTable()
		case govtypes.ModuleName:
			keyTable = govv1.ParamKeyTable()
		case crisistypes.ModuleName:
			keyTable = crisistypes.ParamKeyTable()
		case ibctransfertypes.ModuleName:
			keyTable = ibctransfertypes.ParamKeyTable()
		case wasmtypes.ModuleName:
			keyTable = v2.ParamKeyTable()
		default:
			continue
		}

		if !subspace.HasKeyTable() {
			subspace.WithKeyTable(keyTable)
		}
	}
}
