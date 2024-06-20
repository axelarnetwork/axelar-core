package app

import (
	"encoding/json"
	"fmt"
	"io"
	stdlog "log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmclient "github.com/CosmWasm/wasmd/x/wasm/client"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authAnte "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authrest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v4/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v4/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v4/modules/core/02-client/client"
	porttypes "github.com/cosmos/ibc-go/v4/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v4/modules/core/24-host"
	ibcante "github.com/cosmos/ibc-go/v4/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v4/modules/core/keeper"
	"github.com/gorilla/mux"
	ibchooks "github.com/osmosis-labs/osmosis/x/ibc-hooks"
	ibchookskeeper "github.com/osmosis-labs/osmosis/x/ibc-hooks/keeper"
	ibchookstypes "github.com/osmosis-labs/osmosis/x/ibc-hooks/types"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	axelarParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/x/ante"
	"github.com/axelarnetwork/axelar-core/x/auxiliary"
	auxiliarytypes "github.com/axelarnetwork/axelar-core/x/auxiliary/types"
	"github.com/axelarnetwork/axelar-core/x/axelarnet"
	axelarnetclient "github.com/axelarnetwork/axelar-core/x/axelarnet/client"
	axelarnetKeeper "github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	axelarnetTypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	axelarbankkeeper "github.com/axelarnetwork/axelar-core/x/bank/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/multisig"
	multisigKeeper "github.com/axelarnetwork/axelar-core/x/multisig/keeper"
	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/nexus"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/permission"
	permissionKeeper "github.com/axelarnetwork/axelar-core/x/permission/keeper"
	permissionTypes "github.com/axelarnetwork/axelar-core/x/permission/types"
	"github.com/axelarnetwork/axelar-core/x/reward"
	rewardKeeper "github.com/axelarnetwork/axelar-core/x/reward/keeper"
	rewardTypes "github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot"
	snapKeeper "github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	"github.com/axelarnetwork/axelar-core/x/tss"
	tssKeeper "github.com/axelarnetwork/axelar-core/x/tss/keeper"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/vote"
	voteKeeper "github.com/axelarnetwork/axelar-core/x/vote/keeper"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/funcs"

	// Override with generated statik docs
	_ "github.com/axelarnetwork/axelar-core/client/docs/statik"
)

// Name is the name of the application
const Name = "axelar"

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// WasmEnabled indicates whether wasm module is added to the app.
	// "true" setting means it will be, otherwise it won't.
	// This is configured during the build.
	WasmEnabled = ""

	// IBCWasmHooksEnabled indicates whether wasm hooks for ibc are enabled.
	// "true" setting means it will be, otherwise it won't.
	// When disabled, cosmwasm contracts cannot be called via IBC.
	// This is configured during the build.
	IBCWasmHooksEnabled = ""

	// WasmCapabilities specifies the capabilities of the wasm vm
	// capabilities are detailed here: https://github.com/CosmWasm/cosmwasm/blob/main/docs/CAPABILITIES-BUILT-IN.md
	WasmCapabilities = ""

	// MaxWasmSize specifies the maximum wasm code size (in bytes) that can be uploaded. wasmd's setting is used by default
	// https://github.com/CosmWasm/wasmd/blob/main/README.md#compile-time-parameters
	MaxWasmSize = ""
)

var (
	_ servertypes.Application = (*AxelarApp)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		stdlog.Println("Failed to get home dir %2", err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)

	if !IsWasmEnabled() && IsIBCWasmHooksEnabled() {
		panic("ibc wasm hooks should only be enabled when wasm is enabled")
	}
}

// AxelarApp defines the axelar Cosmos app that runs all modules
type AxelarApp struct {
	*bam.BaseApp
	// Keys and Keepers are necessary for the app to interact with the cosmos-sdk and to be able to test the app in isolation without mocks
	Keepers *KeeperCache
	Keys    map[string]*sdk.KVStoreKey

	appCodec codec.Codec

	interfaceRegistry types.InterfaceRegistry

	mm *module.Manager
}

