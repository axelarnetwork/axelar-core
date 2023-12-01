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

	// WasmCapabilities specifies the capabilities of the wasm vm
	// capabilities are detailed here: https://github.com/CosmWasm/cosmwasm/blob/main/docs/CAPABILITIES-BUILT-IN.md
	WasmCapabilities = ""
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
}

// AxelarApp defines the axelar Cosmos app that runs all modules
type AxelarApp struct {
	*bam.BaseApp

	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	// necessary keepers for export
	stakingKeeper  stakingkeeper.Keeper
	crisisKeeper   crisiskeeper.Keeper
	distrKeeper    distrkeeper.Keeper
	slashingKeeper slashingkeeper.Keeper

	// keys to access the substores
	keys map[string]*sdk.KVStoreKey

	mm            *module.Manager
	upgradeKeeper upgradekeeper.Keeper
}

// NewAxelarApp is a constructor function for axelar
func NewAxelarApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig axelarParams.EncodingConfig,
	appOpts servertypes.AppOptions,
	wasmOpts []wasm.Option,
	baseAppOptions ...func(*bam.BaseApp),
) *AxelarApp {

	keys := createStoreKeys()
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	keepers := newKeeperCache()
	setKeeper(keepers, initParamsKeeper(encodingConfig, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey]))

	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := initBaseApp(db, traceStore, encodingConfig, keepers, baseAppOptions, logger)

	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig := mustReadWasmConfig(appOpts)
	appCodec := encodingConfig.Codec
	moduleAccountPermissions := initModuleAccountPermissions()

	// set up predefined keepers
	setKeeper(keepers, initAccountKeeper(appCodec, keys, keepers, moduleAccountPermissions))
	setKeeper(keepers, initBankKeeper(appCodec, keys, keepers, moduleAccountPermissions))
	setKeeper(keepers, initStakingKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initMintKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initDistributionKeeper(appCodec, keys, keepers, moduleAccountPermissions))
	setKeeper(keepers, initSlashingKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initCrisisKeeper(keepers, invCheckPeriod))
	setKeeper(keepers, initUpgradeKeeper(appCodec, keys, skipUpgradeHeights, homePath, bApp))
	setKeeper(keepers, initEvidenceKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initFeegrantKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initCapabilityKeeper(appCodec, keys, memKeys))
	setKeeper(keepers, initIBCKeeper(appCodec, keys, keepers))

	// set up custom axelar keepers
	setKeeper(keepers, initAxelarnetKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initEvmKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initNexusKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initRewardKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initMultisigKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initTssKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initSnapshotKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initVoteKeeper(appCodec, keys, keepers))
	setKeeper(keepers, initPermissionKeeper(appCodec, keys, keepers))

	// set up ibc/wasm keepers
	wasmHooks := initWasmHooks(keys)
	ics4Wrapper := initICS4Wrapper(keepers, wasmHooks)
	setKeeper(keepers, initIBCTransferKeeper(appCodec, keys, keepers, ics4Wrapper))

	setKeeper(keepers, initAxelarIBCKeeper(keepers))
	setKeeper(keepers, initWasmKeeper(appCodec, keys, keepers, bApp, wasmDir, wasmConfig, wasmOpts))
	setKeeper(keepers, initWasmContractKeeper(keepers))

	// set the contract keeper for the Ics20WasmHooks
	wasmHooks.ContractKeeper = getKeeper[wasmkeeper.PermissionedKeeper](keepers)

	// set up governance keeper last when it has access to all other keepers to set up governance routes
	setKeeper(keepers, initGovernanceKeeper(appCodec, keys, keepers))

	// seal capability keeper after all keepers are set to be certain that all capabilities have been registered
	getKeeper[capabilitykeeper.Keeper](keepers).Seal()

	// set routers
	getKeeper[nexusKeeper.Keeper](keepers).SetMessageRouter(initMessageRouter(keepers))
	getKeeper[ibckeeper.Keeper](keepers).SetRouter(initIBCRouter(keepers, initIBCMiddleware(keepers, ics4Wrapper)))

	// register the staking hooks
	getKeeper[stakingkeeper.Keeper](keepers).SetHooks(
		stakingtypes.NewMultiStakingHooks(
			getKeeper[distrkeeper.Keeper](keepers).Hooks(),
			getKeeper[slashingkeeper.Keeper](keepers).Hooks(),
		),
	)

	/****  Module Options ****/

	appModules := initAppModules(
		keepers,
		bApp,
		encodingConfig,
		appOpts,
		axelarnet.NewAppModule(
			*getKeeper[axelarnetKeeper.IBCKeeper](keepers),
			getKeeper[nexusKeeper.Keeper](keepers),
			axelarbankkeeper.NewBankKeeper(getKeeper[bankkeeper.BaseKeeper](keepers)),
			getKeeper[authkeeper.AccountKeeper](keepers),
			logger,
		),
	)

	mm := module.NewManager(appModules...)
	mm.SetOrderMigrations(orderMigrations()...)
	mm.SetOrderBeginBlockers(orderBeginBlockers()...)
	mm.SetOrderEndBlockers(orderEndBlockers()...)
	mm.SetOrderInitGenesis(orderModulesForGenesis()...)

	mm.RegisterInvariants(getKeeper[crisiskeeper.Keeper](keepers))

	// register all module routes and module queriers
	mm.RegisterRoutes(bApp.Router(), bApp.QueryRouter(), encodingConfig.Amino)
	configurator := module.NewConfigurator(appCodec, bApp.MsgServiceRouter(), bApp.GRPCQueryRouter())
	mm.RegisterServices(configurator)

	var app = &AxelarApp{
		BaseApp:           bApp,
		appCodec:          appCodec,
		interfaceRegistry: encodingConfig.InterfaceRegistry,
		stakingKeeper:     *getKeeper[stakingkeeper.Keeper](keepers),
		crisisKeeper:      *getKeeper[crisiskeeper.Keeper](keepers),
		distrKeeper:       *getKeeper[distrkeeper.Keeper](keepers),
		slashingKeeper:    *getKeeper[slashingkeeper.Keeper](keepers),
		keys:              keys,
		mm:                mm,
		upgradeKeeper:     *getKeeper[upgradekeeper.Keeper](keepers),
	}

	app.setUpgradeBehaviour(configurator)

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// The initChainer handles translating the genesis.json file into initial state for the network
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	app.SetAnteHandler(initAnteHandlers(encodingConfig, keys, keepers, wasmConfig))

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		if IsWasmEnabled() {
			ctx := app.BaseApp.NewUncachedContext(true, tmproto.Header{})

			// Initialize pinned codes in wasmvm as they are not persisted there
			if err := getKeeper[wasm.Keeper](keepers).InitializePinnedCodes(ctx); err != nil {
				tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
			}
		}
	}

	/* ==== at this point all stores are fully loaded ==== */

	// we need to ensure that all chain subspaces are loaded at start-up to prevent unexpected consensus failures
	// when the params keeper is used outside the evm module's context
	getKeeper[evmKeeper.BaseKeeper](keepers).InitChains(app.NewContext(true, tmproto.Header{}))

	return app
}

