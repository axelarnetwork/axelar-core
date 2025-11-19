package app

import (
	"fmt"
	"reflect"
	"strings"

	"cosmossdk.io/log"
	store "cosmossdk.io/store/types"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v8/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	"golang.org/x/mod/semver"

	axelarParams "github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/x/axelarnet"
	axelarnetKeeper "github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	axelarnetTypes "github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	axelarbankkeeper "github.com/axelarnetwork/axelar-core/x/bank/keeper"
	axelardistrkeeper "github.com/axelarnetwork/axelar-core/x/distribution/keeper"
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

var GovModuleAddress = authtypes.NewModuleAddress(govtypes.ModuleName)

type KeeperCache struct {
	repository map[string]any
}

func NewKeeperCache() *KeeperCache {
	return &KeeperCache{
		repository: make(map[string]any),
	}
}

func (k *KeeperCache) getSubspace(moduleName string) paramstypes.Subspace {
	paramsK := GetKeeper[paramskeeper.Keeper](k)
	subspace, ok := paramsK.GetSubspace(moduleName)
	if !ok {
		panic(fmt.Sprintf("subspace %s not found", moduleName))
	}
	return subspace
}

func GetKeeper[T any](k *KeeperCache) *T {
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

func SetKeeper[T any](k *KeeperCache, keeper T) {
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

func initParamsKeeper(encodingConfig axelarParams.EncodingConfig, key, tkey store.StoreKey) *paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(encodingConfig.Codec, encodingConfig.Amino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
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

func initStakingKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *stakingkeeper.Keeper {
	return stakingkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[stakingtypes.StoreKey]),
		GetKeeper[authkeeper.AccountKeeper](keepers),
		GetKeeper[bankkeeper.BaseKeeper](keepers),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)
}

func initWasmKeeper(encodingConfig axelarParams.EncodingConfig, keys map[string]*store.KVStoreKey, keepers *KeeperCache, bApp *bam.BaseApp, appOpts types.AppOptions, wasmOpts []wasm.Option, wasmDir string) *wasm.Keeper {
	wasmConfig := mustReadWasmConfig(appOpts)
	nexusK := GetKeeper[nexusKeeper.Keeper](keepers)

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	gasRegisterConfig := wasmtypes.DefaultGasRegisterConfig()
	// 1 SDK gas = 2500 CWgas
	// given our block gas limit of 400000000, this means ~1 second of execution time.
	gasRegisterConfig.GasMultiplier = 2500
	wasmOpts = append(
		wasmOpts,
		wasmkeeper.WithMessageHandlerDecorator(
			func(old wasmkeeper.Messenger) wasmkeeper.Messenger {
				encoders := wasm.DefaultEncoders(encodingConfig.Codec, GetKeeper[ibctransferkeeper.Keeper](keepers))
				encoders.Custom = nexusKeeper.EncodeRoutingMessage

				return WithAnteHandlers(
					encoders,
					initMessageAnteDecorators(encodingConfig, keepers),
					// for security reasons we disallow some msg types that can be used for arbitrary calls
					wasmkeeper.NewMessageHandlerChain(NewMsgTypeBlacklistMessenger(), old, nexusKeeper.NewMessenger(nexusK)))
			}),
		wasmkeeper.WithWasmEngineDecorator(func(old wasmtypes.WasmEngine) wasmtypes.WasmEngine {
			return nexusKeeper.NewWasmerEngine(old, nexusK)
		}),
		wasmkeeper.WithQueryPlugins(NewQueryPlugins(nexusK)),
		wasmkeeper.WithGasRegister(wasmtypes.NewWasmGasRegister(gasRegisterConfig)),
	)

	scopedWasmK := GetKeeper[capabilitykeeper.Keeper](keepers).ScopeToModule(wasm.ModuleName)
	ibcKeeper := GetKeeper[ibckeeper.Keeper](keepers)

	wasmK := wasm.NewKeeper(
		encodingConfig.Codec,
		runtime.NewKVStoreService(keys[wasm.StoreKey]),
		GetKeeper[authkeeper.AccountKeeper](keepers),
		GetKeeper[bankkeeper.BaseKeeper](keepers),
		GetKeeper[stakingkeeper.Keeper](keepers),
		distrkeeper.NewQuerier(*GetKeeper[distrkeeper.Keeper](keepers)),
		ibcKeeper.ChannelKeeper,
		ibcKeeper.ChannelKeeper,
		ibcKeeper.PortKeeper,
		scopedWasmK,
		GetKeeper[ibctransferkeeper.Keeper](keepers),
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		wasmtypes.VMConfig{},
		wasmkeeper.BuiltInCapabilities(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	return &wasmK
}

func initGovernanceKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache, msgServiceRouter *bam.MsgServiceRouter) *govkeeper.Keeper {
	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(*GetKeeper[paramskeeper.Keeper](keepers))).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(GetKeeper[ibckeeper.Keeper](keepers).ClientKeeper)).
		AddRoute(axelarnetTypes.RouterKey, axelarnet.NewProposalHandler(*GetKeeper[axelarnetKeeper.Keeper](keepers), GetKeeper[nexusKeeper.Keeper](keepers), GetKeeper[authkeeper.AccountKeeper](keepers)))

	govK := govkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[govtypes.StoreKey]),
		GetKeeper[authkeeper.AccountKeeper](keepers),
		GetKeeper[bankkeeper.BaseKeeper](keepers),
		GetKeeper[stakingkeeper.Keeper](keepers),
		GetKeeper[distrkeeper.Keeper](keepers),
		msgServiceRouter,
		govtypes.DefaultConfig(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Set legacy router for backwards compatibility with gov v1beta1
	govK.SetLegacyRouter(govRouter)

	axelarnetK := GetKeeper[axelarnetKeeper.Keeper](keepers)
	govK.SetHooks(govtypes.NewMultiGovHooks(axelarnetK.Hooks(GetKeeper[nexusKeeper.Keeper](keepers), *govK)))

	return govK
}

func initPermissionKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *permissionKeeper.Keeper {
	permissionK := permissionKeeper.NewKeeper(appCodec, keys[permissionTypes.StoreKey], keepers.getSubspace(permissionTypes.ModuleName))
	return &permissionK
}

func initVoteKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *voteKeeper.Keeper {
	voteRouter := voteTypes.NewRouter()
	voteRouter.AddHandler(
		evmTypes.ModuleName,
		evmKeeper.NewVoteHandler(
			appCodec,
			GetKeeper[evmKeeper.BaseKeeper](keepers),
			GetKeeper[nexusKeeper.Keeper](keepers),
			GetKeeper[rewardKeeper.Keeper](keepers),
		),
	)

	voteK := voteKeeper.NewKeeper(
		appCodec,
		keys[voteTypes.StoreKey],
		keepers.getSubspace(voteTypes.ModuleName),
		GetKeeper[snapKeeper.Keeper](keepers),
		GetKeeper[stakingkeeper.Keeper](keepers),
		GetKeeper[rewardKeeper.Keeper](keepers),
	)
	voteK.SetVoteRouter(voteRouter)
	return &voteK
}

func initSnapshotKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *snapKeeper.Keeper {
	snapK := snapKeeper.NewKeeper(
		appCodec,
		keys[snapTypes.StoreKey],
		keepers.getSubspace(snapTypes.ModuleName),
		GetKeeper[stakingkeeper.Keeper](keepers),
		axelarbankkeeper.NewBankKeeper(GetKeeper[bankkeeper.BaseKeeper](keepers)),
		GetKeeper[slashingkeeper.Keeper](keepers),
	)
	return &snapK
}

func initTssKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *tssKeeper.Keeper {
	tssK := tssKeeper.NewKeeper(appCodec, keys[tssTypes.StoreKey], keepers.getSubspace(tssTypes.ModuleName))
	return &tssK
}

func initMultisigKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *multisigKeeper.Keeper {
	multisigRouter := multisigTypes.NewSigRouter()
	multisigRouter.AddHandler(evmTypes.ModuleName, evmKeeper.NewSigHandler(appCodec, GetKeeper[evmKeeper.BaseKeeper](keepers)))

	multisigK := multisigKeeper.NewKeeper(appCodec, keys[multisigTypes.StoreKey], keepers.getSubspace(multisigTypes.ModuleName))
	multisigK.SetSigRouter(multisigRouter)
	return &multisigK
}

func initRewardKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *rewardKeeper.Keeper {
	rewardK := rewardKeeper.NewKeeper(
		appCodec,
		keys[rewardTypes.StoreKey],
		keepers.getSubspace(rewardTypes.ModuleName),
		axelarbankkeeper.NewBankKeeper(GetKeeper[bankkeeper.BaseKeeper](keepers)),
		GetKeeper[distrkeeper.Keeper](keepers),
		GetKeeper[stakingkeeper.Keeper](keepers),
	)
	return &rewardK
}

func initIBCKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *ibckeeper.Keeper {
	scopedIBCK := GetKeeper[capabilitykeeper.Keeper](keepers).ScopeToModule(ibcexported.ModuleName)
	return ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		keepers.getSubspace(ibcexported.ModuleName),
		GetKeeper[stakingkeeper.Keeper](keepers),
		GetKeeper[upgradekeeper.Keeper](keepers),
		scopedIBCK,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
}

func initIBCTransferKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache, ics4Wrapper porttypes.ICS4Wrapper) *ibctransferkeeper.Keeper {
	scopedTransferK := GetKeeper[capabilitykeeper.Keeper](keepers).ScopeToModule(ibctransfertypes.ModuleName)
	transferK := ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		keepers.getSubspace(ibctransfertypes.ModuleName),
		// Use the IBC middleware stack
		ics4Wrapper,
		GetKeeper[ibckeeper.Keeper](keepers).ChannelKeeper,
		GetKeeper[ibckeeper.Keeper](keepers).PortKeeper,
		GetKeeper[authkeeper.AccountKeeper](keepers),
		GetKeeper[bankkeeper.BaseKeeper](keepers),
		scopedTransferK,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	return &transferK
}