// NewAxelarApp is a constructor function for axelar
func NewAxelarApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	wasmDir string,
	invCheckPeriod uint,
	encodingConfig axelarParams.EncodingConfig,
	appOpts servertypes.AppOptions,
	wasmOpts []wasm.Option,
	baseAppOptions ...func(*bam.BaseApp),
) *AxelarApp {

	keys := CreateStoreKeys()
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	keepers := NewKeeperCache()
	SetKeeper(keepers, initParamsKeeper(encodingConfig, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey]))

	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := initBaseApp(db, traceStore, encodingConfig, keepers, baseAppOptions, logger)

	appCodec := encodingConfig.Codec
	moduleAccountPermissions := InitModuleAccountPermissions()

	// set up predefined keepers
	SetKeeper(keepers, initAccountKeeper(appCodec, keys, keepers, moduleAccountPermissions))
	SetKeeper(keepers, initBankKeeper(appCodec, keys, keepers, moduleAccountPermissions))
	SetKeeper(keepers, initStakingKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initMintKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initDistributionKeeper(appCodec, keys, keepers, moduleAccountPermissions))
	SetKeeper(keepers, initSlashingKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initCrisisKeeper(keepers, invCheckPeriod))
	SetKeeper(keepers, initUpgradeKeeper(appCodec, keys, skipUpgradeHeights, homePath, bApp))
	SetKeeper(keepers, initEvidenceKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initFeegrantKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initCapabilityKeeper(appCodec, keys, memKeys))
	SetKeeper(keepers, initIBCKeeper(appCodec, keys, keepers))

	// set up custom axelar keepers
	SetKeeper(keepers, initAxelarnetKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initEvmKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initNexusKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initRewardKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initMultisigKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initTssKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initSnapshotKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initVoteKeeper(appCodec, keys, keepers))
	SetKeeper(keepers, initPermissionKeeper(appCodec, keys, keepers))

	// set up ibc/wasm keepers
	wasmHooks := InitWasmHooks(keys)
	ics4Wrapper := InitICS4Wrapper(keepers, wasmHooks)
	SetKeeper(keepers, initIBCTransferKeeper(appCodec, keys, keepers, ics4Wrapper))

	SetKeeper(keepers, initAxelarIBCKeeper(keepers))

	if IsWasmEnabled() {
		if wasmDir == "" {
			dbDir := cast.ToString(appOpts.Get("db_dir"))
			wasmDir = filepath.Join(homePath, dbDir, "wasm")
		}

		wasmPath, err := filepath.Abs(wasmDir)
		if err != nil {
			panic(fmt.Sprintf("failed to resolve absolute path for new wasm dir %s: %v", wasmDir, err))
		}

		// Migrate wasm dir from old path to new path
		// TODO: Remove this once nodes have migrated
		oldWasmDir := filepath.Join(homePath, "wasm")
		funcs.MustNoErr(migrateWasmDir(oldWasmDir, wasmPath))

		SetKeeper(keepers, initWasmKeeper(encodingConfig, keys, keepers, bApp, appOpts, wasmOpts, wasmPath))
		SetKeeper(keepers, initWasmContractKeeper(keepers))

		// set the contract keeper for the Ics20WasmHooks
		if wasmHooks != nil {
			wasmHooks.ContractKeeper = GetKeeper[wasmkeeper.PermissionedKeeper](keepers)
		}
	}

	// set up governance keeper last when it has access to all other keepers to set up governance routes
	SetKeeper(keepers, initGovernanceKeeper(appCodec, keys, keepers))

	// seal capability keeper after all keepers are set to be certain that all capabilities have been registered
	GetKeeper[capabilitykeeper.Keeper](keepers).Seal()

	// set routers
	GetKeeper[nexusKeeper.Keeper](keepers).SetMessageRouter(initMessageRouter(keepers))
	GetKeeper[ibckeeper.Keeper](keepers).SetRouter(initIBCRouter(keepers, initIBCMiddleware(keepers, ics4Wrapper)))

	// register the staking hooks
	GetKeeper[stakingkeeper.Keeper](keepers).SetHooks(
		stakingtypes.NewMultiStakingHooks(
			GetKeeper[distrkeeper.Keeper](keepers).Hooks(),
			GetKeeper[slashingkeeper.Keeper](keepers).Hooks(),
		),
	)

	/****  Module Options ****/

	appModules := initAppModules(
		keepers,
		bApp,
		encodingConfig,
		appOpts,
		axelarnet.NewAppModule(
			*GetKeeper[axelarnetKeeper.IBCKeeper](keepers),
			GetKeeper[nexusKeeper.Keeper](keepers),
			axelarbankkeeper.NewBankKeeper(GetKeeper[bankkeeper.BaseKeeper](keepers)),
			GetKeeper[authkeeper.AccountKeeper](keepers),
			logger,
		),
	)

	mm := module.NewManager(appModules...)
	mm.SetOrderMigrations(orderMigrations()...)
	mm.SetOrderBeginBlockers(orderBeginBlockers()...)
	mm.SetOrderEndBlockers(orderEndBlockers()...)
	mm.SetOrderInitGenesis(orderModulesForGenesis()...)

	mm.RegisterInvariants(GetKeeper[crisiskeeper.Keeper](keepers))

	// register all module routes and module queriers
	mm.RegisterRoutes(bApp.Router(), bApp.QueryRouter(), encodingConfig.Amino)
	configurator := module.NewConfigurator(appCodec, bApp.MsgServiceRouter(), bApp.GRPCQueryRouter())
	mm.RegisterServices(configurator)

	var app = &AxelarApp{
		BaseApp:           bApp,
		appCodec:          appCodec,
		interfaceRegistry: encodingConfig.InterfaceRegistry,
		Keepers:           keepers,
		Keys:              keys,
		mm:                mm,
	}

	app.setUpgradeBehaviour(configurator, keepers)

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// The initChainer handles translating the genesis.json file into initial state for the network
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	app.SetAnteHandler(initAnteHandlers(encodingConfig, keys, keepers, appOpts))

	// Register wasm snapshot extension for state-sync compatibility
	// MUST be done before loading the version
	app.registerWasmSnapshotExtension(keepers)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		if IsWasmEnabled() {
			ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

			// Initialize pinned codes in wasmvm as they are not persisted there
			if err := GetKeeper[wasm.Keeper](keepers).InitializePinnedCodes(ctx); err != nil {
				tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
			}
		}
	}

	/* ==== at this point all stores are fully loaded ==== */

	// we need to ensure that all chain subspaces are loaded at start-up to prevent unexpected consensus failures
	// when the params keeper is used outside the evm module's context
	GetKeeper[evmKeeper.BaseKeeper](keepers).InitChains(app.NewContext(true, tmproto.Header{}))

	return app
}

