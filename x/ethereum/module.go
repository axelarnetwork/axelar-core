package ethereum

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/axelarnetwork/axelar-core/x/ethereum/client/cli"
	"github.com/axelarnetwork/axelar-core/x/ethereum/keeper"
	"github.com/axelarnetwork/axelar-core/x/ethereum/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct {
}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (AppModuleBasic) ValidateGenesis(message json.RawMessage) error {
	var data types.GenesisState
	err := types.ModuleCdc.UnmarshalJSON(message, &data)
	if err != nil {
		return err
	}
	return types.ValidateGenesis(data)
}

func (AppModuleBasic) RegisterRESTRoutes(_ context.CLIContext, _ *mux.Router) {
	// TODO: implement rest interface
}

func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}

func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueryCmd(types.QuerierRoute, cdc)
}

type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
	voter  types.Voter
	rpc    types.RPCClient
	signer types.Signer
}

// NewAppModule creates a new AppModule object
func NewAppModule(k keeper.Keeper, voter types.Voter, signer types.Signer, rpc types.RPCClient) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		voter:          voter,
		signer:         signer,
		rpc:            rpc,
	}
}

func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {
	// No invariants yet
}

func (am AppModule) InitGenesis(ctx sdk.Context, message json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(message, &genesisState)
	InitGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return types.ModuleCdc.MustMarshalJSON(gs)
}

func (AppModule) Route() string {
	return types.RouterKey
}

func (am AppModule) NewHandler() sdk.Handler {
	/*if am.rpc == nil {
		return NewDummyHandler(am.keeper, am.voter)
	}*/

	return NewHandler(am.keeper, am.rpc, am.voter, am.signer)
}

func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

func (am AppModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(am.signer)
}

func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req, am.keeper)
}

func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return EndBlocker(ctx, req, am.keeper)
}