func initWasmContractKeeper(keepers *KeeperCache) *wasmkeeper.PermissionedKeeper {
	return wasmkeeper.NewDefaultPermissionKeeper(GetKeeper[wasm.Keeper](keepers))
}

func initAxelarIBCKeeper(keepers *KeeperCache) *axelarnetKeeper.IBCKeeper {
	ibcK := axelarnetKeeper.NewIBCKeeper(*GetKeeper[axelarnetKeeper.Keeper](keepers), GetKeeper[ibctransferkeeper.Keeper](keepers))
	return &ibcK
}

func initAxelarnetKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *axelarnetKeeper.Keeper {
	axelarnetK := axelarnetKeeper.NewKeeper(
		appCodec,
		keys[axelarnetTypes.StoreKey],
		keepers.getSubspace(axelarnetTypes.ModuleName),
		GetKeeper[ibckeeper.Keeper](keepers).ChannelKeeper,
		GetKeeper[feegrantkeeper.Keeper](keepers),
	)
	return &axelarnetK
}

func initEvmKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *evmKeeper.BaseKeeper {
	return evmKeeper.NewKeeper(appCodec, keys[evmTypes.StoreKey], GetKeeper[paramskeeper.Keeper](keepers))
}

func initNexusKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *nexusKeeper.Keeper {
	// setting validator will finalize all by sealing it
	// no more validators can be added
	addressValidators := nexusTypes.NewAddressValidators().
		AddAddressValidator(evmTypes.ModuleName, evmKeeper.NewAddressValidator()).
		AddAddressValidator(axelarnetTypes.ModuleName, axelarnetKeeper.NewAddressValidator(GetKeeper[axelarnetKeeper.Keeper](keepers)))
	addressValidators.Seal()

	nexusK := nexusKeeper.NewKeeper(appCodec, keys[nexusTypes.StoreKey], keepers.getSubspace(nexusTypes.ModuleName))
	nexusK.SetAddressValidators(addressValidators)

	return &nexusK
}

func initCapabilityKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, memKeys map[string]*store.MemoryStoreKey) *capabilitykeeper.Keeper {
	return capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])
}

func initFeegrantKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *feegrantkeeper.Keeper {
	feegrantK := feegrantkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[feegrant.StoreKey]), GetKeeper[authkeeper.AccountKeeper](keepers))
	return &feegrantK
}

func initEvidenceKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *evidencekeeper.Keeper {
	return evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),
		GetKeeper[stakingkeeper.Keeper](keepers),
		GetKeeper[slashingkeeper.Keeper](keepers),
		GetKeeper[authkeeper.AccountKeeper](keepers).AddressCodec(),
		runtime.ProvideCometInfoService(),
	)
}

func initUpgradeKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, skipUpgradeHeights map[int64]bool, homePath string, bApp *bam.BaseApp) *upgradekeeper.Keeper {
	return upgradekeeper.NewKeeper(skipUpgradeHeights, runtime.NewKVStoreService(keys[upgradetypes.StoreKey]), appCodec, homePath, bApp, authtypes.NewModuleAddress(govtypes.ModuleName).String())
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

func initCrisisKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache, invCheckPeriod uint) *crisiskeeper.Keeper {
	return crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
		invCheckPeriod,
		GetKeeper[bankkeeper.BaseKeeper](keepers),
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		GetKeeper[authkeeper.AccountKeeper](keepers).AddressCodec(),
	)
}

