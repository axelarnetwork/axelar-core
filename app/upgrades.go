package app

import (
	"context"

	store "cosmossdk.io/store/types"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	v2 "github.com/CosmWasm/wasmd/x/wasm/migrations/v2"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
)

// upgrade defines a coordinated chain upgrade: the name matched against the
// governance upgrade plan, an optional custom handler, and the store changes
// applied at the upgrade height.
//
// Every release line adds its own entry to the upgrades list below. Entries of
// shipped upgrades are immutable history: their store changes belong to the
// height they executed at and must not be carried over to later upgrades.
type chainUpgrade struct {
	name          string
	storeUpgrades store.StoreUpgrades
	// createHandler may be nil, in which case the upgrade runs the default
	// handler (mm.RunMigrations only)
	createHandler func(app *AxelarApp, configurator module.Configurator, keepers *KeeperCache) upgradetypes.UpgradeHandler
}

// upgrades lists all upgrades known to this binary, oldest first
var chainUpgrades = []chainUpgrade{
	{
		// cosmos-sdk v0.53 / ibc-go v10 / wasmd v0.60. mm.RunMigrations runs
		// the ibc core (6->8) and ibc transfer (5->6, DenomTrace->Denom)
		// state migrations.
		name: "v1.5",
		storeUpgrades: store.StoreUpgrades{
			Added: []string{
				authz.ModuleName,
			},
			Deleted: []string{
				// x/capability is removed entirely at ibc-go v10
				"capability",
				// x/crisis still exists in cosmos-sdk v0.53 but is deprecated
				// and removed upstream in the next SDK line, so dropping it now
				// avoids a forced follow-up store migration
				"crisis",
			},
		},
	},
}

func (app *AxelarApp) defaultUpgradeHandler(configurator module.Configurator, name string) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		app.Logger().Info("Running upgrade handler", "name", name, "version", app.Version())
		return app.mm.RunMigrations(ctx, configurator, fromVM)
	}
}

func (app *AxelarApp) setUpgradeBehaviour(configurator module.Configurator, keepers *KeeperCache) {
	setupLegacyKeyTables(GetKeeper[paramskeeper.Keeper](keepers))

	upgradeKeeper := GetKeeper[upgradekeeper.Keeper](keepers)

	currentName := upgradeName(app.Version())
	currentRegistered := false
	for _, u := range chainUpgrades {
		handler := app.defaultUpgradeHandler(configurator, u.name)
		if u.createHandler != nil {
			handler = u.createHandler(app, configurator, keepers)
		}
		upgradeKeeper.SetUpgradeHandler(u.name, handler)

		currentRegistered = currentRegistered || u.name == currentName
	}

	// safety net: a binary must always carry a handler for its own version so
	// an upgrade plan named after it can execute even if the upgrades list has
	// no entry (e.g. a release without store changes)
	if !currentRegistered {
		upgradeKeeper.SetUpgradeHandler(currentName, app.defaultUpgradeHandler(configurator, currentName))
	}

	upgradeInfo, err := upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	for _, u := range chainUpgrades {
		if upgradeInfo.Name == u.name {
			// configure store loader that checks if version == upgradeHeight and applies store upgrades
			storeUpgrades := u.storeUpgrades
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
			return
		}
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
		case distrtypes.ModuleName:
			keyTable = distrtypes.ParamKeyTable()
		case slashingtypes.ModuleName:
			keyTable = slashingtypes.ParamKeyTable()
		case govtypes.ModuleName:
			keyTable = govv1.ParamKeyTable()
		case ibcexported.ModuleName:
			// Register legacy key table for IBC client and connection params so migrations can read them
			keyTable = paramstypes.NewKeyTable().
				RegisterParamSet(&ibcclienttypes.Params{}).
				RegisterParamSet(&ibcconnectiontypes.Params{})
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
