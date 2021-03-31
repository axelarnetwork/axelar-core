package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	"github.com/spf13/viper"

	"github.com/axelarnetwork/axelar-core/x/ante"
	snapshotExported "github.com/axelarnetwork/axelar-core/x/snapshot/exported"
	snapshotExportedMock "github.com/axelarnetwork/axelar-core/x/snapshot/exported/mock"
	snapTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/bank"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/rpc/client/http"
	dbm "github.com/tendermint/tm-db"

	"github.com/axelarnetwork/axelar-core/x/nexus"

	keyring "github.com/cosmos/cosmos-sdk/crypto/keys"

	"github.com/axelarnetwork/axelar-core/x/bitcoin"
	btcKeeper "github.com/axelarnetwork/axelar-core/x/bitcoin/keeper"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	"github.com/axelarnetwork/axelar-core/x/broadcast"
	broadcastKeeper "github.com/axelarnetwork/axelar-core/x/broadcast/keeper"
	broadcastTypes "github.com/axelarnetwork/axelar-core/x/broadcast/types"
	"github.com/axelarnetwork/axelar-core/x/ethereum"
	ethKeeper "github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	nexusKeeper "github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	nexusTypes "github.com/axelarnetwork/axelar-core/x/nexus/types"
	"github.com/axelarnetwork/axelar-core/x/snapshot"
	snapKeeper "github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss"
	tssKeeper "github.com/axelarnetwork/axelar-core/x/tss/keeper"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	"github.com/axelarnetwork/axelar-core/x/vote"
	voteKeeper "github.com/axelarnetwork/axelar-core/x/vote/keeper"
	voteTypes "github.com/axelarnetwork/axelar-core/x/vote/types"
)

const (
	appName          = "axelar"
	bech32MainPrefix = "axelar"

	// Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
	Bech32PrefixAccAddr = bech32MainPrefix
	// Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
	Bech32PrefixAccPub = bech32MainPrefix + sdk.PrefixPublic
	// Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
	Bech32PrefixValAddr = bech32MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator
	// Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
	Bech32PrefixValPub = bech32MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic
	// Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
	Bech32PrefixConsAddr = bech32MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus
	// Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
	Bech32PrefixConsPub = bech32MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic
)

var (
	// DefaultCLIHome sets the default home directories for the application CLI
	DefaultCLIHome = os.ExpandEnv("$HOME/.axelarcli")

	// DefaultNodeHome sets the folder where the applcation data and configuration will be stored
	DefaultNodeHome = os.ExpandEnv("$HOME/.axelard")

	// ModuleBasics is in charge of setting up basic module elements
	ModuleBasics = module.NewBasicManager(
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		distr.AppModuleBasic{},
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		supply.AppModuleBasic{},

		tss.AppModuleBasic{},
		vote.AppModuleBasic{},
		bitcoin.AppModuleBasic{},
		ethereum.AppModuleBasic{},
		broadcast.AppModuleBasic{},
		snapshot.AppModuleBasic{},
		nexus.AppModuleBasic{},
	)
	// account permissions
	maccPerms = map[string][]string{
		auth.FeeCollectorName:     nil,
		distr.ModuleName:          nil,
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
	}
)

// MakeCodec generates the necessary codecs for Amino
func MakeCodec() *codec.Codec {
	var cdc = codec.New()

	ModuleBasics.RegisterCodec(cdc)
	vesting.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	cdc = cdc.Seal()

	return cdc
}

// AxelarApp defines the axelar Cosmos app that runs all modules
type AxelarApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint

	// keys to access the substores
	keys  map[string]*sdk.KVStoreKey
	tkeys map[string]*sdk.TransientStoreKey

	// Keepers
	stakingKeeper  staking.Keeper
	slashingKeeper slashing.Keeper
	distrKeeper    distr.Keeper

	// Module Manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager
}

// verify app interface at compile time
var _ simapp.App = &AxelarApp{}