func initSlashingKeeper(appCodec codec.Codec, legacyAmino *codec.LegacyAmino, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *slashingkeeper.Keeper {
	slashK := slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		runtime.NewKVStoreService(keys[slashingtypes.StoreKey]),
		GetKeeper[stakingkeeper.Keeper](keepers),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	return &slashK
}

func initDistributionKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *distrkeeper.Keeper {
	distrK := distrkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[distrtypes.StoreKey]),
		GetKeeper[authkeeper.AccountKeeper](keepers),
		GetKeeper[bankkeeper.BaseKeeper](keepers),
		GetKeeper[stakingkeeper.Keeper](keepers),
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	return &distrK
}

func initAxelarDistributionKeeper(keepers *KeeperCache) *axelardistrkeeper.Keeper {
	axelardistrK := axelardistrkeeper.NewKeeper(
		*GetKeeper[distrkeeper.Keeper](keepers),
		GetKeeper[authkeeper.AccountKeeper](keepers),
		GetKeeper[bankkeeper.BaseKeeper](keepers),
		GetKeeper[stakingkeeper.Keeper](keepers),
		authtypes.FeeCollectorName,
	)

	return &axelardistrK
}

func initMintKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache) *mintkeeper.Keeper {
	mintK := mintkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[minttypes.StoreKey]),
		GetKeeper[stakingkeeper.Keeper](keepers),
		GetKeeper[authkeeper.AccountKeeper](keepers),
		GetKeeper[bankkeeper.BaseKeeper](keepers),
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	return &mintK
}

func initBankKeeper(logger log.Logger, appCodec codec.Codec, keys map[string]*store.KVStoreKey, keepers *KeeperCache, moduleAccPerms map[string][]string) *bankkeeper.BaseKeeper {
	bankK := bankkeeper.NewBaseKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		GetKeeper[authkeeper.AccountKeeper](keepers),
		maps.Filter(moduleAccountAddrs(moduleAccPerms), func(addr string, _ bool) bool {
			// we do not rely on internal balance tracking for invariance checks in the nexus module
			// (https://github.com/cosmos/cosmos-sdk/issues/12825 for more details on the purpose of the blocked list),
			// but the nexus module account must be able to send or receive coins to mint/burn them for cross-chain transfers,
			// so we exclude this address from the blocked list
			return addr != authtypes.NewModuleAddress(nexusTypes.ModuleName).String()
		}),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		logger,
	)
	return &bankK
}

func initAccountKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey, moduleAccPerms map[string][]string) *authkeeper.AccountKeeper {
	authK := authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		moduleAccPerms,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	return &authK
}

func initConsensusParamsKeeper(appCodec codec.Codec, keys map[string]*store.KVStoreKey) *consensusparamkeeper.Keeper {
	consensusparamK := consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		runtime.EventService{},
	)

	return &consensusparamK
}

// moduleAccountAddrs returns all the app's module account addresses.
func moduleAccountAddrs(moduleAccPerms map[string][]string) map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range moduleAccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}