func InitICS4Wrapper(keepers *KeeperCache, wasmHooks *ibchooks.WasmHooks) ibchooks.ICS4Middleware {
	// ICS4Wrapper deals with sending IBC packets. These need to get rate limited when appropriate,
	// so we wrap the channel keeper (which implements the ICS4Wrapper interface) with a rate limiter.
	ics4Wrapper := axelarnet.NewRateLimitedICS4Wrapper(
		GetKeeper[ibckeeper.Keeper](keepers).ChannelKeeper,
		axelarnet.NewRateLimiter(GetKeeper[axelarnetKeeper.Keeper](keepers), GetKeeper[nexusKeeper.Keeper](keepers)),
		GetKeeper[axelarnetKeeper.Keeper](keepers),
	)
	// create a middleware to integrate wasm hooks into the ibc pipeline
	if wasmHooks != nil {
		return ibchooks.NewICS4Middleware(ics4Wrapper, wasmHooks)
	} else {
		// we need to erase the type of the wasm hooks when it is nil so the middleware's type casts do not succeed.
		// Otherwise, it will try to call an interface function on wasmHooks and create a nil pointer panic
		return ibchooks.NewICS4Middleware(ics4Wrapper, nil)
	}
}

func initIBCMiddleware(keepers *KeeperCache, ics4Middleware ibchooks.ICS4Middleware) ibchooks.IBCMiddleware {
	// IBCModule deals with received IBC packets. These need to get rate limited when appropriate,
	// so we wrap the transfer module's IBCModule with a rate limiter.
	ibcModule := axelarnet.NewAxelarnetIBCModule(
		transfer.NewIBCModule(*GetKeeper[ibctransferkeeper.Keeper](keepers)),
		*GetKeeper[axelarnetKeeper.IBCKeeper](keepers),
		axelarnet.NewRateLimiter(GetKeeper[axelarnetKeeper.Keeper](keepers), GetKeeper[nexusKeeper.Keeper](keepers)),
		GetKeeper[nexusKeeper.Keeper](keepers),
		axelarbankkeeper.NewBankKeeper(GetKeeper[bankkeeper.BaseKeeper](keepers)),
	)

	// By merging the middlewares the receiving IBC Module has access to all registered hooks in the ICS4Middleware
	return ibchooks.NewIBCMiddleware(ibcModule, &ics4Middleware)
}

func InitWasmHooks(keys map[string]*sdk.KVStoreKey) *ibchooks.WasmHooks {
	if !(IsWasmEnabled() && IsIBCWasmHooksEnabled()) {
		return nil
	}

	// Configure the IBC hooks keeper to make wasm calls via IBC transfer memo
	ibcHooksKeeper := ibchookskeeper.NewKeeper(keys[ibchookstypes.StoreKey])

	// The contract keeper needs to be set later
	var wasmHooks = ibchooks.NewWasmHooks(&ibcHooksKeeper, nil, sdk.GetConfig().GetBech32AccountAddrPrefix())
	return &wasmHooks
}

func initIBCRouter(keepers *KeeperCache, axelarnetModule porttypes.IBCModule) *porttypes.Router {
	// Finalize the IBC router
	// Create static IBC router, add axelarnet module as the IBC transfer route, and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, axelarnetModule)
	if IsWasmEnabled() {
		// Create wasm ibc stack
		var wasmStack porttypes.IBCModule = wasm.NewIBCHandler(
			GetKeeper[wasm.Keeper](keepers),
			GetKeeper[ibckeeper.Keeper](keepers).ChannelKeeper,
			GetKeeper[ibckeeper.Keeper](keepers).ChannelKeeper,
		)
		ibcRouter.AddRoute(wasm.ModuleName, wasmStack)
	}
	return ibcRouter
}

