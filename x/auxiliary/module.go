package auxiliary

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils/grpc"
	"github.com/axelarnetwork/axelar-core/x/auxiliary/keeper"
	"github.com/axelarnetwork/axelar-core/x/auxiliary/types"
)

var (
	_ appmodule.AppModule   = AppModule{}
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

// QuerierRoute returns this module's query route
func (AppModule) QuerierRoute() string {
	return ""
}

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	logger := func(ctx sdk.Context) log.Logger {
		return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
	}

	types.RegisterMsgServiceServer(grpc.ServerWithSDKErrors{Server: cfg.MsgServer(), Err: types.ErrAuxiliary, Logger: logger}, keeper.NewMsgServer(am.router, am.cdc))
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 {
	return 1
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() { // marker
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}
