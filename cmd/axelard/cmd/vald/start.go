package vald

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	"github.com/axelarnetwork/tm-events/pkg/pubsub"
	"github.com/axelarnetwork/tm-events/pkg/tendermint/client"
	tmEvents "github.com/axelarnetwork/tm-events/pkg/tendermint/events"
	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/utils"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/blocks"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast"
	bcTypes "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast/types"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/btc"
	btcRPC "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/btc/rpc"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/eth"
	ethRPC "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/eth/rpc"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/events"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/jobs"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetValdCommand returns the command to start vald
func GetValdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "vald-start",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config
			genFile := config.GenesisFile()
			appState, _, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			logger := serverCtx.Logger.With("module", "vald")

			cliCtx, err := sdkClient.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// dynamically adjust gas limit by simulating the tx first
			txf := tx.NewFactoryCLI(cliCtx, cmd.Flags()).WithSimulateAndExecute(true)

			hub, err := newHub(logger)
			if err != nil {
				return err
			}

			axConf, valAddr := loadConfig()
			if valAddr == "" {
				return fmt.Errorf("validator address not set")
			}

			logger.Info("Start listening to events")
			listen(cliCtx, appState, hub, txf, axConf, valAddr, logger)
			logger.Info("Shutting down")
			return nil
		},
	}
	setPersistentFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)

	values := map[string]string{
		flags.FlagChainID:        app.Name,
		flags.FlagKeyringBackend: "test",
		flags.FlagGasAdjustment:  "2",
		flags.FlagBroadcastMode:  flags.BroadcastSync,
	}
	utils.OverwriteFlagDefaults(cmd, values, true)

	return cmd
}

func setPersistentFlags(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().String("tofnd-host", "", "host name for tss daemon")
	_ = viper.BindPFlag("tofnd_host", rootCmd.PersistentFlags().Lookup("tofnd-host"))

	rootCmd.PersistentFlags().String("tofnd-port", "50051", "port for tss daemon")
	_ = viper.BindPFlag("tofnd_port", rootCmd.PersistentFlags().Lookup("tofnd-port"))

	rootCmd.PersistentFlags().String("validator-addr", "", "the address of the validator operator")
	_ = viper.BindPFlag("validator-addr", rootCmd.PersistentFlags().Lookup("validator-addr"))

	rootCmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")
	_ = viper.BindPFlag(flags.FlagChainID, rootCmd.PersistentFlags().Lookup(flags.FlagChainID))
}

func newHub(logger log.Logger) (*tmEvents.Hub, error) {
	c, err := client.NewClient(client.DefaultAddress, client.DefaultWSEndpoint, logger)
	if err != nil {
		return nil, err
	}

	hub := tmEvents.NewHub(c, logger)
	return &hub, nil
}

func loadConfig() (app.Config, string) {
	// need to merge in cli config because axelard now has its own broadcasting client
	conf := app.DefaultConfig()
	cliCfgFile := path.Join(app.DefaultNodeHome, "config", "config.toml")
	viper.SetConfigFile(cliCfgFile)
	if err := viper.MergeInConfig(); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(&conf); err != nil {
		panic(err)
	}
	// for some reason gas is not being filled
	conf.Gas = viper.GetUint64("gas")

	return conf, viper.GetString("validator-addr")
}