func initMessageRouter(keepers *KeeperCache) nexusTypes.MessageRouter {
	messageRouter := nexusTypes.NewMessageRouter().
		AddRoute(evmTypes.ModuleName, evmKeeper.NewMessageRoute()).
		AddRoute(axelarnetTypes.ModuleName, axelarnetKeeper.NewMessageRoute(
			*GetKeeper[axelarnetKeeper.Keeper](keepers),
			GetKeeper[axelarnetKeeper.IBCKeeper](keepers),
			GetKeeper[feegrantkeeper.Keeper](keepers),
			axelarbankkeeper.NewBankKeeper(GetKeeper[bankkeeper.BaseKeeper](keepers)),
			GetKeeper[nexusKeeper.Keeper](keepers),
			GetKeeper[authkeeper.AccountKeeper](keepers),
		))

	if IsWasmEnabled() {
		messageRouter.AddRoute(wasm.ModuleName, nexusKeeper.NewMessageRoute(
			GetKeeper[nexusKeeper.Keeper](keepers),
			GetKeeper[authkeeper.AccountKeeper](keepers),
			GetKeeper[wasmkeeper.PermissionedKeeper](keepers),
		))
	}
	return messageRouter
}

func migrateWasmDir(oldWasmDir, newWasmDir string) error {
	// If the new wasm dir exists, there's nothing to do
	if _, err := os.Stat(newWasmDir); err == nil {
		return nil
	}

	// If the old wasm dir doesn't exist, there's nothing to do
	if _, err := os.Stat(oldWasmDir); err != nil && os.IsNotExist(err) {
		return nil
	}

	// Move the wasm dir from old path to new path
	if err := os.Rename(oldWasmDir, newWasmDir); err != nil {
		return fmt.Errorf("failed to move wasm directory from %s to %s: %v", oldWasmDir, newWasmDir, err)
	}

	return nil
}

func (app *AxelarApp) registerWasmSnapshotExtension(keepers *KeeperCache) {
	// Register wasm snapshot extension to enable state-sync compatibility for wasm.
	// MUST be done before loading the version
	// Requires the snapshot store to be created and registered as a BaseAppOption
	if IsWasmEnabled() {
		if manager := app.SnapshotManager(); manager != nil {
			err := manager.RegisterExtensions(
				wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), GetKeeper[wasm.Keeper](keepers)),
			)
			if err != nil {
				panic(fmt.Errorf("failed to register snapshot extension: %s", err))
			}
		}
	}
}

func (app *AxelarApp) setUpgradeBehaviour(configurator module.Configurator, keepers *KeeperCache) {
	upgradeKeeper := GetKeeper[upgradekeeper.Keeper](keepers)
	upgradeKeeper.SetUpgradeHandler(
		upgradeName(app.Version()),
		func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			updatedVM, err := app.mm.RunMigrations(ctx, configurator, fromVM)
			if err != nil {
				return updatedVM, err
			}

			// TODO: remove after v35 upgrade
			// Override wasm module default params
			if upgradeName(app.Version()) == "v0.35" && IsWasmEnabled() {
				GetKeeper[wasm.Keeper](keepers).SetParams(ctx, wasmtypes.Params{
					CodeUploadAccess:             wasmtypes.AllowNobody,
					InstantiateDefaultPermission: wasmtypes.AccessTypeNobody,
				})
			}

			return updatedVM, err
		},
	)

	upgradeInfo, err := upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == upgradeName(app.Version()) && !upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := store.StoreUpgrades{}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

func initBaseApp(db dbm.DB, traceStore io.Writer, encodingConfig axelarParams.EncodingConfig, keepers *KeeperCache, baseAppOptions []func(*bam.BaseApp), logger log.Logger) *bam.BaseApp {
	bApp := bam.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	bApp.SetParamStore(keepers.getSubspace(bam.Paramspace))
	return bApp
}

