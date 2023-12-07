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
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v4/modules/core/05-port/types"
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
	"github.com/axelarnetwork/utils/funcs"
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
	logger  log.Logger
	ibcK    keeper.IBCKeeper
	keeper  keeper.Keeper
	nexus   types.Nexus
	bank    types.BankKeeper
	account types.AccountKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(ibcK keeper.IBCKeeper, nexus types.Nexus, bank types.BankKeeper, account types.AccountKeeper, logger log.Logger) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		logger:         logger,
		ibcK:           ibcK,
		keeper:         ibcK.Keeper,
		nexus:          nexus,
		bank:           bank,
		account:        account,
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

	err := cfg.RegisterMigration(types.ModuleName, 5, keeper.Migrate5to6(am.keeper))
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
func (AppModule) ConsensusVersion() uint64 { return 6 }

// AxelarnetIBCModule is an IBCModule that adds rate limiting and gmp processing to the ibc middleware
type AxelarnetIBCModule struct {
	porttypes.IBCModule
	keeper      keeper.Keeper
	nexus       types.Nexus
	ibcK        keeper.IBCKeeper
	bank        types.BankKeeper
	rateLimiter RateLimiter
}

// NewAxelarnetIBCModule creates a new AxelarnetIBCModule instance
func NewAxelarnetIBCModule(
	transferModule porttypes.IBCModule,
	ibcK keeper.IBCKeeper,
	rateLimiter RateLimiter,
	nexus types.Nexus,
	bank types.BankKeeper,
) AxelarnetIBCModule {
	return AxelarnetIBCModule{
		IBCModule:   transferModule,
		keeper:      ibcK.Keeper,
		nexus:       nexus,
		ibcK:        ibcK,
		bank:        bank,
		rateLimiter: rateLimiter,
	}
}

// OnRecvPacket implements the IBCModule interface. A successful acknowledgement
// is returned if the packet data is succesfully decoded and the receive application
// logic returns without error.
func (m AxelarnetIBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	// TODO: split axelar routed packets and gmp into separated middleware?

	ack := m.IBCModule.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	return OnRecvMessage(ctx, m.keeper, m.ibcK, m.nexus, m.bank, m.rateLimiter, packet)
}

// OnAcknowledgementPacket implements the IBCModule interface
func (m AxelarnetIBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	err := m.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
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
		return setRoutedPacketCompleted(ctx, m.keeper, m.nexus, port, channel, sequence)
	default:
		// AckError causes a refund of the token (i.e unlock from the escrow address/mint of token depending on whether it's native to chain).
		// Hence, it's rate limited on the Incoming direction (tokens coming in to Axelar hub).
		if err := m.rateLimiter.RateLimitPacket(ctx, packet, nexus.Incoming, types.NewIBCPath(port, channel)); err != nil {
			return err
		}

		return m.setRoutedPacketFailed(ctx, packet)
	}
}

// OnTimeoutPacket implements the IBCModule interface
func (m AxelarnetIBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	err := m.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
	if err != nil {
		return err
	}

	// IBC timeout packets, by convention, use the source port/channel to represent native chain -> counterparty chain channel id
	// https://github.com/cosmos/ibc/tree/main/spec/core/ics-004-channel-and-packet-semantics#definitions
	port, channel := packet.GetSourcePort(), packet.GetSourceChannel()

	// Timeout causes a refund of the token (i.e unlock from the escrow address/mint of token depending on whether it's native to chain).
	// Hence, it's rate limited on the Incoming direction (tokens coming in to Axelar hub).
	if err := m.rateLimiter.RateLimitPacket(ctx, packet, nexus.Incoming, types.NewIBCPath(port, channel)); err != nil {
		return err
	}

	return m.setRoutedPacketFailed(ctx, packet)
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

		k.Logger(ctx).Info(fmt.Sprintf("IBC transfer %d completed", transferID),
			"transferID", transferID, "portID", portID, "channelID", channelID, "sequence", seq)

		return k.SetTransferCompleted(ctx, transferID)
	}

	// check if the packet is Axelar routed general message
	messageID, ok := getSeqMessageIDMapping(ctx, k, portID, channelID, seq)
	if ok {
		k.Logger(ctx).Debug(fmt.Sprintf("general message %s executed", messageID),
			"messageID", messageID, "portID", portID, "channelID", channelID, "sequence", seq)

		return n.SetMessageExecuted(ctx, messageID)
	}

	return nil
}

func (m AxelarnetIBCModule) setRoutedPacketFailed(ctx sdk.Context, packet channeltypes.Packet) error {
	// IBC ack/timeout packets, by convention, use the source port/channel to represent native chain -> counterparty chain channel id
	// https://github.com/cosmos/ibc/tree/main/spec/core/ics-004-channel-and-packet-semantics#definitions
	port, channel, sequence := packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetSequence()

	// check if the packet is Axelar routed cross chain transfer
	transferID, ok := getSeqIDMapping(ctx, m.keeper, port, channel, sequence)
	if ok {
		events.Emit(ctx,
			&types.IBCTransferFailed{
				ID:        transferID,
				Sequence:  sequence,
				PortID:    port,
				ChannelID: channel,
			})
		m.keeper.Logger(ctx).Info(fmt.Sprintf("IBC transfer %d failed", transferID),
			"transferID", transferID, "portID", port, "channelID", channel, "sequence", sequence)

		return m.keeper.SetTransferFailed(ctx, transferID)
	}

	// check if the packet is Axelar routed general message
	messageID, ok := getSeqMessageIDMapping(ctx, m.keeper, port, channel, sequence)
	if ok {
		coin, err := keeper.NewCoin(ctx, m.ibcK, m.nexus, extractTokenFromAckOrTimeoutPacket(packet))
		if err != nil {
			return err
		}

		err = coin.Lock(m.bank, types.AxelarGMPAccount)
		if err != nil {
			return err
		}

		m.keeper.Logger(ctx).Debug(fmt.Sprintf("general message %s failed to execute", messageID),
			"messageID", messageID, "portID", port, "channelID", channel, "sequence", sequence)

		return m.nexus.SetMessageFailed(ctx, messageID)
	}

	return nil
}

func extractTokenFromAckOrTimeoutPacket(packet channeltypes.Packet) sdk.Coin {
	data := funcs.Must(types.ToICS20Packet(packet))

	trace := ibctransfertypes.ParseDenomTrace(data.Denom)
	amount := funcs.MustOk(sdk.NewIntFromString(data.Amount))

	return sdk.NewCoin(trace.IBCDenom(), amount)
}