func initICS4Wrapper(keepers *keeperCache, wasmHooks ibchooks.WasmHooks) ibchooks.ICS4Middleware {
	// ICS4Wrapper deals with sending IBC packets. These need to get rate limited when appropriate,
	// so we wrap the channel keeper (which implements the ICS4Wrapper interface) with a rate limiter.
	ics4Wrapper := axelarnet.NewRateLimitedICS4Wrapper(
		getKeeper[ibckeeper.Keeper](keepers).ChannelKeeper,
		axelarnet.NewRateLimiter(getKeeper[axelarnetKeeper.Keeper](keepers), getKeeper[nexusKeeper.Keeper](keepers)),
		getKeeper[axelarnetKeeper.Keeper](keepers),
	)
	// create a middleware to integrate wasm hooks into the ibc pipeline
	return ibchooks.NewICS4Middleware(ics4Wrapper, wasmHooks)
}

func initIBCMiddleware(keepers *keeperCache, ics4Middleware ibchooks.ICS4Middleware) ibchooks.IBCMiddleware {
	// IBCModule deals with received IBC packets. These need to get rate limited when appropriate,
	// so we wrap the transfer module's IBCModule with a rate limiter.
	ibcModule := axelarnet.NewAxelarnetIBCModule(
		transfer.NewIBCModule(*getKeeper[ibctransferkeeper.Keeper](keepers)),
		*getKeeper[axelarnetKeeper.IBCKeeper](keepers),
		axelarnet.NewRateLimiter(getKeeper[axelarnetKeeper.Keeper](keepers), getKeeper[nexusKeeper.Keeper](keepers)),
		getKeeper[nexusKeeper.Keeper](keepers),
		axelarbankkeeper.NewBankKeeper(getKeeper[bankkeeper.BaseKeeper](keepers)),
	)

	// By merging the middlewares the receiving IBC Module has access to all registered hooks in the ICS4Middleware
	return ibchooks.NewIBCMiddleware(ibcModule, &ics4Middleware)
}