func listen(ctx sdkClient.Context, appState map[string]json.RawMessage, hub *tmEvents.Hub, txf tx.Factory, axelarCfg app.Config, valAddr string, logger log.Logger) {
	encCfg := app.MakeEncodingConfig()
	cdc := encCfg.Amino
	protoCdc := encCfg.Marshaler
	sender, err := ctx.Keyring.Key(axelarCfg.BroadcastConfig.From)
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to read broadcaster account info from keyring"))
	}
	ctx = ctx.
		WithFromAddress(sender.GetAddress()).
		WithFromName(sender.GetName())

	tssGenesisState := tssTypes.GetGenesisStateFromAppState(protoCdc, appState)

	broadcaster := createBroadcaster(ctx, txf, axelarCfg, logger)

	eventMgr := createEventMgr(ctx, logger)
	tssMgr := createTSSMgr(broadcaster, ctx.FromAddress, &tssGenesisState, axelarCfg, logger, valAddr, cdc)
	btcMgr := createBTCMgr(axelarCfg, broadcaster, ctx.FromAddress, logger, cdc)
	ethMgr := createETHMgr(axelarCfg, broadcaster, ctx.FromAddress, logger, cdc)

	blockHeader := tmEvents.MustSubscribeNewBlockHeader(hub)

	keygenStart := tmEvents.MustSubscribeTx(eventMgr, tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueStart)
	keygenMsg := tmEvents.MustSubscribeTx(eventMgr, tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueMsg)
	signStart := tmEvents.MustSubscribeTx(eventMgr, tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueStart)
	signMsg := tmEvents.MustSubscribeTx(eventMgr, tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueMsg)

	btcConf := tmEvents.MustSubscribeTx(eventMgr, btcTypes.EventTypeOutpointConfirmation, btcTypes.ModuleName, btcTypes.AttributeValueStart)

	ethNewChain := tmEvents.MustSubscribeTx(hub, evmTypes.EventTypeNewChain, evmTypes.ModuleName, evmTypes.AttributeValueUpdate)
	ethChainConf := tmEvents.MustSubscribeTx(hub, evmTypes.EventTypeChainConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	ethDepConf := tmEvents.MustSubscribeTx(eventMgr, evmTypes.EventTypeDepositConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	ethTokConf := tmEvents.MustSubscribeTx(eventMgr, evmTypes.EventTypeTokenConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)

	js := []jobs.Job{
		events.Consume(blockHeader, func(h int64, _ []sdk.Attribute) error { return eventMgr.QueryTxEvents(h) }),
		events.Consume(blockHeader, tssMgr.ProcessNewBlockHeader),
		events.Consume(keygenStart, tssMgr.ProcessKeygenStart),
		events.Consume(keygenMsg, tssMgr.ProcessKeygenMsg),
		events.Consume(signStart, tssMgr.ProcessSignStart),
		events.Consume(signMsg, tssMgr.ProcessSignMsg),
		events.Consume(btcConf, btcMgr.ProcessConfirmation),
		events.Consume(ethNewChain, ethMgr.ProcessNewChain),
		events.Consume(ethChainConf, ethMgr.ProcessChainConfirmation),
		events.Consume(ethDepConf, ethMgr.ProcessDepositConfirmation),
		events.Consume(ethTokConf, ethMgr.ProcessTokenConfirmation),
	}

	// errGroup runs async processes and cancels their context if ANY of them returns an error.
	// Here, we don't want to stop on errors, but simply log it and continue, so errGroup doesn't cut it
	logErr := func(err error) { logger.Error(err.Error()) }
	mgr := jobs.NewMgr(logErr)
	mgr.AddJobs(js...)
	mgr.Wait()
}

// Somewhere in the event pipeline abci.Event needs to be converted into an event type FilteredSubscriber understands
// to decouple event routing from type mapping and to make the conversion explicit, wrappedBus wraps around a given bus to
// convert incoming events
type wrappedBus struct {
	pubsub.Bus
}

// Publish implements the tmEvents.Publisher interface
func (b wrappedBus) Publish(event pubsub.Event) error {
	abciEvent, ok := event.(abci.Event)
	if !ok {
		return fmt.Errorf("expected event of type %T, got %T", abci.Event{}, event)
	}
	e, ok := tmEvents.ProcessEvent(abciEvent)
	if !ok {
		return fmt.Errorf("could not parse event %v", event)
	}
	return b.Bus.Publish(e)
}

func createEventMgr(ctx sdkClient.Context, logger log.Logger) *events.Mgr {
	node, err := ctx.GetNode()
	if err != nil {
		panic(err)
	}

	pubSubFactory := func() pubsub.Bus {
		return wrappedBus{pubsub.NewBus()}
	}

	return events.NewMgr(node, pubSubFactory, logger)
}

func createBroadcaster(ctx sdkClient.Context, txf tx.Factory, axelarCfg app.Config, logger log.Logger) bcTypes.Broadcaster {
	pipeline := broadcast.NewPipelineWithRetry(10000, axelarCfg.MaxRetries, broadcast.LinearBackOff(axelarCfg.MinTimeout), logger)
	return broadcast.NewBroadcaster(ctx, txf, pipeline, logger)
}

func createTSSMgr(broadcaster bcTypes.Broadcaster, sender sdk.AccAddress, genesisState *tssTypes.GenesisState, axelarCfg app.Config, logger log.Logger, valAddr string, cdc *codec.LegacyAmino) *tss.Mgr {
	create := func() (*tss.Mgr, error) {
		gg20client, err := tss.CreateTOFNDClient(axelarCfg.TssConfig.Host, axelarCfg.TssConfig.Port, logger)
		if err != nil {
			return nil, err
		}

		tssMgr := tss.NewMgr(gg20client, 2*time.Hour, valAddr, broadcaster, sender, genesisState.Params.TimeoutInBlocks, logger, cdc)

		return tssMgr, nil
	}
	mgr, err := create()
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to create tss manager"))
	}
	return mgr
}

func createBTCMgr(axelarCfg app.Config, b bcTypes.Broadcaster, sender sdk.AccAddress, logger log.Logger, cdc *codec.LegacyAmino) *btc.Mgr {
	rpc, err := btcRPC.NewRPCClient(axelarCfg.BtcConfig, logger)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
	// clean up btcRPC connection on process shutdown
	tmos.TrapSignal(logger, rpc.Shutdown)

	btcMgr := btc.NewMgr(rpc, b, sender, logger, cdc)
	return btcMgr
}

func createETHMgr(axelarCfg app.Config, b bcTypes.Broadcaster, sender sdk.AccAddress, logger log.Logger, cdc *codec.LegacyAmino) *eth.Mgr {
	rpc, err := ethRPC.NewClient(axelarCfg.EthRPCAddr)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
	// clean up ethRPC connection on process shutdown
	tmos.TrapSignal(logger, rpc.Close)

	ethMgr := eth.NewMgr(rpc, b, sender, logger, cdc)
	return ethMgr
}
