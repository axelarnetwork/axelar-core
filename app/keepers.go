package app

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
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
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v4/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v4/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibchost "github.com/cosmos/ibc-go/v4/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v4/modules/core/keeper"
	"golang.org/x/mod/semver"

	axelarParams "github.com/axelarnetwork/axelar-core/app/params"
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

func getKeeper[T any](k *keeperCache) *T {
	if reflect.TypeOf(*new(T)).Kind() == reflect.Ptr {
		panic(fmt.Sprintf("the generic parameter for %s cannot be a reference type", fullTypeName[T]()))
	}
	key := fullTypeName[T]()
	keeper, ok := k.repository[key].(*T)
	if !ok {
		panic(fmt.Sprintf("keeper %s not found", key))
	}
	return keeper
}

func setKeeper[T any](k *keeperCache, keeper T) {
	if reflect.TypeOf(keeper).Kind() != reflect.Ptr {
		panic(fmt.Sprintf("keeper %s must be a reference type", fullTypeName[T]()))
	}

	k.repository[fullTypeName[T]()] = keeper
}

func fullTypeName[T any]() string {
	keeperType := reflect.TypeOf(*new(T))

	if keeperType.Kind() == reflect.Ptr {
		keeperType = keeperType.Elem()
	}

	return keeperType.PkgPath() + "." + keeperType.Name()
}

func initParamsKeeper(encodingConfig axelarParams.EncodingConfig, key, tkey sdk.StoreKey) *paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(encodingConfig.Codec, encodingConfig.Amino, key, tkey)

	paramsKeeper.Subspace(bam.Paramspace).WithKeyTable(paramskeeper.ConsensusParamsKeyTable())

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypes.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(wasm.ModuleName)
	paramsKeeper.Subspace(snapTypes.ModuleName)
	paramsKeeper.Subspace(multisigTypes.ModuleName)
	paramsKeeper.Subspace(tssTypes.ModuleName)
	paramsKeeper.Subspace(nexusTypes.ModuleName)
	paramsKeeper.Subspace(axelarnetTypes.ModuleName)
	paramsKeeper.Subspace(rewardTypes.ModuleName)
	paramsKeeper.Subspace(voteTypes.ModuleName)
	paramsKeeper.Subspace(permissionTypes.ModuleName)

	return &paramsKeeper
}

func initStakingKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *stakingkeeper.Keeper {
	stakingK := stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		getKeeper[authkeeper.AccountKeeper](keepers),
		getKeeper[bankkeeper.BaseKeeper](keepers),
		keepers.getSubspace(stakingtypes.ModuleName),
	)
	return &stakingK
}

func initWasmKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, bApp *bam.BaseApp, wasmDir string, wasmConfig wasmtypes.WasmConfig, wasmOpts []wasm.Option) *wasm.Keeper {
	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	wasmOpts = append(wasmOpts, wasmkeeper.WithMessageHandlerDecorator(func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
		return wasmkeeper.NewMessageHandlerChain(old, nexusKeeper.NewMessenger(getKeeper[nexusKeeper.Keeper](keepers)))
	}))

	scopedWasmK := getKeeper[capabilitykeeper.Keeper](keepers).ScopeToModule(wasm.ModuleName)
	ibcKeeper := getKeeper[ibckeeper.Keeper](keepers)
	wasmK := wasm.NewKeeper(
		appCodec,
		keys[wasm.StoreKey],
		keepers.getSubspace(wasm.ModuleName),
		getKeeper[authkeeper.AccountKeeper](keepers),
		getKeeper[bankkeeper.BaseKeeper](keepers),
		getKeeper[stakingkeeper.Keeper](keepers),
		getKeeper[distrkeeper.Keeper](keepers),
		ibcKeeper.ChannelKeeper,
		ibcKeeper.ChannelKeeper,
		&ibcKeeper.PortKeeper,
		scopedWasmK,
		getKeeper[ibctransferkeeper.Keeper](keepers),
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		WasmCapabilities,
		wasmOpts...,
	)

	return &wasmK
}

func initGovernanceKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *govkeeper.Keeper {
	// Add governance proposal hooks
	govRouter := govtypes.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypes.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(*getKeeper[paramskeeper.Keeper](keepers))).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(*getKeeper[distrkeeper.Keeper](keepers))).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(*getKeeper[upgradekeeper.Keeper](keepers))).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(getKeeper[ibckeeper.Keeper](keepers).ClientKeeper)).
		AddRoute(axelarnetTypes.RouterKey, axelarnet.NewProposalHandler(*getKeeper[axelarnetKeeper.Keeper](keepers), getKeeper[nexusKeeper.Keeper](keepers), getKeeper[authkeeper.AccountKeeper](keepers)))

	if IsWasmEnabled() {
		govRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(getKeeper[wasm.Keeper](keepers), wasm.EnableAllProposals))
	}

	govK := govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], keepers.getSubspace(govtypes.ModuleName), getKeeper[authkeeper.AccountKeeper](keepers), getKeeper[bankkeeper.BaseKeeper](keepers),
		getKeeper[stakingkeeper.Keeper](keepers), govRouter,
	)

	axelarnetK := getKeeper[axelarnetKeeper.Keeper](keepers)
	govK.SetHooks(govtypes.NewMultiGovHooks(axelarnetK.Hooks(getKeeper[nexusKeeper.Keeper](keepers), govK)))
	return &govK
}

func initPermissionKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *permissionKeeper.Keeper {
	permissionK := permissionKeeper.NewKeeper(appCodec, keys[permissionTypes.StoreKey], keepers.getSubspace(permissionTypes.ModuleName))
	return &permissionK
}

func initVoteKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *voteKeeper.Keeper {
	voteRouter := voteTypes.NewRouter()
	voteRouter.AddHandler(
		evmTypes.ModuleName,
		evmKeeper.NewVoteHandler(
			appCodec,
			getKeeper[evmKeeper.BaseKeeper](keepers),
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
	return &voteK
}

func initSnapshotKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *snapKeeper.Keeper {
	snapK := snapKeeper.NewKeeper(
		appCodec,
		keys[snapTypes.StoreKey],
		keepers.getSubspace(snapTypes.ModuleName),
		getKeeper[stakingkeeper.Keeper](keepers),
		axelarbankkeeper.NewBankKeeper(getKeeper[bankkeeper.BaseKeeper](keepers)),
		getKeeper[slashingkeeper.Keeper](keepers),
	)
	return &snapK
}

func initTssKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *tssKeeper.Keeper {
	tssK := tssKeeper.NewKeeper(appCodec, keys[tssTypes.StoreKey], keepers.getSubspace(tssTypes.ModuleName))
	return &tssK
}

func initMultisigKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *multisigKeeper.Keeper {
	multisigRouter := multisigTypes.NewSigRouter()
	multisigRouter.AddHandler(evmTypes.ModuleName, evmKeeper.NewSigHandler(appCodec, getKeeper[evmKeeper.BaseKeeper](keepers)))

	multisigK := multisigKeeper.NewKeeper(appCodec, keys[multisigTypes.StoreKey], keepers.getSubspace(multisigTypes.ModuleName))
	multisigK.SetSigRouter(multisigRouter)
	return &multisigK
}

func initRewardKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *rewardKeeper.Keeper {
	rewardK := rewardKeeper.NewKeeper(
		appCodec,
		keys[rewardTypes.StoreKey],
		keepers.getSubspace(rewardTypes.ModuleName),
		axelarbankkeeper.NewBankKeeper(getKeeper[bankkeeper.BaseKeeper](keepers)),
		getKeeper[distrkeeper.Keeper](keepers),
		getKeeper[stakingkeeper.Keeper](keepers),
	)
	return &rewardK
}

func initIBCKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *ibckeeper.Keeper {
	scopedIBCK := getKeeper[capabilitykeeper.Keeper](keepers).ScopeToModule(ibchost.ModuleName)
	return ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		keepers.getSubspace(ibchost.ModuleName),
		getKeeper[stakingkeeper.Keeper](keepers),
		getKeeper[upgradekeeper.Keeper](keepers),
		scopedIBCK,
	)
}

func initIBCTransferKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, ics4Wrapper ibctransfertypes.ICS4Wrapper) *ibctransferkeeper.Keeper {
	scopedTransferK := getKeeper[capabilitykeeper.Keeper](keepers).ScopeToModule(ibctransfertypes.ModuleName)
	transferK := ibctransferkeeper.NewKeeper(
		appCodec, keys[ibctransfertypes.StoreKey], keepers.getSubspace(ibctransfertypes.ModuleName),
		// Use the IBC middleware stack
		ics4Wrapper,
		getKeeper[ibckeeper.Keeper](keepers).ChannelKeeper, &getKeeper[ibckeeper.Keeper](keepers).PortKeeper,
		getKeeper[authkeeper.AccountKeeper](keepers), getKeeper[bankkeeper.BaseKeeper](keepers), scopedTransferK,
	)
	return &transferK
}

func initWasmContractKeeper(keepers *keeperCache) *wasmkeeper.PermissionedKeeper {
	return wasmkeeper.NewDefaultPermissionKeeper(getKeeper[wasm.Keeper](keepers))
}

func initAxelarIBCKeeper(keepers *keeperCache) *axelarnetKeeper.IBCKeeper {
	ibcK := axelarnetKeeper.NewIBCKeeper(*getKeeper[axelarnetKeeper.Keeper](keepers), getKeeper[ibctransferkeeper.Keeper](keepers))
	return &ibcK
}

func initAxelarnetKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *axelarnetKeeper.Keeper {
	axelarnetK := axelarnetKeeper.NewKeeper(
		appCodec,
		keys[axelarnetTypes.StoreKey],
		keepers.getSubspace(axelarnetTypes.ModuleName),
		getKeeper[ibckeeper.Keeper](keepers).ChannelKeeper,
		getKeeper[feegrantkeeper.Keeper](keepers),
	)
	return &axelarnetK
}

func initEvmKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *evmKeeper.BaseKeeper {
	return evmKeeper.NewKeeper(appCodec, keys[evmTypes.StoreKey], getKeeper[paramskeeper.Keeper](keepers))
}

func initNexusKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *nexusKeeper.Keeper {
	// setting validator will finalize all by sealing it
	// no more validators can be added
	addressValidators := nexusTypes.NewAddressValidators().
		AddAddressValidator(evmTypes.ModuleName, evmKeeper.NewAddressValidator()).
		AddAddressValidator(axelarnetTypes.ModuleName, axelarnetKeeper.NewAddressValidator(getKeeper[axelarnetKeeper.Keeper](keepers)))
	addressValidators.Seal()

	nexusK := nexusKeeper.NewKeeper(appCodec, keys[nexusTypes.StoreKey], keepers.getSubspace(nexusTypes.ModuleName))
	nexusK.SetAddressValidators(addressValidators)

	return &nexusK
}

func initCapabilityKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, memKeys map[string]*sdk.MemoryStoreKey) *capabilitykeeper.Keeper {
	return capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
}

func initFeegrantKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *feegrantkeeper.Keeper {
	feegrantK := feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], getKeeper[authkeeper.AccountKeeper](keepers))
	return &feegrantK
}

func initEvidenceKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *evidencekeeper.Keeper {
	return evidencekeeper.NewKeeper(appCodec, keys[evidencetypes.StoreKey], getKeeper[stakingkeeper.Keeper](keepers), getKeeper[slashingkeeper.Keeper](keepers))
}

func initUpgradeKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, skipUpgradeHeights map[int64]bool, homePath string, bApp *bam.BaseApp) *upgradekeeper.Keeper {
	upgradeK := upgradekeeper.NewKeeper(skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath, bApp)
	return &upgradeK
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

func initCrisisKeeper(keepers *keeperCache, invCheckPeriod uint) *crisiskeeper.Keeper {
	crisisK := crisiskeeper.NewKeeper(
		keepers.getSubspace(crisistypes.ModuleName),
		invCheckPeriod,
		getKeeper[bankkeeper.BaseKeeper](keepers),
		authtypes.FeeCollectorName,
	)
	return &crisisK
}

func initSlashingKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *slashingkeeper.Keeper {
	slashK := slashingkeeper.NewKeeper(appCodec, keys[slashingtypes.StoreKey], getKeeper[stakingkeeper.Keeper](keepers), keepers.getSubspace(slashingtypes.ModuleName))
	return &slashK
}

func initDistributionKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, moduleAccPerms map[string][]string) *distrkeeper.Keeper {
	distrK := distrkeeper.NewKeeper(
		appCodec,
		keys[distrtypes.StoreKey],
		keepers.getSubspace(distrtypes.ModuleName),
		getKeeper[authkeeper.AccountKeeper](keepers),
		getKeeper[bankkeeper.BaseKeeper](keepers),
		getKeeper[stakingkeeper.Keeper](keepers),
		authtypes.FeeCollectorName,
		moduleAccountAddrs(moduleAccPerms),
	)
	return &distrK
}

func initMintKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache) *mintkeeper.Keeper {
	mintK := mintkeeper.NewKeeper(
		appCodec,
		keys[minttypes.StoreKey],
		keepers.getSubspace(minttypes.ModuleName),
		getKeeper[stakingkeeper.Keeper](keepers),
		getKeeper[authkeeper.AccountKeeper](keepers),
		getKeeper[bankkeeper.BaseKeeper](keepers),
		authtypes.FeeCollectorName,
	)
	return &mintK
}

func initBankKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, moduleAccPerms map[string][]string) *bankkeeper.BaseKeeper {
	bankK := bankkeeper.NewBaseKeeper(
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
	return &bankK
}

func initAccountKeeper(appCodec codec.Codec, keys map[string]*sdk.KVStoreKey, keepers *keeperCache, moduleAccPerms map[string][]string) *authkeeper.AccountKeeper {
	authK := authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		keepers.getSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		moduleAccPerms,
	)

	return &authK
}

// moduleAccountAddrs returns all the app's module account addresses.
func moduleAccountAddrs(moduleAccPerms map[string][]string) map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range moduleAccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}
