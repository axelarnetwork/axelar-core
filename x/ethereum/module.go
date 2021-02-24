package ethereum

import (
	"context"
	"encoding/json"
	"fmt"

	sdkCli "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/ethereum/client/cli"
	"github.com/axelarnetwork/axelar-core/x/ethereum/client/rest"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
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

// RegisterCodec registers the types necessary in this module with the given codec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// DefaultGenesis returns the default genesis state
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis checks the given genesis state for validity
func (AppModuleBasic) ValidateGenesis(message json.RawMessage) error {
	var data types.GenesisState
	err := types.ModuleCdc.UnmarshalJSON(message, &data)
	if err != nil {
		return err
	}
	return types.ValidateGenesis(data)
}

// RegisterRESTRoutes registers the REST routes for this module
func (AppModuleBasic) RegisterRESTRoutes(cliCtx sdkCli.CLIContext, rtr *mux.Router) {
	rest.RegisterRoutes(cliCtx, rtr)
}

// GetTxCmd returns all CLI tx commands for this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}

// GetQueryCmd returns all CLI query commands for this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(types.QuerierRoute, cdc)
}

// AppModule implements module.AppModule
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
	voter  types.Voter
	nexus  types.Nexus
	rpc    types.RPCClient
	signer types.Signer
}

// NewAppModule creates a new AppModule object
func NewAppModule(k keeper.Keeper, voter types.Voter, signer types.Signer, nexus types.Nexus, rpc types.RPCClient) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		voter:          voter,
		signer:         signer,
		nexus:          nexus,
		rpc:            rpc,
	}
}

// RegisterInvariants registers this module's invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {
	// No invariants yet
}

// InitGenesis initializes the module's keeper from the given genesis state
func (am AppModule) InitGenesis(ctx sdk.Context, message json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(message, &genesisState)
	id, err := am.rpc.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	actualNetwork := types.NetworkByID(id)
	if genesisState.Params.Network != actualNetwork {
		panic(fmt.Sprintf(
			"local ethereum client not configured correctly: expected network %s, got %s",
			genesisState.Params.Network,
			actualNetwork,
		))
	}
	InitGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports a genesis state from the module's keeper
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return types.ModuleCdc.MustMarshalJSON(gs)
}

// Route returns the module's route
func (AppModule) Route() string {
	return types.RouterKey
}

// NewHandler returns a new handler for this module
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(am.keeper, am.rpc, am.voter, am.signer, am.nexus)
}

// QuerierRoute returns this module's query route
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// NewQuerierHandler returns a new query handler for this module
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(am.rpc, am.keeper, am.signer)
}

// BeginBlock executes all state transitions this module requires at the beginning of each new block
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req, am.keeper)
}

// EndBlock executes all state transitions this module requires at the end of each new block
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return EndBlocker(ctx, req, am.keeper)
}