func initWasmHooks(keys map[string]*sdk.KVStoreKey) ibchooks.WasmHooks {
	var wasmHooks ibchooks.WasmHooks
	if !IsWasmEnabled() {
		return wasmHooks
	}

	// Configure the IBC hooks keeper to make wasm calls via IBC transfer memo
	ibcHooksKeeper := ibchookskeeper.NewKeeper(keys[ibchookstypes.StoreKey])

	// The contract keeper needs to be set later
	return ibchooks.NewWasmHooks(&ibcHooksKeeper, nil, sdk.GetConfig().GetBech32AccountAddrPrefix())
}

func initIBCRouter(keepers *keeperCache, axelarnetModule porttypes.IBCModule) *porttypes.Router {
	// Finalize the IBC router
	// Create static IBC router, add axelarnet module as the IBC transfer route, and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, axelarnetModule)
	if IsWasmEnabled() {
		// Create wasm ibc stack
		var wasmStack porttypes.IBCModule = wasm.NewIBCHandler(
			getKeeper[wasm.Keeper](keepers),
			getKeeper[ibckeeper.Keeper](keepers).ChannelKeeper,
			getKeeper[ibckeeper.Keeper](keepers).ChannelKeeper,
		)
		ibcRouter.AddRoute(wasm.ModuleName, wasmStack)
	}
	return ibcRouter
}

func initMessageRouter(keepers *keeperCache) nexusTypes.MessageRouter {
	messageRouter := nexusTypes.NewMessageRouter().
		AddRoute(evmTypes.ModuleName, evmKeeper.NewMessageRoute()).
		AddRoute(axelarnetTypes.ModuleName, axelarnetKeeper.NewMessageRoute(
			*getKeeper[axelarnetKeeper.Keeper](keepers),
			getKeeper[axelarnetKeeper.IBCKeeper](keepers),
			getKeeper[feegrantkeeper.Keeper](keepers),
			axelarbankkeeper.NewBankKeeper(getKeeper[bankkeeper.BaseKeeper](keepers)),
			getKeeper[nexusKeeper.Keeper](keepers),
			getKeeper[authkeeper.AccountKeeper](keepers),
		))

	if IsWasmEnabled() {
		messageRouter.AddRoute(wasm.ModuleName, nexusKeeper.NewMessageRoute(
			getKeeper[nexusKeeper.Keeper](keepers),
			getKeeper[authkeeper.AccountKeeper](keepers),
			getKeeper[wasmkeeper.PermissionedKeeper](keepers),
		))
	}
	return messageRouter
}

func (app *AxelarApp) setUpgradeBehaviour(configurator module.Configurator) {
	app.upgradeKeeper.SetUpgradeHandler(
		upgradeName(app.Version()),
		func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			return app.mm.RunMigrations(ctx, configurator, fromVM)
		},
	)

	upgradeInfo, err := app.upgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	if upgradeInfo.Name == upgradeName(app.Version()) && !app.upgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		storeUpgrades := store.StoreUpgrades{}

		if IsWasmEnabled() {
			storeUpgrades.Added = append(storeUpgrades.Added, ibchookstypes.StoreKey)
			storeUpgrades.Added = append(storeUpgrades.Added, wasm.ModuleName)
		}

		// configure store loader that checks if version == upgradeHeight and applies store upgrades
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
	}
}