func initAppModules(keepers *KeeperCache, bApp *bam.BaseApp, encodingConfig axelarParams.EncodingConfig, appOpts servertypes.AppOptions, axelarnetModule axelarnet.AppModule) []module.AppModule {
	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	appCodec := encodingConfig.Codec

	appModules := []module.AppModule{
		genutil.NewAppModule(GetKeeper[authkeeper.AccountKeeper](keepers), GetKeeper[stakingkeeper.Keeper](keepers), bApp.DeliverTx, encodingConfig.TxConfig),
		auth.NewAppModule(appCodec, *GetKeeper[authkeeper.AccountKeeper](keepers), nil),
		vesting.NewAppModule(*GetKeeper[authkeeper.AccountKeeper](keepers), GetKeeper[bankkeeper.BaseKeeper](keepers)),

		// bank module accepts a reference to the base keeper, but panics when RegisterService is called on a reference, so we need to dereference it
		bank.NewAppModule(appCodec, *GetKeeper[bankkeeper.BaseKeeper](keepers), GetKeeper[authkeeper.AccountKeeper](keepers)),
		crisis.NewAppModule(GetKeeper[crisiskeeper.Keeper](keepers), skipGenesisInvariants),
		gov.NewAppModule(appCodec, *GetKeeper[govkeeper.Keeper](keepers), GetKeeper[authkeeper.AccountKeeper](keepers), GetKeeper[bankkeeper.BaseKeeper](keepers)),
		mint.NewAppModule(appCodec, *GetKeeper[mintkeeper.Keeper](keepers), GetKeeper[authkeeper.AccountKeeper](keepers)),
		slashing.NewAppModule(appCodec, *GetKeeper[slashingkeeper.Keeper](keepers), GetKeeper[authkeeper.AccountKeeper](keepers), GetKeeper[bankkeeper.BaseKeeper](keepers), GetKeeper[stakingkeeper.Keeper](keepers)),
		distr.NewAppModule(appCodec, *GetKeeper[distrkeeper.Keeper](keepers), GetKeeper[authkeeper.AccountKeeper](keepers), GetKeeper[bankkeeper.BaseKeeper](keepers), GetKeeper[stakingkeeper.Keeper](keepers)),
		staking.NewAppModule(appCodec, *GetKeeper[stakingkeeper.Keeper](keepers), GetKeeper[authkeeper.AccountKeeper](keepers), GetKeeper[bankkeeper.BaseKeeper](keepers)),
		upgrade.NewAppModule(*GetKeeper[upgradekeeper.Keeper](keepers)),
		evidence.NewAppModule(*GetKeeper[evidencekeeper.Keeper](keepers)),
		params.NewAppModule(*GetKeeper[paramskeeper.Keeper](keepers)),
		capability.NewAppModule(appCodec, *GetKeeper[capabilitykeeper.Keeper](keepers)),
	}

	// wasm module needs to be added in a specific order, so we cannot just append it at the end
	if IsWasmEnabled() {
		appModules = append(
			appModules,
			wasm.NewAppModule(
				appCodec,
				GetKeeper[wasm.Keeper](keepers),
				GetKeeper[stakingkeeper.Keeper](keepers),
				GetKeeper[authkeeper.AccountKeeper](keepers),
				GetKeeper[bankkeeper.BaseKeeper](keepers),
			),
		)
	}

	if IsIBCWasmHooksEnabled() {
		appModules = append(appModules, ibchooks.NewAppModule(GetKeeper[authkeeper.AccountKeeper](keepers)))
	}

	appModules = append(appModules,
		evidence.NewAppModule(*GetKeeper[evidencekeeper.Keeper](keepers)),
		ibc.NewAppModule(GetKeeper[ibckeeper.Keeper](keepers)),
		transfer.NewAppModule(*GetKeeper[ibctransferkeeper.Keeper](keepers)),
		feegrantmodule.NewAppModule(
			appCodec,
			GetKeeper[authkeeper.AccountKeeper](keepers),
			GetKeeper[bankkeeper.BaseKeeper](keepers),
			*GetKeeper[feegrantkeeper.Keeper](keepers),
			encodingConfig.InterfaceRegistry,
		),
		snapshot.NewAppModule(*GetKeeper[snapKeeper.Keeper](keepers)),
		multisig.NewAppModule(
			*GetKeeper[multisigKeeper.Keeper](keepers),
			GetKeeper[stakingkeeper.Keeper](keepers),
			GetKeeper[slashingkeeper.Keeper](keepers),
			GetKeeper[snapKeeper.Keeper](keepers),
			GetKeeper[rewardKeeper.Keeper](keepers),
			GetKeeper[nexusKeeper.Keeper](keepers),
		),
		tss.NewAppModule(
			*GetKeeper[tssKeeper.Keeper](keepers),
			GetKeeper[snapKeeper.Keeper](keepers),
			GetKeeper[nexusKeeper.Keeper](keepers),
			GetKeeper[stakingkeeper.Keeper](keepers),
			GetKeeper[multisigKeeper.Keeper](keepers),
		),
		vote.NewAppModule(*GetKeeper[voteKeeper.Keeper](keepers)),
		nexus.NewAppModule(
			*GetKeeper[nexusKeeper.Keeper](keepers),
			GetKeeper[snapKeeper.Keeper](keepers),
			GetKeeper[slashingkeeper.Keeper](keepers),
			GetKeeper[stakingkeeper.Keeper](keepers),
			GetKeeper[axelarnetKeeper.Keeper](keepers),
			GetKeeper[rewardKeeper.Keeper](keepers),
		),
		evm.NewAppModule(
			GetKeeper[evmKeeper.BaseKeeper](keepers),
			GetKeeper[voteKeeper.Keeper](keepers),
			GetKeeper[nexusKeeper.Keeper](keepers),
			GetKeeper[snapKeeper.Keeper](keepers),
			GetKeeper[stakingkeeper.Keeper](keepers),
			GetKeeper[slashingkeeper.Keeper](keepers),
			GetKeeper[multisigKeeper.Keeper](keepers),
		),
		axelarnetModule,
		reward.NewAppModule(
			*GetKeeper[rewardKeeper.Keeper](keepers),
			GetKeeper[nexusKeeper.Keeper](keepers),
			GetKeeper[mintkeeper.Keeper](keepers),
			GetKeeper[stakingkeeper.Keeper](keepers),
			GetKeeper[slashingkeeper.Keeper](keepers),
			GetKeeper[multisigKeeper.Keeper](keepers),
			GetKeeper[snapKeeper.Keeper](keepers),
			GetKeeper[bankkeeper.BaseKeeper](keepers),
			bApp.MsgServiceRouter(),
			bApp.Router(),
		),
		permission.NewAppModule(*GetKeeper[permissionKeeper.Keeper](keepers)),
		auxiliary.NewAppModule(encodingConfig.Codec, bApp.MsgServiceRouter()),
	)
	return appModules
}

