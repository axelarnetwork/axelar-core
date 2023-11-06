package app

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibcclient "github.com/cosmos/ibc-go/v4/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibchost "github.com/cosmos/ibc-go/v4/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v4/modules/core/keeper"
	"golang.org/x/mod/semver"

	"github.com/axelarnetwork/axelar-core/x/axelarnet"
	axelarnetKeeper "github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	axelarnetTypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	axelarbankkeeper "github.com/axelarnetwork/axelar-core/x/bank/keeper"
	evmKeeper "github.com/axelarnetwork/axelar-core/x/evm/keeper"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	multisigKeeper "github.com/axelarnetwork/axelar-core/x/multisig/keeper"
	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	permissionKeeper "github.com/axelarnetwork/axelar-core/x/permission/keeper"
	permissionTypes "github.com/axelarnetwork/axelar-core/x/permission/types"
	rewardKeeper "github.com/axelarnetwork/axelar-core/x/reward/keeper"
	rewardTypes "github.com/axelarnetwork/axelar-core/x/reward/types"
	snapKeeper "github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	tssKeeper "github.com/axelarnetwork/axelar-core/x/tss/keeper"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	voteKeeper "github.com/axelarnetwork/axelar-core/x/vote/keeper"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
	"github.com/axelarnetwork/utils/maps"
)

type keeperCache struct {
	repository map[string]any
}

func newKeeperCache() *keeperCache {
	return &keeperCache{
		repository: make(map[string]any),
	}
}

func (k *keeperCache) getSubspace(moduleName string) paramstypes.Subspace {
	paramsK := getKeeper[paramskeeper.Keeper](k)
	subspace, ok := paramsK.GetSubspace(moduleName)
	if !ok {
		panic(fmt.Sprintf("subspace %s not found", moduleName))
	}
	return subspace
}

func getKeeper[T any](k *keeperCache) T {
	key := fullTypeName[T]()
	keeper, ok := k.repository[key].(T)
	if !ok {
		panic(fmt.Sprintf("keeper %s not found", key))
	}
	return keeper
}

func setKeeper[T any](k *keeperCache, keeper T) {
	k.repository[fullTypeName[T]()] = keeper
}

func fullTypeName[T any]() string {
	keeperType := reflect.TypeOf(*new(T))

	var prefix string
	if keeperType.Kind() == reflect.Ptr {
		prefix = "*"
		keeperType = keeperType.Elem()
	}

	return prefix + keeperType.PkgPath() + "." + keeperType.Name()
}

func initGovernanceKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) govkeeper.Keeper {
	// Add governance proposal hooks
	govRouter := govtypes.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(getKeeper[paramskeeper.Keeper](keepers))).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(getKeeper[distrkeeper.Keeper](keepers))).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(getKeeper[upgradekeeper.Keeper](keepers))).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(getKeeper[*ibckeeper.Keeper](keepers).ClientKeeper)).
		AddRoute(axelarnetTypes.RouterKey, axelarnet.NewProposalHandler(getKeeper[axelarnetKeeper.Keeper](keepers), getKeeper[nexusKeeper.Keeper](keepers), getKeeper[authkeeper.AccountKeeper](keepers)))

	if IsWasmEnabled() {
		govRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(getKeeper[wasm.Keeper](keepers), wasm.EnableAllProposals))
	}

	govK := govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], keepers.getSubspace(govtypes.ModuleName), getKeeper[authkeeper.AccountKeeper](keepers), getKeeper[bankkeeper.BaseKeeper](keepers),
		getKeeper[stakingkeeper.Keeper](keepers), govRouter,
	)

	axelarnetK := getKeeper[axelarnetKeeper.Keeper](keepers)
	govK.SetHooks(govtypes.NewMultiGovHooks(axelarnetK.Hooks(getKeeper[nexusKeeper.Keeper](keepers), govK)))
	return govK
}

func initPermissionKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) permissionKeeper.Keeper {
	return permissionKeeper.NewKeeper(appCodec, keys[permissionTypes.StoreKey], keepers.getSubspace(permissionTypes.ModuleName))
}

func initVoteKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) voteKeeper.Keeper {
	voteRouter := voteTypes.NewRouter()
	voteRouter.AddHandler(
		evmTypes.ModuleName,
		evmKeeper.NewVoteHandler(
			appCodec,
			getKeeper[*evmKeeper.BaseKeeper](keepers),
			getKeeper[nexusKeeper.Keeper](keepers),
			getKeeper[rewardKeeper.Keeper](keepers),
		),
	)

	voteK := voteKeeper.NewKeeper(
		appCodec,
		keys[voteTypes.StoreKey],
		keepers.getSubspace(voteTypes.ModuleName),
		getKeeper[snapKeeper.Keeper](keepers),
		getKeeper[stakingkeeper.Keeper](keepers),
		getKeeper[rewardKeeper.Keeper](keepers),
	)
	voteK.SetVoteRouter(voteRouter)
	return voteK
}

func initSnapshotKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) snapKeeper.Keeper {
	return snapKeeper.NewKeeper(
		appCodec,
		keys[snapTypes.StoreKey],
		keepers.getSubspace(snapTypes.ModuleName),
		getKeeper[stakingkeeper.Keeper](keepers),
		axelarbankkeeper.NewBankKeeper(getKeeper[bankkeeper.BaseKeeper](keepers)),
		getKeeper[slashingkeeper.Keeper](keepers),
	)
}

func initTssKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) tssKeeper.Keeper {
	return tssKeeper.NewKeeper(appCodec, keys[tssTypes.StoreKey], keepers.getSubspace(tssTypes.ModuleName))
}

func initMultisigKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) multisigKeeper.Keeper {
	multisigRouter := multisigTypes.NewSigRouter()
	multisigRouter.AddHandler(evmTypes.ModuleName, evmKeeper.NewSigHandler(appCodec, getKeeper[*evmKeeper.BaseKeeper](keepers)))

	multisigK := multisigKeeper.NewKeeper(appCodec, keys[multisigTypes.StoreKey], keepers.getSubspace(multisigTypes.ModuleName))
	multisigK.SetSigRouter(multisigRouter)
	return multisigK
}

func initRewardKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) rewardKeeper.Keeper {
	return rewardKeeper.NewKeeper(
		appCodec,
		keys[rewardTypes.StoreKey],
		keepers.getSubspace(rewardTypes.ModuleName),
		axelarbankkeeper.NewBankKeeper(getKeeper[bankkeeper.BaseKeeper](keepers)),
		getKeeper[distrkeeper.Keeper](keepers),
		getKeeper[stakingkeeper.Keeper](keepers),
	)
}

func initIBCKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, scopedIBCK capabilitykeeper.ScopedKeeper) *ibckeeper.Keeper {
	return ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		keepers.getSubspace(ibchost.ModuleName),
		getKeeper[stakingkeeper.Keeper](keepers),
		getKeeper[upgradekeeper.Keeper](keepers),
		scopedIBCK,
	)
}

func initNexusKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) nexusKeeper.Keeper {
	// Setting Router will finalize all routes by sealing router
	// No more routes can be added
	nexusRouter := nexusTypes.NewRouter()
	nexusRouter.
		AddAddressValidator(evmTypes.ModuleName, evmKeeper.NewAddressValidator()).
		AddAddressValidator(axelarnetTypes.ModuleName, axelarnetKeeper.NewAddressValidator(getKeeper[axelarnetKeeper.Keeper](keepers)))

	nexusK := nexusKeeper.NewKeeper(appCodec, keys[nexusTypes.StoreKey], keepers.getSubspace(nexusTypes.ModuleName))
	nexusK.SetRouter(nexusRouter)
	return nexusK
}

func initFeegrantKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) feegrantkeeper.Keeper {
	return feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], getKeeper[authkeeper.AccountKeeper](keepers))
}

func initEvidenceKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, stakingK *stakingkeeper.Keeper) evidencekeeper.Keeper {
	// there is no point in this constructor returning a reference, so we deref it
	evidenceK := evidencekeeper.NewKeeper(appCodec, keys[evidencetypes.StoreKey], stakingK, getKeeper[slashingkeeper.Keeper](keepers))
	return *evidenceK
}

// todo: clean this up
func initUpgradeKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, skipUpgradeHeights map[int64]bool, homePath string, bApp *bam.BaseApp, configurator *module.Configurator, mm *module.Manager) upgradekeeper.Keeper {
	upgradeK := upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath, bApp)
	upgradeK.SetUpgradeHandler(
		upgradeName(bApp.Version()),
		func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			return mm.RunMigrations(ctx, *configurator, fromVM)
		},
	)
	return upgradeK
}

func upgradeName(version string) string {
	if !strings.HasPrefix(version, "v") {
		version = fmt.Sprintf("v%s", version)
	}
	name := semver.MajorMinor(version)
	if name == "" {
		panic(fmt.Errorf("invalid app version %s", version))
	}
	return name
}

func initCrisisKeeper(keepers *keeperCache, invCheckPeriod uint) crisiskeeper.Keeper {
	return crisiskeeper.NewKeeper(
		keepers.getSubspace(crisistypes.ModuleName),
		invCheckPeriod,
		getKeeper[bankkeeper.BaseKeeper](keepers),
		authtypes.FeeCollectorName,
	)
}

func initSlashingKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, stakingK *stakingkeeper.Keeper) slashingkeeper.Keeper {
	return slashingkeeper.NewKeeper(appCodec, keys[slashingtypes.StoreKey], stakingK, keepers.getSubspace(slashingtypes.ModuleName))
}

func initDistributionKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, stakingK *stakingkeeper.Keeper, moduleAccPerms map[string][]string) distrkeeper.Keeper {
	return distrkeeper.NewKeeper(
		appCodec,
		keys[distrtypes.StoreKey],
		keepers.getSubspace(distrtypes.ModuleName),
		getKeeper[authkeeper.AccountKeeper](keepers),
		getKeeper[bankkeeper.BaseKeeper](keepers),
		stakingK,
		authtypes.FeeCollectorName,
		moduleAccountAddrs(moduleAccPerms),
	)
}

func initMintKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, stakingK *stakingkeeper.Keeper) mintkeeper.Keeper {
	return mintkeeper.NewKeeper(
		appCodec,
		keys[minttypes.StoreKey],
		keepers.getSubspace(minttypes.ModuleName),
		stakingK,
		getKeeper[authkeeper.AccountKeeper](keepers),
		getKeeper[bankkeeper.BaseKeeper](keepers),
		authtypes.FeeCollectorName,
	)
}

func initBankKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, moduleAccPerms map[string][]string) bankkeeper.BaseKeeper {
	return bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		getKeeper[authkeeper.AccountKeeper](keepers),
		keepers.getSubspace(banktypes.ModuleName),
		maps.Filter(moduleAccountAddrs(moduleAccPerms), func(addr string, _ bool) bool {
			// we do not rely on internal balance tracking for invariance checks in the axelarnet module
			// (https://github.com/cosmos/cosmos-sdk/issues/12825 for more details on the purpose of the blocked list),
			// but the module address must be able to use ibc transfers,
			// so we exclude this address from the blocked list
			return addr != authtypes.NewModuleAddress(axelarnetTypes.ModuleName).String()
		}),
	)
}

func initAccountKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, moduleAccPerms map[string][]string) authkeeper.AccountKeeper {
	return authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		keepers.getSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		moduleAccPerms,
	)
}

// moduleAccountAddrs returns all the app's module account addresses.
func moduleAccountAddrs(moduleAccPerms map[string][]string) map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range moduleAccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}
