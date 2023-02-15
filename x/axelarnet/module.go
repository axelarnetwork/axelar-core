package axelarnet

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/ibc-go/v4/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v4/modules/core/exported"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/utils/events"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/client/cli"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/client/rest"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/keeper"
	"github.com/axelarnetwork/axelar-core/x/axelarnet/types"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
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
	return cli.GetQueryCmd()
}

// AppModule implements module.AppModule
type AppModule struct {
	AppModuleBasic
	logger      log.Logger
	keeper      keeper.Keeper
	nexus       types.Nexus
	bank        types.BankKeeper
	channel     types.ChannelKeeper
	account     types.AccountKeeper
	ibcK        keeper.IBCKeeper
	rateLimiter RateLimiter

	transferModule transfer.IBCModule
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	k keeper.Keeper,
	nexus types.Nexus,
	bank types.BankKeeper,
	account types.AccountKeeper,
	ibcK keeper.IBCKeeper,
	transferModule transfer.IBCModule,
	rateLimiter RateLimiter,
	logger log.Logger) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		logger:         logger,
		keeper:         k,
		nexus:          nexus,
		bank:           bank,
		account:        account,
		ibcK:           ibcK,
		transferModule: transferModule,
		rateLimiter:    rateLimiter,
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
	return sdk.NewRoute(types.RouterKey, NewHandler(am.keeper, am.nexus, am.bank, am.account, am.ibcK))
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
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServiceServer(cfg.QueryServer(), keeper.NewGRPCQuerier(am.keeper, am.nexus))

	err := cfg.RegisterMigration(types.ModuleName, 4, keeper.Migrate4To5(am.keeper))
	if err != nil {
		panic(err)
	}
}

// BeginBlock executes all state transitions this module requires at the beginning of each new block
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, req)
}

// EndBlock executes all state transitions this module requires at the end of each new block
func (am AppModule) EndBlock(ctx sdk.Context, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return utils.RunCached(ctx, am.keeper, func(ctx sdk.Context) ([]abci.ValidatorUpdate, error) {
		return EndBlocker(ctx, req, am.keeper, am.ibcK)
	})
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 5 }

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
) (string, error) {
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
	counterpartyVersion string,
) (string, error) {
	return am.transferModule.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface
func (am AppModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return am.transferModule.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
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
	// TODO: split axelar routed packets and gmp into separated middleware?

	ack := am.transferModule.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	return OnRecvMessage(ctx, am.keeper, am.ibcK, am.nexus, am.bank, am.rateLimiter, packet)
}

// OnAcknowledgementPacket implements the IBCModule interface
func (am AppModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	err := am.transferModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
	if err != nil {
		return err
	}

	var ack channeltypes.Acknowledgement
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	if err := ack.ValidateBasic(); err != nil {
		return err
	}

	// IBC ack packets, by convention, use the source port/channel to represent native chain -> counterparty chain channel id
	// https://github.com/cosmos/ibc/tree/main/spec/core/ics-004-channel-and-packet-semantics#definitions
	port, channel, sequence := packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence()

	switch ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		return setRoutedPacketCompleted(ctx, am.keeper, am.nexus, port, channel, sequence)
	default:
		// AckError causes a refund of the token (i.e unlock from the escrow address/mint of token depending on whether it's native to chain).
		// Hence, it's rate limited on the Incoming direction (tokens coming in to Axelar hub).
		if err := am.rateLimiter.RateLimitPacket(ctx, packet, nexus.Incoming, types.NewIBCPath(port, channel)); err != nil {
			return err
		}

		return setRoutedPacketFailed(ctx, am.keeper, am.nexus, port, channel, sequence)
	}
}

// OnTimeoutPacket implements the IBCModule interface
func (am AppModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	err := am.transferModule.OnTimeoutPacket(ctx, packet, relayer)
	if err != nil {
		return err
	}

	// IBC timeout packets, by convention, use the source port/channel to represent native chain -> counterparty chain channel id
	// https://github.com/cosmos/ibc/tree/main/spec/core/ics-004-channel-and-packet-semantics#definitions
	port, channel, sequence := packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence()

	// Timeout causes a refund of the token (i.e unlock from the escrow address/mint of token depending on whether it's native to chain).
	// Hence, it's rate limited on the Incoming direction (tokens coming in to Axelar hub).
	if err := am.rateLimiter.RateLimitPacket(ctx, packet, nexus.Incoming, types.NewIBCPath(port, channel)); err != nil {
		return err
	}

	return setRoutedPacketFailed(ctx, am.keeper, am.nexus, port, channel, sequence)
}

// GetAppVersion implements the ChannelKeeper interface
func (am AppModule) GetAppVersion(ctx sdk.Context, portID string, channelID string) (string, bool) {
	return am.channel.GetAppVersion(ctx, portID, channelID)
}

// returns the transfer id and delete the existing mapping
func getSeqIDMapping(ctx sdk.Context, k keeper.Keeper, portID, channelID string, seq uint64) (nexus.TransferID, bool) {
	defer k.DeleteSeqIDMapping(ctx, portID, channelID, seq)

	return k.GetSeqIDMapping(ctx, portID, channelID, seq)
}

// returns the general message id and delete the existing mapping
func getSeqMessageIDMapping(ctx sdk.Context, k keeper.Keeper, portID, channelID string, seq uint64) (string, bool) {
	defer k.DeleteSeqMessageIDMapping(ctx, portID, channelID, seq)

	return k.GetSeqMessageIDMapping(ctx, portID, channelID, seq)
}

func setRoutedPacketCompleted(ctx sdk.Context, k keeper.Keeper, n types.Nexus, portID, channelID string, seq uint64) error {
	// check if the packet is Axelar routed cross chain transfer
	transferID, ok := getSeqIDMapping(ctx, k, portID, channelID, seq)
	if ok {
		events.Emit(ctx,
			&types.IBCTransferCompleted{
				ID:        transferID,
				Sequence:  seq,
				PortID:    portID,
				ChannelID: channelID,
			})
		k.Logger(ctx).Info(fmt.Sprintf("set IBC transfer %d completed", transferID))

		return k.SetTransferCompleted(ctx, transferID)
	}

	// check if the packet is Axelar routed general message
	messageID, ok := getSeqMessageIDMapping(ctx, k, portID, channelID, seq)
	if ok {
		k.Logger(ctx).Debug("set general message status to executed", "messageID", messageID)
		return n.SetMessageExecuted(ctx, messageID)
	}

	return nil
}

func setRoutedPacketFailed(ctx sdk.Context, k keeper.Keeper, n types.Nexus, portID, channelID string, seq uint64) error {
	// check if the packet is Axelar routed cross chain transfer
	transferID, ok := getSeqIDMapping(ctx, k, portID, channelID, seq)
	if ok {
		events.Emit(ctx,
			&types.IBCTransferFailed{
				ID:        transferID,
				Sequence:  seq,
				PortID:    portID,
				ChannelID: channelID,
			})
		k.Logger(ctx).Info(fmt.Sprintf("set IBC transfer %d failed", transferID))

		return k.SetTransferFailed(ctx, transferID)
	}

	// check if the packet is Axelar routed general message
	messageID, ok := getSeqMessageIDMapping(ctx, k, portID, channelID, seq)
	if ok {
		k.Logger(ctx).Debug("set general message status to failed", "messageID", messageID)
		return n.SetMessageFailed(ctx, messageID)
	}

	return nil
}
