package vald

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/utils"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast"
	bcTypes "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast/types"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/btc"
	btcRPC "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/btc/rpc"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/events"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/evm"
	evmRPC "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/jobs"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

var once sync.Once
var cleanupCommands []func()

// GetValdCommand returns the command to start vald
func GetValdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "vald-start",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			logger := serverCtx.Logger.With("module", "vald")

			// in case of panic we still want to try and cleanup resources,
			// but we have to make sure it's not called more than once if the program is stopped by an interrupt signal
			defer once.Do(cleanUp)

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigs
				logger.Info(fmt.Sprintf("captured signal \"%s\"", sig))
				once.Do(cleanUp)
			}()

			config := serverCtx.Config
			genFile := config.GenesisFile()
			appState, _, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

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

			valdHome := filepath.Join(cliCtx.HomeDir, "vald")
			if _, err := os.Stat(valdHome); os.IsNotExist(err) {
				logger.Info(fmt.Sprintf("folder %s does not exist, creating...", valdHome))
				err := os.Mkdir(valdHome, 0755)
				if err != nil {
					return err
				}
			}

			fPath := filepath.Join(valdHome, "state.json")
			f, err := os.OpenFile(fPath, os.O_CREATE|os.O_RDWR, 0755)
			if err != nil {
				return err
			}
			stateStore := events.NewStateStore(f)

			logger.Info("Start listening to events")
			listen(cliCtx, appState, hub, txf, axConf, valAddr, stateStore, logger)
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

func cleanUp() {
	for _, cmd := range cleanupCommands {
		cmd()
	}
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

func listen(ctx sdkClient.Context, appState map[string]json.RawMessage, hub *tmEvents.Hub, txf tx.Factory, axelarCfg app.Config, valAddr string, store events.StateStore, logger log.Logger) {
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

	eventMgr := createEventMgr(ctx, store, logger)
	tssMgr := createTSSMgr(broadcaster, ctx.FromAddress, &tssGenesisState, axelarCfg, logger, valAddr, cdc)
	btcMgr := createBTCMgr(axelarCfg, broadcaster, ctx.FromAddress, logger, cdc)
	ethMgr := createEVMMgr(axelarCfg, broadcaster, ctx.FromAddress, logger, cdc)

	// we have two processes listening to block headers
	blockHeader1 := tmEvents.MustSubscribeNewBlockHeader(hub)
	blockHeader2 := tmEvents.MustSubscribeNewBlockHeader(hub)

	keygenStart := tmEvents.MustSubscribeTx(eventMgr, tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueStart)
	keygenMsg := tmEvents.MustSubscribeTx(eventMgr, tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueMsg)
	signStart := tmEvents.MustSubscribeTx(eventMgr, tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueStart)
	signMsg := tmEvents.MustSubscribeTx(eventMgr, tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueMsg)

	btcConf := tmEvents.MustSubscribeTx(eventMgr, btcTypes.EventTypeOutpointConfirmation, btcTypes.ModuleName, btcTypes.AttributeValueStart)

	ethDepConf := tmEvents.MustSubscribeTx(eventMgr, evmTypes.EventTypeDepositConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	ethTokConf := tmEvents.MustSubscribeTx(eventMgr, evmTypes.EventTypeTokenConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)

	// stop the jobs if process gets interrupted/terminated
	cleanupCommands = append(cleanupCommands, func() {
		logger.Info("persisting event state...")
		eventMgr.Shutdown()
		logger.Info("event state persisted")
		logger.Info("stopping listening for blocks...")
		blockHeader1.Close()
		blockHeader2.Close()
		logger.Info("block listener stopped")
	})

	fetchEvents := func(errChan chan<- error) {
		for err := range eventMgr.FetchEvents() {
			errChan <- err
		}
	}
	js := []jobs.Job{
		fetchEvents,
		events.Consume(blockHeader1, events.OnlyBlockHeight(eventMgr.NotifyNewBlock)),
		events.Consume(blockHeader2, events.OnlyBlockHeight(tssMgr.ProcessNewBlockHeader)),
		events.Consume(keygenStart, tssMgr.ProcessKeygenStart),
		events.Consume(keygenMsg, events.OnlyAttributes(tssMgr.ProcessKeygenMsg)),
		events.Consume(signStart, tssMgr.ProcessSignStart),
		events.Consume(signMsg, events.OnlyAttributes(tssMgr.ProcessSignMsg)),
		events.Consume(btcConf, events.OnlyAttributes(btcMgr.ProcessConfirmation)),
		events.Consume(ethDepConf, events.OnlyAttributes(ethMgr.ProcessDepositConfirmation)),
		events.Consume(ethTokConf, events.OnlyAttributes(ethMgr.ProcessTokenConfirmation)),
	}

	// errGroup runs async processes and cancels their context if ANY of them returns an error.
	// Here, we don't want to stop on errors, but simply log it and continue, so errGroup doesn't cut it
	logErr := func(err error) { logger.Error(err.Error()) }
	mgr := jobs.NewMgr(logErr)
	mgr.AddJobs(js...)
	mgr.Wait()
}

func createEventMgr(ctx sdkClient.Context, store events.StateStore, logger log.Logger) *events.Mgr {
	node, err := ctx.GetNode()
	if err != nil {
		panic(err)
	}

	return events.NewMgr(node, store, pubsub.NewBus, logger)
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
	cleanupCommands = append(cleanupCommands, rpc.Shutdown)

	logger.Info("Successfully connected to Bitcoin bridge ")

	btcMgr := btc.NewMgr(rpc, b, sender, logger, cdc)
	return btcMgr
}

func createEVMMgr(axelarCfg app.Config, b bcTypes.Broadcaster, sender sdk.AccAddress, logger log.Logger, cdc *codec.LegacyAmino) *evm.Mgr {
	rpcs := make(map[string]evmRPC.Client)

	for _, evmChainConf := range axelarCfg.EVMConfig {
		if !evmChainConf.WithBridge {
			continue
		}

		if _, found := rpcs[strings.ToLower(evmChainConf.Name)]; found {
			msg := fmt.Errorf("duplicate bridge configuration found for EVM chain %s", evmChainConf.Name)
			logger.Error(msg.Error())
			panic(msg)
		}

		rpc, err := evmRPC.NewClient(evmChainConf.RPCAddr)
		if err != nil {
			logger.Error(err.Error())
			panic(err)
		}
		// clean up ethRPC connection on process shutdown
		cleanupCommands = append(cleanupCommands, rpc.Close)

		rpcs[strings.ToLower(evmChainConf.Name)] = rpc
		logger.Info(fmt.Sprintf("Successfully connected to EVM bridge for chain %s", evmChainConf.Name))
	}

	ethMgr := evm.NewMgr(rpcs, b, sender, logger, cdc)
	return ethMgr
}