func initBaseApp(db dbm.DB, traceStore io.Writer, encodingConfig axelarParams.EncodingConfig, keepers *keeperCache, baseAppOptions []func(*bam.BaseApp), logger log.Logger) *bam.BaseApp {
	bApp := bam.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)
	bApp.SetParamStore(keepers.getSubspace(bam.Paramspace))
	return bApp
}

func initAppModules(keepers *keeperCache, bApp *bam.BaseApp, encodingConfig axelarParams.EncodingConfig, appOpts servertypes.AppOptions, axelarnetModule axelarnet.AppModule) []module.AppModule {
	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	appCodec := encodingConfig.Codec

	appModules := []module.AppModule{
		genutil.NewAppModule(getKeeper[authkeeper.AccountKeeper](keepers), getKeeper[stakingkeeper.Keeper](keepers), bApp.DeliverTx, encodingConfig.TxConfig),
		auth.NewAppModule(appCodec, *getKeeper[authkeeper.AccountKeeper](keepers), nil),
		vesting.NewAppModule(*getKeeper[authkeeper.AccountKeeper](keepers), getKeeper[bankkeeper.BaseKeeper](keepers)),

		// bank module accepts a reference to the base keeper, but panics when RegisterService is called on a reference, so we need to dereference it
		bank.NewAppModule(appCodec, *getKeeper[bankkeeper.BaseKeeper](keepers), getKeeper[authkeeper.AccountKeeper](keepers)),
		crisis.NewAppModule(getKeeper[crisiskeeper.Keeper](keepers), skipGenesisInvariants),
		gov.NewAppModule(appCodec, *getKeeper[govkeeper.Keeper](keepers), getKeeper[authkeeper.AccountKeeper](keepers), getKeeper[bankkeeper.BaseKeeper](keepers)),
		mint.NewAppModule(appCodec, *getKeeper[mintkeeper.Keeper](keepers), getKeeper[authkeeper.AccountKeeper](keepers)),
		slashing.NewAppModule(appCodec, *getKeeper[slashingkeeper.Keeper](keepers), getKeeper[authkeeper.AccountKeeper](keepers), getKeeper[bankkeeper.BaseKeeper](keepers), getKeeper[stakingkeeper.Keeper](keepers)),
		distr.NewAppModule(appCodec, *getKeeper[distrkeeper.Keeper](keepers), getKeeper[authkeeper.AccountKeeper](keepers), getKeeper[bankkeeper.BaseKeeper](keepers), getKeeper[stakingkeeper.Keeper](keepers)),
		staking.NewAppModule(appCodec, *getKeeper[stakingkeeper.Keeper](keepers), getKeeper[authkeeper.AccountKeeper](keepers), getKeeper[bankkeeper.BaseKeeper](keepers)),
		upgrade.NewAppModule(*getKeeper[upgradekeeper.Keeper](keepers)),
		evidence.NewAppModule(*getKeeper[evidencekeeper.Keeper](keepers)),
		params.NewAppModule(*getKeeper[paramskeeper.Keeper](keepers)),
		capability.NewAppModule(appCodec, *getKeeper[capabilitykeeper.Keeper](keepers)),
	}

	// wasm module needs to be added in a specific order
	if IsWasmEnabled() {
		appModules = append(
			appModules,
			wasm.NewAppModule(
				appCodec,
				getKeeper[wasm.Keeper](keepers),
				getKeeper[stakingkeeper.Keeper](keepers),
				getKeeper[authkeeper.AccountKeeper](keepers),
				getKeeper[bankkeeper.BaseKeeper](keepers),
			),
			ibchooks.NewAppModule(getKeeper[authkeeper.AccountKeeper](keepers)),
		)
	}

	appModules = append(appModules,
		evidence.NewAppModule(*getKeeper[evidencekeeper.Keeper](keepers)),
		ibc.NewAppModule(getKeeper[ibckeeper.Keeper](keepers)),
		transfer.NewAppModule(*getKeeper[ibctransferkeeper.Keeper](keepers)),
		feegrantmodule.NewAppModule(
			appCodec,
			getKeeper[authkeeper.AccountKeeper](keepers),
			getKeeper[bankkeeper.BaseKeeper](keepers),
			*getKeeper[feegrantkeeper.Keeper](keepers),
			encodingConfig.InterfaceRegistry,
		),

		snapshot.NewAppModule(*getKeeper[snapKeeper.Keeper](keepers)),
		multisig.NewAppModule(
			*getKeeper[multisigKeeper.Keeper](keepers),
			getKeeper[stakingkeeper.Keeper](keepers),
			getKeeper[slashingkeeper.Keeper](keepers),
			getKeeper[snapKeeper.Keeper](keepers),
			getKeeper[rewardKeeper.Keeper](keepers),
			getKeeper[nexusKeeper.Keeper](keepers),
		),
		tss.NewAppModule(
			*getKeeper[tssKeeper.Keeper](keepers),
			getKeeper[snapKeeper.Keeper](keepers),
			getKeeper[nexusKeeper.Keeper](keepers),
			getKeeper[stakingkeeper.Keeper](keepers),
			getKeeper[multisigKeeper.Keeper](keepers),
		),
		vote.NewAppModule(*getKeeper[voteKeeper.Keeper](keepers)),
		nexus.NewAppModule(
			*getKeeper[nexusKeeper.Keeper](keepers),
			getKeeper[snapKeeper.Keeper](keepers),
			getKeeper[slashingkeeper.Keeper](keepers),
			getKeeper[stakingkeeper.Keeper](keepers),
			getKeeper[axelarnetKeeper.Keeper](keepers),
			getKeeper[rewardKeeper.Keeper](keepers),
		),
		evm.NewAppModule(
			getKeeper[evmKeeper.BaseKeeper](keepers),
			getKeeper[voteKeeper.Keeper](keepers),
			getKeeper[nexusKeeper.Keeper](keepers),
			getKeeper[snapKeeper.Keeper](keepers),
			getKeeper[stakingkeeper.Keeper](keepers),
			getKeeper[slashingkeeper.Keeper](keepers),
			getKeeper[multisigKeeper.Keeper](keepers),
		),
		axelarnetModule,
		reward.NewAppModule(
			*getKeeper[rewardKeeper.Keeper](keepers),
			getKeeper[nexusKeeper.Keeper](keepers),
			getKeeper[mintkeeper.Keeper](keepers),
			getKeeper[stakingkeeper.Keeper](keepers),
			getKeeper[slashingkeeper.Keeper](keepers),
			getKeeper[multisigKeeper.Keeper](keepers),
			getKeeper[snapKeeper.Keeper](keepers),
			getKeeper[bankkeeper.BaseKeeper](keepers),
			bApp.MsgServiceRouter(),
			bApp.Router(),
		),
		permission.NewAppModule(*getKeeper[permissionKeeper.Keeper](keepers)),
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

func initAnteHandlers(encodingConfig axelarParams.EncodingConfig, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, wasmConfig wasmtypes.WasmConfig) sdk.AnteHandler {
	// The baseAnteHandler handles signature verification and transaction pre-processing
	baseAnteHandler, err := authAnte.NewAnteHandler(
		authAnte.HandlerOptions{
			AccountKeeper:   getKeeper[authkeeper.AccountKeeper](keepers),
			BankKeeper:      getKeeper[bankkeeper.BaseKeeper](keepers),
			SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
			FeegrantKeeper:  getKeeper[feegrantkeeper.Keeper](keepers),
			SigGasConsumer:  authAnte.DefaultSigVerificationGasConsumer,
		},
	)
	if err != nil {
		panic(err)
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewAnteHandlerDecorator(baseAnteHandler),
	}

	// enforce wasm limits earlier in the ante handler chain
	if IsWasmEnabled() {
		wasmAnteDecorators := []sdk.AnteDecorator{
			wasmkeeper.NewLimitSimulationGasDecorator(wasmConfig.SimulationGasLimit),
			wasmkeeper.NewCountTXDecorator(keys[wasm.StoreKey]),
		}

		anteDecorators = append(anteDecorators, wasmAnteDecorators...)
	}

	anteDecorators = append(anteDecorators,
		ante.NewLogMsgDecorator(encodingConfig.Codec),
		ante.NewCheckCommissionRate(getKeeper[stakingkeeper.Keeper](keepers)),
		ante.NewUndelegateDecorator(
			getKeeper[multisigKeeper.Keeper](keepers),
			getKeeper[nexusKeeper.Keeper](keepers),
			getKeeper[snapKeeper.Keeper](keepers),
		),
		ante.NewCheckRefundFeeDecorator(
			encodingConfig.InterfaceRegistry,
			getKeeper[authkeeper.AccountKeeper](keepers),
			getKeeper[stakingkeeper.Keeper](keepers),
			getKeeper[snapKeeper.Keeper](keepers),
			getKeeper[rewardKeeper.Keeper](keepers),
		),
		ante.NewCheckProxy(getKeeper[snapKeeper.Keeper](keepers)),
		ante.NewRestrictedTx(getKeeper[permissionKeeper.Keeper](keepers)),
		ibcante.NewAnteDecorator(getKeeper[ibckeeper.Keeper](keepers)),
	)

	anteHandler := sdk.ChainAnteDecorators(
		anteDecorators...,
	)
	return anteHandler
}

func initModuleAccountPermissions() map[string][]string {
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

	// wasm module needs to be added in a specific order
	if IsWasmEnabled() {
		migrationOrder = append(migrationOrder, wasm.ModuleName, ibchookstypes.ModuleName)
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

	// wasm module needs to be added in a specific order
	if IsWasmEnabled() {
		beginBlockerOrder = append(beginBlockerOrder, wasm.ModuleName, ibchookstypes.ModuleName)
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

	// wasm module needs to be added in a specific order
	if IsWasmEnabled() {
		endBlockerOrder = append(endBlockerOrder, wasm.ModuleName, ibchookstypes.ModuleName)
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

	// wasm module needs to be added in a specific order
	if IsWasmEnabled() {
		genesisOrder = append(genesisOrder, wasm.ModuleName, ibchookstypes.ModuleName)
	}

	genesisOrder = append(genesisOrder,
		snapTypes.ModuleName,
		multisigTypes.ModuleName,
		tssTypes.ModuleName,
		evmTypes.ModuleName,
		nexusTypes.ModuleName,
		voteTypes.ModuleName,
		axelarnetTypes.ModuleName,
		rewardTypes.ModuleName,
		permissionTypes.ModuleName,
	)
	return genesisOrder
}

func createStoreKeys() map[string]*sdk.KVStoreKey {
	return sdk.NewKVStoreKeys(authtypes.StoreKey,
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
		wasm.StoreKey,
		ibchookstypes.StoreKey,
		voteTypes.StoreKey,
		evmTypes.StoreKey,
		snapTypes.StoreKey,
		multisigTypes.StoreKey,
		tssTypes.StoreKey,
		nexusTypes.StoreKey,
		axelarnetTypes.StoreKey,
		rewardTypes.StoreKey,
		permissionTypes.StoreKey)
}

// GenesisState represents chain state at the start of the chain. Any initial state (account balances) are stored here.
type GenesisState map[string]json.RawMessage

// InitChainer handles the chain initialization from a genesis file
func (app *AxelarApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	app.upgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())

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
	}

	if IsWasmEnabled() {
		managers = append(managers, wasm.AppModuleBasic{}, ibchooks.AppModuleBasic{})
	}

	return module.NewBasicManager(managers...)
}

// IsWasmEnabled returns whether wasm is enabled
func IsWasmEnabled() bool {
	return WasmEnabled != ""
}
