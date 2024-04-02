package auxiliary

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
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

	"github.com/axelarnetwork/axelar-core/utils/grpc"
	"github.com/axelarnetwork/axelar-core/x/auxiliary/keeper"
	"github.com/axelarnetwork/axelar-core/x/auxiliary/types"
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
func (AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

// RegisterInterfaces registers the module's interface types
func (AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns the default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genesis := types.GenesisState{}
	return cdc.MustMarshalJSON(&genesis)
}

// ValidateGenesis checks the given genesis state for validity
func (AppModuleBasic) ValidateGenesis(_ codec.JSONCodec, _ client.TxEncodingConfig, _ json.RawMessage) error {
	return nil
}

// RegisterRESTRoutes registers the REST routes for this module
func (AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

// GetTxCmd returns all CLI tx commands for this module
func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

// GetQueryCmd returns all CLI query commands for this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil }

// AppModule implements module.AppModule
type AppModule struct {
	AppModuleBasic
	cdc    codec.Codec
	router *baseapp.MsgServiceRouter
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, router *baseapp.MsgServiceRouter) AppModule {
	return AppModule{
		cdc:    cdc,
		router: router,
	}
}

// RegisterInvariants registers this module's invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {
	// No invariants yet
}

// InitGenesis initializes the module's keeper from the given genesis state
func (am AppModule) InitGenesis(_ sdk.Context, _ codec.JSONCodec, _ json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports a genesis state from the module's keeper
func (am AppModule) ExportGenesis(_ sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis := types.GenesisState{}

	return cdc.MustMarshalJSON(&genesis)
}

// Route returns the module's route
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, nil)
}

// QuerierRoute returns this module's query route
func (AppModule) QuerierRoute() string {
	return ""
}

// LegacyQuerierHandler returns a new query handler for this module
func (am AppModule) LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	logger := func(ctx sdk.Context) log.Logger {
		return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
	}

	types.RegisterMsgServiceServer(grpc.ServerWithSDKErrors{Server: cfg.MsgServer(), Err: types.ErrAuxiliary, Logger: logger}, keeper.NewMsgServer(am.router))
}

// BeginBlock executes all state transitions this module requires at the beginning of each new block
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock executes all state transitions this module requires at the end of each new block
func (am AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 {
	return 1
}