func mustReadWasmConfig(appOpts servertypes.AppOptions) wasmtypes.WasmConfig {
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}
	return wasmConfig
}

func initAnteHandlers(encodingConfig axelarParams.EncodingConfig, keys map[string]*sdk.KVStoreKey, keepers *KeeperCache, appOpts servertypes.AppOptions) sdk.AnteHandler {
	// The baseAnteHandler handles signature verification and transaction pre-processing
	baseAnteHandler, err := authAnte.NewAnteHandler(
		authAnte.HandlerOptions{
			AccountKeeper:   GetKeeper[authkeeper.AccountKeeper](keepers),
			BankKeeper:      GetKeeper[bankkeeper.BaseKeeper](keepers),
			SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
			FeegrantKeeper:  GetKeeper[feegrantkeeper.Keeper](keepers),
			SigGasConsumer:  authAnte.DefaultSigVerificationGasConsumer,
		},
	)
	if err != nil {
		panic(err)
	}

	customAnteDecorators := InitCustomAnteDecorators(encodingConfig, keys, keepers, appOpts)
	anteDecorators := append([]sdk.AnteDecorator{ante.NewAnteHandlerDecorator(baseAnteHandler)}, customAnteDecorators...)

	return sdk.ChainAnteDecorators(anteDecorators...)
}

func InitCustomAnteDecorators(
	encodingConfig axelarParams.EncodingConfig,
	keys map[string]*sdk.KVStoreKey,
	keepers *KeeperCache,
	appOpts servertypes.AppOptions,
) []sdk.AnteDecorator {
	var anteDecorators []sdk.AnteDecorator

	// unwrap batch messages, must be done before any other custom decorators
	anteDecorators = append(anteDecorators, ante.NewBatchDecorator(encodingConfig.Codec))

	// enforce wasm limits earlier in the ante handler chain
	if IsWasmEnabled() {
		wasmConfig := mustReadWasmConfig(appOpts)
		wasmAnteDecorators := []sdk.AnteDecorator{
			ante.NewLimitSimulationGasDecorator(wasmConfig.SimulationGasLimit),
			wasmkeeper.NewCountTXDecorator(keys[wasm.StoreKey]),
		}

		anteDecorators = append(anteDecorators, wasmAnteDecorators...)
	}

	anteDecorators = append(anteDecorators,
		ibcante.NewAnteDecorator(GetKeeper[ibckeeper.Keeper](keepers)),
		ante.NewCheckRefundFeeDecorator(
			encodingConfig.InterfaceRegistry,
			GetKeeper[authkeeper.AccountKeeper](keepers),
			GetKeeper[stakingkeeper.Keeper](keepers),
			GetKeeper[snapKeeper.Keeper](keepers),
			GetKeeper[rewardKeeper.Keeper](keepers),
		),
		ante.NewAnteHandlerDecorator(
			initMessageAnteDecorators(encodingConfig, keepers).ToAnteHandler()),
	)
	return anteDecorators
}

