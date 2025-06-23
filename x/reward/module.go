package reward

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/grpc"
	"github.com/axelarnetwork/axelar-core/x/reward/client/cli"
	"github.com/axelarnetwork/axelar-core/x/reward/keeper"
	"github.com/axelarnetwork/axelar-core/x/reward/types"
	"github.com/axelarnetwork/utils/funcs"
)

var (
	_ module.AppModule       = AppModule{}
	_ module.AppModuleBasic  = AppModuleBasic{}
	_ module.HasABCIEndBlock = AppModule{}
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
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis checks the given genesis state for validity
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	funcs.MustNoErr(types.RegisterQueryServiceHandlerClient(context.Background(), mux, types.NewQueryServiceClient(clientCtx)))
}

// GetTxCmd returns all CLI tx commands for this module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns all CLI query commands for this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements module.AppModule
type AppModule struct {
	AppModuleBasic
	keeper       keeper.Keeper
	nexus        types.Nexus
	minter       mintkeeper.Keeper
	staker       types.Staker
	slasher      types.Slasher
	multiSig     types.MultiSig
	snapshotter  types.Snapshotter
	msgSvcRouter *baseapp.MsgServiceRouter
	bank         types.Banker
	cdc          codec.Codec
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	k keeper.Keeper,
	nexus types.Nexus,
	minter mintkeeper.Keeper,
	staker types.Staker,
	slasher types.Slasher,
	multiSig types.MultiSig,
	snapshotter types.Snapshotter,
	bank types.Banker,
	msgSvcRouter *baseapp.MsgServiceRouter,
	cdc codec.Codec,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
		nexus:          nexus,
		minter:         minter,
		staker:         staker,
		slasher:        slasher,
		multiSig:       multiSig,
		snapshotter:    snapshotter,
		msgSvcRouter:   msgSvcRouter,
		bank:           bank,
		cdc:            cdc,
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
	am.keeper.InitGenesis(ctx, &genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports a genesis state from the module's keeper
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(am.keeper.ExportGenesis(ctx))
}

// QuerierRoute returns this module's query route
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	msgServer := keeper.NewMsgServerImpl(am.keeper, am.bank, am.msgSvcRouter, am.cdc)
	types.RegisterMsgServiceServer(grpc.ServerWithSDKErrors{Server: cfg.MsgServer(), Err: types.ErrReward, Logger: am.keeper.Logger}, msgServer)
	types.RegisterQueryServiceServer(cfg.QueryServer(), keeper.NewGRPCQuerier(am.keeper, am.minter, am.nexus))

	err := cfg.RegisterMigration(types.ModuleName, 1, keeper.GetMigrationHandler(am.keeper))
	if err != nil {
		panic(err)
	}
}

// EndBlock executes all state transitions this module requires at the end of each new block
func (am AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	return utils.RunCached(sdk.UnwrapSDKContext(ctx), am.keeper, func(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
		return EndBlocker(sdk.UnwrapSDKContext(ctx), am.keeper, am.nexus, am.minter, am.staker, am.slasher, am.multiSig, am.snapshotter)
	}), nil
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 2 }

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() { // marker
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}