// NewInitApp is a constructor function for axelarApp
func NewInitApp(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, axelarCfg Config, baseAppOptions ...func(*bam.BaseApp)) *AxelarApp {

	// First define the top level codec that will be shared by the different modules
	cdc := MakeCodec()

	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetAppVersion(version.Version)

	keys := sdk.NewKVStoreKeys(
		bam.MainStoreKey,
		auth.StoreKey,
		staking.StoreKey,
		supply.StoreKey,
		distr.StoreKey,
		slashing.StoreKey,
		params.StoreKey,
		voteTypes.StoreKey,
		broadcastTypes.StoreKey,
		btcTypes.StoreKey,
		ethTypes.StoreKey,
		snapTypes.StoreKey,
		tssTypes.StoreKey,
		nexusTypes.StoreKey,
	)

	tkeys := sdk.NewTransientStoreKeys(staking.TStoreKey, params.TStoreKey)

	// Here you initialize your application with the store keys it requires
	var app = &AxelarApp{
		BaseApp:        bApp,
		cdc:            cdc,
		keys:           keys,
		tkeys:          tkeys,
		invCheckPeriod: invCheckPeriod,
	}

	// The ParamsKeeper handles parameter storage for the application
	paramsK := params.NewKeeper(app.cdc, keys[params.StoreKey], tkeys[params.TStoreKey])
	// Set specific subspaces
	authSubspace := paramsK.Subspace(auth.DefaultParamspace)
	bankSubspace := paramsK.Subspace(bank.DefaultParamspace)
	stakingSubspace := paramsK.Subspace(staking.DefaultParamspace)
	distrSubspace := paramsK.Subspace(distr.DefaultParamspace)
	slashingSubspace := paramsK.Subspace(slashing.DefaultParamspace)
	snapshotSubspace := paramsK.Subspace(snapTypes.DefaultParamspace)
	tssSubspace := paramsK.Subspace(tssTypes.DefaultParamspace)
	btcSubspace := paramsK.Subspace(btcTypes.DefaultParamspace)
	ethSubspace := paramsK.Subspace(ethTypes.DefaultParamspace)
	nexusSubspace := paramsK.Subspace(nexusTypes.DefaultParamspace)

	// The AccountKeeper handles address -> account lookups
	accountK := auth.NewAccountKeeper(
		app.cdc,
		keys[auth.StoreKey],
		authSubspace,
		auth.ProtoBaseAccount,
	)
	// The BankKeeper allows you perform sdk.Coins interactions
	bankK := bank.NewBaseKeeper(
		accountK,
		bankSubspace,
		app.ModuleAccountAddrs(),
	)
	// The SupplyKeeper collects transaction fees and renders them to the fee distribution module
	supplyK := supply.NewKeeper(
		app.cdc,
		keys[supply.StoreKey],
		accountK,
		bankK,
		maccPerms,
	)
	stakingK := staking.NewKeeper(
		app.cdc,
		keys[staking.StoreKey],
		supplyK,
		stakingSubspace,
	)
	distrK := distr.NewKeeper(
		app.cdc,
		keys[distr.StoreKey],
		distrSubspace,
		&stakingK,
		supplyK,
		auth.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)
	slashingK := slashing.NewKeeper(
		app.cdc,
		keys[slashing.StoreKey],
		&stakingK,
		slashingSubspace,
	)
	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	stakingK = *stakingK.SetHooks(
		staking.NewMultiStakingHooks(
			distrK.Hooks(),
			slashingK.Hooks()),
	)
	btcK := btcKeeper.NewKeeper(
		app.cdc,
		keys[btcTypes.StoreKey],
		btcSubspace,
	)
	ethK := ethKeeper.NewEthKeeper(
		app.cdc,
		keys[ethTypes.StoreKey],
		ethSubspace,
	)

	keybase, err := keyring.NewKeyring(sdk.KeyringServiceName(), axelarCfg.ClientConfig.KeyringBackend, viper.GetString("clihome"), os.Stdin)
	if err != nil {
		tmos.Exit(err.Error())
	}
	abciClient, err := http.New(axelarCfg.TendermintNodeUri, "/websocket")
	if err != nil {
		tmos.Exit(err.Error())
	}
	broadcastK, err := broadcastKeeper.NewKeeper(
		app.cdc,
		keys[broadcastTypes.StoreKey],
		dbadapter.Store{DB: dbm.NewMemDB()},
		keybase,
		accountK,
		stakingK,
		abciClient,
		axelarCfg.ClientConfig,
		logger,
	)
	if err != nil {
		tmos.Exit(err.Error())
	}

	slashingKCast := &snapshotExportedMock.SlasherMock{
		GetValidatorSigningInfoFunc: func(ctx sdk.Context, address sdk.ConsAddress) (snapshotExported.ValidatorInfo, bool) {
			signingInfo, found := slashingK.GetValidatorSigningInfo(ctx, address)

			return snapshotExported.ValidatorInfo{ValidatorSigningInfo: signingInfo}, found
		},
	}
	tssK := tssKeeper.NewKeeper(
		app.cdc,
		keys[tssTypes.StoreKey],
		tssSubspace,
		broadcastK,
		slashingKCast,
	)
	snapK := snapKeeper.NewKeeper(
		app.cdc,
		keys[snapTypes.StoreKey],
		snapshotSubspace,
		broadcastK,
		stakingK,
		slashingKCast,
		tssK,
	)
	nexusK := nexusKeeper.NewKeeper(
		app.cdc,
		keys[nexusTypes.StoreKey],
		nexusSubspace,
	)
	votingK := voteKeeper.NewKeeper(
		app.cdc,
		keys[voteTypes.StoreKey],
		dbadapter.Store{DB: dbm.NewMemDB()},
		snapK,
		broadcastK,
	)

	app.stakingKeeper = stakingK
	app.distrKeeper = distrK
	app.slashingKeeper = slashingK

	var rpcEth ethTypes.RPCClient
	if axelarCfg.WithEthBridge {
		rpcEth, err = ethTypes.NewRPCClient(axelarCfg.EthRpcAddr)
		if err != nil {
			tmos.Exit(err.Error())
		}
		logger.With("module", fmt.Sprintf("x/%s", ethTypes.ModuleName)).Debug("Successfully connected to ethereum node")
	} else {
		rpcEth = ethTypes.NewDummyRPC()
	}

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		genutil.NewAppModule(accountK, stakingK, app.BaseApp.DeliverTx),
		auth.NewAppModule(accountK),
		bank.NewAppModule(bankK, accountK),
		supply.NewAppModule(supplyK, accountK),
		distr.NewAppModule(distrK, accountK, supplyK, stakingK),
		slashing.NewAppModule(slashingK, accountK, stakingK),
		staking.NewAppModule(stakingK, accountK, supplyK),

		snapshot.NewAppModule(snapK),
		tss.NewAppModule(tssK, snapK, votingK, nexusK, stakingK),
		vote.NewAppModule(votingK),
		broadcast.NewAppModule(broadcastK),
		nexus.NewAppModule(nexusK),
		ethereum.NewAppModule(ethK, votingK, tssK, nexusK, snapK, rpcEth),
		bitcoin.NewAppModule(btcK, votingK, tssK, nexusK, snapK),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	app.mm.SetOrderBeginBlockers(distr.ModuleName, slashing.ModuleName)
	app.mm.SetOrderEndBlockers(staking.ModuleName, voteTypes.ModuleName, btcTypes.ModuleName)

	// Sets the order of Genesis - Order matters, genutil is to always come last
	// NOTE: The genutils moodule must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(
		distr.ModuleName,
		staking.ModuleName,
		auth.ModuleName,
		bank.ModuleName,
		slashing.ModuleName,
		snapTypes.ModuleName,
		tssTypes.ModuleName,
		btcTypes.ModuleName,
		ethTypes.ModuleName,
		nexusTypes.ModuleName,
		broadcastTypes.ModuleName,
		voteTypes.ModuleName,
		supply.ModuleName,
		genutil.ModuleName,
	)

	// register all module routes and module queriers
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// The initChainer handles translating the genesis.json file into initial state for the network
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	// The baseAnteHandler handles signature verification and transaction pre-processing
	baseAnteHandler := auth.NewAnteHandler(
		accountK,
		supplyK,
		auth.DefaultSigVerificationGasConsumer,
	)
	anteHandler := sdk.ChainAnteDecorators(
		ante.NewAnteHandlerDecorator(baseAnteHandler),
		ante.NewValidateValidatorDeregisteredTssDecorator(tssK, nexusK, snapK),
	)
	app.SetAnteHandler(anteHandler)

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)

	if loadLatest {
		err := app.LoadLatestVersion(app.keys[bam.MainStoreKey])
		if err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

// GenesisState represents chain state at the start of the chain. Any initial state (account balances) are stored here.
type GenesisState map[string]json.RawMessage

// InitChainer handles the chain initialization from a genesis file
func (app *AxelarApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState

	err := app.cdc.UnmarshalJSON(req.AppStateBytes, &genesisState)
	if err != nil {
		panic(err)
	}

	return app.mm.InitGenesis(ctx, genesisState)
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
	return app.LoadVersion(height, app.keys[bam.MainStoreKey])
}

// Codec returns simapp's codec
func (app *AxelarApp) Codec() *codec.Codec {
	return app.cdc
}

// SimulationManager implements the SimulationApp interface
func (app *AxelarApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *AxelarApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}
