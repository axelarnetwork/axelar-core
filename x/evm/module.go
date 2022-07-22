package evm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/client/cli"
	"github.com/axelarnetwork/axelar-core/x/evm/client/rest"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic implements module.AppModuleBasic
type AppModuleBasic struct {
}

// Name returns the name of the module
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the types necessary in this module with the given codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns the default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	state := types.DefaultGenesisState()
	return cdc.MustMarshalJSON(&state)
}

// ValidateGenesis checks the given genesis state for validity
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterRESTRoutes registers the REST routes for this module
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(clientCtx, rtr)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryServiceHandlerClient(context.Background(), mux, types.NewQueryServiceClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns all CLI tx commands for this module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns all CLI query commands for this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(types.QuerierRoute)
}

// AppModule implements module.AppModule
type AppModule struct {
	AppModuleBasic
	logger      log.Logger
	keeper      keeper.BaseKeeper
	voter       types.Voter
	nexus       types.Nexus
	signer      types.Signer
	snapshotter types.Snapshotter
	staking     types.StakingKeeper
	slashing    types.SlashingKeeper
	multisig    types.MultisigKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	k keeper.BaseKeeper,
	voter types.Voter,
	signer types.Signer,
	nexus types.Nexus,
	snapshotter types.Snapshotter,
	staking types.StakingKeeper,
	slashing types.SlashingKeeper,
	multisig types.MultisigKeeper,
	logger log.Logger) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		logger:         logger,
		keeper:         k,
		voter:          voter,
		nexus:          nexus,
		snapshotter:    snapshotter,
		staking:        staking,
		slashing:       slashing,
		multisig:       multisig,
		signer:         signer,
	}
}

// RegisterInvariants registers this module's invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {
	// No invariants yet
}

// InitGenesis initializes the module's keeper from the given genesis state
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)
	am.keeper.InitGenesis(ctx, genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports a genesis state from the module's keeper
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(&genState)
}

// Route returns the module's route
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.keeper, am.voter, am.nexus, am.snapshotter, am.staking, am.slashing, am.multisig))
}

// QuerierRoute returns this module's query route
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// LegacyQuerierHandler returns a new query handler for this module
func (am AppModule) LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier {
	return keeper.NewQuerier(am.keeper, am.nexus)
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServiceServer(cfg.QueryServer(), keeper.NewGRPCQuerier(am.keeper, am.nexus, am.signer, am.multisig))

	err := cfg.RegisterMigration(types.ModuleName, 4, keeper.GetMigrationHandler(am.keeper, am.nexus, am.signer, am.multisig))
	if err != nil {
		panic(err)
	}
}

// BeginBlock executes all state transitions this module requires at the beginning of each new block
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req, am.keeper)
}

// EndBlock executes all state transitions this module requires at the end of each new block
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return utils.RunCached(ctx, am.keeper, func(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
		return EndBlocker(ctx, req, am.keeper, am.nexus, am.multisig)
	})
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 5 }
