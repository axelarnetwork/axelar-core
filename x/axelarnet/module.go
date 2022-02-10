package axelarnet

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/ibc-go/modules/apps/transfer"
	channeltypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/modules/core/exported"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/x/axelarnet/client/cli"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/client/rest"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
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

// RegisterRESTRoutes registers the REST routes for this module
func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(clientCtx, rtr)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {
}

// GetTxCmd returns all CLI tx commands for this module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns all CLI query commands for this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// AppModule implements module.AppModule
type AppModule struct {
	AppModuleBasic
	logger   log.Logger
	keeper   keeper.Keeper
	nexus    types.Nexus
	bank     types.BankKeeper
	transfer types.IBCTransferKeeper
	channel  types.ChannelKeeper
	account  types.AccountKeeper

	transferModule transfer.AppModule
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	k keeper.Keeper,
	nexus types.Nexus,
	bank types.BankKeeper,
	transfer types.IBCTransferKeeper,
	channel types.ChannelKeeper,
	account types.AccountKeeper,
	transferModule transfer.AppModule,
	logger log.Logger) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		logger:         logger,
		keeper:         k,
		nexus:          nexus,
		bank:           bank,
		transfer:       transfer,
		channel:        channel,
		account:        account,
		transferModule: transferModule,
	}
}

// RegisterInvariants registers this module's invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {
	// No invariants yet
}

// InitGenesis initializes the module's keeper from the given genesis state
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	am.keeper.InitGenesis(ctx, &genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports a genesis state from the module's keeper
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(genState)
}

// Route returns the module's route
func (am AppModule) Route() sdk.Route {
	return sdk.NewRoute(types.RouterKey, NewHandler(am.keeper, am.nexus, am.bank, am.transfer, am.channel, am.account))
}

// QuerierRoute returns this module's query route
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// LegacyQuerierHandler returns a new query handler for this module
func (am AppModule) LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier {
	return nil
}

// RegisterServices registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterServices(module.Configurator) {
}

// BeginBlock executes all state transitions this module requires at the beginning of each new block
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req)
}

// EndBlock executes all state transitions this module requires at the end of each new block
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return EndBlocker(ctx, req)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// OnChanOpenInit implements the IBCModule interface
func (am AppModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	return am.transferModule.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, version)
}

// OnChanOpenTry implements the IBCModule interface
func (am AppModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version,
	counterpartyVersion string,
) error {
	return am.transferModule.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, version, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface
func (am AppModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyVersion string,
) error {
	return am.transferModule.OnChanOpenAck(ctx, portID, channelID, counterpartyVersion)
}

// OnChanOpenConfirm implements the IBCModule interface
func (am AppModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return am.transferModule.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface
func (am AppModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return am.transferModule.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements the IBCModule interface
func (am AppModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return am.transferModule.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket implements the IBCModule interface. A successful acknowledgement
// is returned if the packet data is succesfully decoded and the receive application
// logic returns without error.
func (am AppModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return am.transferModule.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements the IBCModule interface
func (am AppModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) (*sdk.Result, error) {
	result, err := am.transferModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	if err == nil {
		var ack channeltypes.Acknowledgement
		_ = types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack)
		switch ack.Response.(type) {
		case *channeltypes.Acknowledgement_Error:
			err = resendTransferRoutedByAxelar(ctx, am.keeper, am.transfer, am.channel, packet)
		default:
			// the acknowledgement succeeded on the receiving chain, delete the pending ibc transfer if it routed by axelarnet
			am.keeper.DeletePendingIBCTransfer(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
		}
	}
	return result, err
}

// OnTimeoutPacket implements the IBCModule interface
func (am AppModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) (*sdk.Result, error) {
	result, err := am.transferModule.OnTimeoutPacket(ctx, packet, relayer)
	if err == nil {
		err = resendTransferRoutedByAxelar(ctx, am.keeper, am.transfer, am.channel, packet)
	}
	return result, err
}

func resendTransferRoutedByAxelar(ctx sdk.Context, k keeper.Keeper, t types.IBCTransferKeeper, c types.ChannelKeeper, packet channeltypes.Packet) error {
	// resend pending IBC transfer routed by axelarnet
	p, ok := k.GetPendingIBCTransfer(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	if !ok {
		return nil
	}
	path := fmt.Sprintf("%s/%s", packet.GetSourcePort(), packet.GetSourceChannel())
	err := keeper.IBCTransfer(ctx, k, t, c, p.Token, p.Sender, p.Receiver, path)
	// delete the timed out transfer
	if err != nil {
		k.DeletePendingIBCTransfer(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence())
	}
	return err
}