func initMessageAnteDecorators(encodingConfig axelarParams.EncodingConfig, keepers *KeeperCache) ante.MessageAnteHandler {
	return ante.ChainMessageAnteDecorators(
		ante.NewLogMsgDecorator(encodingConfig.Codec),
		ante.NewCheckCommissionRate(GetKeeper[stakingkeeper.Keeper](keepers)),
		ante.NewUndelegateDecorator(
			GetKeeper[multisigKeeper.Keeper](keepers),
			GetKeeper[nexusKeeper.Keeper](keepers),
			GetKeeper[snapKeeper.Keeper](keepers),
		),

		ante.NewCheckProxy(GetKeeper[snapKeeper.Keeper](keepers)),
		ante.NewRestrictedTx(GetKeeper[permissionKeeper.Keeper](keepers)),
	)
}

func InitModuleAccountPermissions() map[string][]string {
	return map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		axelarnetTypes.ModuleName:      {authtypes.Minter, authtypes.Burner},
		rewardTypes.ModuleName:         {authtypes.Minter},
		wasm.ModuleName:                {authtypes.Burner},
		nexusTypes.ModuleName:          nil,
	}
}

func orderMigrations() []string {
	migrationOrder := []string{
		// auth module needs to go first
		authtypes.ModuleName,
		// sdk modules
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibchost.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
	}

	// wasm module needs to be added in a specific order, so we cannot just append it at the end
	if IsWasmEnabled() {
		migrationOrder = append(migrationOrder, wasm.ModuleName)
	}

	if IsIBCWasmHooksEnabled() {
		migrationOrder = append(migrationOrder, ibchookstypes.ModuleName)
	}

	// axelar modules
	migrationOrder = append(migrationOrder,
		multisigTypes.ModuleName,
		tssTypes.ModuleName,
		rewardTypes.ModuleName,
		voteTypes.ModuleName,
		evmTypes.ModuleName,
		nexusTypes.ModuleName,
		permissionTypes.ModuleName,
		snapTypes.ModuleName,
		axelarnetTypes.ModuleName,
		auxiliarytypes.ModuleName,
	)
	return migrationOrder
}

func orderBeginBlockers() []string {
	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	beginBlockerOrder := []string{
		// upgrades should be run first
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibchost.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
	}

	// wasm module needs to be added in a specific order, so we cannot just append it at the end
	if IsWasmEnabled() {
		beginBlockerOrder = append(beginBlockerOrder, wasm.ModuleName)
	}

	if IsIBCWasmHooksEnabled() {
		beginBlockerOrder = append(beginBlockerOrder, ibchookstypes.ModuleName)
	}

	// axelar custom modules
	beginBlockerOrder = append(beginBlockerOrder,
		rewardTypes.ModuleName,
		nexusTypes.ModuleName,
		permissionTypes.ModuleName,
		multisigTypes.ModuleName,
		tssTypes.ModuleName,
		evmTypes.ModuleName,
		snapTypes.ModuleName,
		axelarnetTypes.ModuleName,
		voteTypes.ModuleName,
		auxiliarytypes.ModuleName,
	)
	return beginBlockerOrder
}

func orderEndBlockers() []string {
	endBlockerOrder := []string{
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibchost.ModuleName,
		feegrant.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
	}

	// wasm module needs to be added in a specific order, so we cannot just append it at the end
	if IsWasmEnabled() {
		endBlockerOrder = append(endBlockerOrder, wasm.ModuleName)
	}

	if IsIBCWasmHooksEnabled() {
		endBlockerOrder = append(endBlockerOrder, ibchookstypes.ModuleName)
	}

	// axelar custom modules
	endBlockerOrder = append(endBlockerOrder,
		multisigTypes.ModuleName,
		tssTypes.ModuleName,
		evmTypes.ModuleName,
		nexusTypes.ModuleName,
		rewardTypes.ModuleName,
		snapTypes.ModuleName,
		axelarnetTypes.ModuleName,
		permissionTypes.ModuleName,
		voteTypes.ModuleName,
		auxiliarytypes.ModuleName,
	)
	return endBlockerOrder
}

func orderModulesForGenesis() []string {
	// Sets the order of Genesis - Order matters, genutil is to always come last
	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	genesisOrder := []string{
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		ibchost.ModuleName,
		evidencetypes.ModuleName,
		ibctransfertypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
	}

	// wasm module needs to be added in a specific order, so we cannot just append it at the end
	if IsWasmEnabled() {
		genesisOrder = append(genesisOrder, wasm.ModuleName)
	}

	if IsIBCWasmHooksEnabled() {
		genesisOrder = append(genesisOrder, ibchookstypes.ModuleName)
	}

	genesisOrder = append(genesisOrder,
		snapTypes.ModuleName,
		multisigTypes.ModuleName,
		tssTypes.ModuleName,
		nexusTypes.ModuleName,
		evmTypes.ModuleName, // Run evm end blocker after nexus so GMP calls routed to EVM get processed within the same block
		voteTypes.ModuleName,
		axelarnetTypes.ModuleName,
		rewardTypes.ModuleName,
		permissionTypes.ModuleName,
		auxiliarytypes.ModuleName,
	)
	return genesisOrder
}

func CreateStoreKeys() map[string]*sdk.KVStoreKey {
	keys := []string{authtypes.StoreKey,
		banktypes.StoreKey,
		stakingtypes.StoreKey,
		minttypes.StoreKey,
		distrtypes.StoreKey,
		slashingtypes.StoreKey,
		govtypes.StoreKey,
		paramstypes.StoreKey,
		upgradetypes.StoreKey,
		evidencetypes.StoreKey,
		ibchost.StoreKey,
		ibctransfertypes.StoreKey,
		capabilitytypes.StoreKey,
		feegrant.StoreKey,
		voteTypes.StoreKey,
		evmTypes.StoreKey,
		snapTypes.StoreKey,
		multisigTypes.StoreKey,
		tssTypes.StoreKey,
		nexusTypes.StoreKey,
		axelarnetTypes.StoreKey,
		rewardTypes.StoreKey,
		permissionTypes.StoreKey}

	if IsWasmEnabled() {
		keys = append(keys, wasm.StoreKey)
	}

	if IsIBCWasmHooksEnabled() {
		keys = append(keys, ibchookstypes.StoreKey)
	}

	return sdk.NewKVStoreKeys(keys...)
}

// GenesisState represents chain state at the start of the chain. Any initial state (account balances) are stored here.
type GenesisState map[string]json.RawMessage

// InitChainer handles the chain initialization from a genesis file
func (app *AxelarApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	GetKeeper[upgradekeeper.Keeper](app.Keepers).SetModuleVersionMap(ctx, app.mm.GetVersionMap())

	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// BeginBlocker calls the BeginBlock() function of every module at the beginning of a new block
func (app *AxelarApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker calls the EndBlock() function of every module at the end of a block
func (app *AxelarApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// LoadHeight loads the application version at a given height. It will panic if called
// more than once on a running baseapp.
func (app *AxelarApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// AppCodec returns AxelarApp's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *AxelarApp) AppCodec() codec.Codec {
	return app.appCodec
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *AxelarApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	rpc.RegisterRoutes(clientCtx, apiSvr.Router)
	// Register legacy tx routes.
	authrest.RegisterTxRoutes(clientCtx, apiSvr.Router)
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register legacy and grpc-gateway routes for all modules.
	GetModuleBasics().RegisterRESTRoutes(clientCtx, apiSvr.Router)
	GetModuleBasics().RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(apiSvr.Router)
	}
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticServer))
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *AxelarApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *AxelarApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.interfaceRegistry)
}

// GetModuleBasics initializes the module BasicManager is in charge of setting up basic,
// non-dependant module elements, such as codec registration and genesis verification.
// Initialization is dependent on whether wasm is enabled.
func GetModuleBasics() module.BasicManager {
	var wasmProposals []govclient.ProposalHandler
	if IsWasmEnabled() {
		wasmProposals = wasmclient.ProposalHandlers
	}

	managers := []module.AppModuleBasic{
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			append(
				wasmProposals,
				paramsclient.ProposalHandler,
				distrclient.ProposalHandler,
				upgradeclient.ProposalHandler,
				upgradeclient.CancelProposalHandler,
				ibcclientclient.UpdateClientProposalHandler,
				ibcclientclient.UpgradeProposalHandler,
				axelarnetclient.ProposalHandler,
			)...,
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		vesting.AppModuleBasic{},
		ibc.AppModuleBasic{},
		transfer.AppModuleBasic{},

		multisig.AppModuleBasic{},
		tss.AppModuleBasic{},
		vote.AppModuleBasic{},
		evm.AppModuleBasic{},
		snapshot.AppModuleBasic{},
		nexus.AppModuleBasic{},
		axelarnet.AppModuleBasic{},
		reward.AppModuleBasic{},
		permission.AppModuleBasic{},
		auxiliary.AppModuleBasic{},
	}

	if IsWasmEnabled() {
		managers = append(managers, NewWasmAppModuleBasicOverride(wasm.AppModuleBasic{}))
	}

	if IsIBCWasmHooksEnabled() {
		managers = append(managers, ibchooks.AppModuleBasic{})
	}

	return module.NewBasicManager(managers...)
}

// IsWasmEnabled returns whether wasm is enabled
func IsWasmEnabled() bool {
	return WasmEnabled == "true"
}

// IsIBCWasmHooksEnabled returns whether ibc wasm hooks are enabled
func IsIBCWasmHooksEnabled() bool {
	return IBCWasmHooksEnabled == "true"
}
