package vald

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
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
	"github.com/tendermint/tendermint/libs/log"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/utils"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster"
	broadcasterTypes "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types"
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

// RWALL grants rw-rw-rw- file permissions
const RWALL = 0555

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

			axConf := app.DefaultConfig()
			if err := serverCtx.Viper.Unmarshal(&axConf); err != nil {
				panic(err)
			}

			valAddr := serverCtx.Viper.GetString("validator-addr")
			if valAddr == "" {
				return fmt.Errorf("validator address not set")
			}

			valdHome := filepath.Join(cliCtx.HomeDir, "vald")
			if _, err := os.Stat(valdHome); os.IsNotExist(err) {
				logger.Info(fmt.Sprintf("folder %s does not exist, creating...", valdHome))
				err := os.Mkdir(valdHome, RWALL)
				if err != nil {
					return err
				}
			}

			fPath := filepath.Join(valdHome, "state.json")
			stateSource := NewRWFile(fPath)

			logger.Info("Start listening to events")
			listen(cliCtx, appState, hub, txf, axConf, valAddr, stateSource, logger)
			logger.Info("Shutting down")
			return nil
		},
	}
	setPersistentFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)

	values := map[string]string{
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

func setPersistentFlags(cmd *cobra.Command) {
	defaultConf := tssTypes.DefaultConfig()
	cmd.PersistentFlags().String("tofnd-host", defaultConf.Host, "host name for tss daemon")
	cmd.PersistentFlags().String("tofnd-port", defaultConf.Port, "port for tss daemon")
	cmd.PersistentFlags().String("validator-addr", "", "the address of the validator operator")
	cmd.PersistentFlags().String(flags.FlagChainID, app.Name, "The network chain ID")
}

func newHub(logger log.Logger) (*tmEvents.Hub, error) {
	c, err := client.NewClient(client.DefaultAddress, client.DefaultWSEndpoint, logger)
	if err != nil {
		return nil, err
	}

	hub := tmEvents.NewHub(c, logger)
	return &hub, nil
}

func listen(ctx sdkClient.Context, appState map[string]json.RawMessage, hub *tmEvents.Hub, txf tx.Factory, axelarCfg app.Config, valAddr string, stateSource events.ReadWriter, logger log.Logger) {
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

	eventMgr := createEventMgr(ctx, stateSource, logger)
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

	ethNewChain := tmEvents.MustSubscribeTx(hub, evmTypes.EventTypeNewChain, evmTypes.ModuleName, evmTypes.AttributeValueUpdate)
	ethChainConf := tmEvents.MustSubscribeTx(hub, evmTypes.EventTypeChainConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	ethDepConf := tmEvents.MustSubscribeTx(eventMgr, evmTypes.EventTypeDepositConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	ethTokConf := tmEvents.MustSubscribeTx(eventMgr, evmTypes.EventTypeTokenConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	ethTraConf := tmEvents.MustSubscribeTx(eventMgr, evmTypes.EventTypeTransferOwnershipConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)

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
		events.Consume(ethNewChain, events.OnlyAttributes(ethMgr.ProcessNewChain)),
		events.Consume(ethChainConf, events.OnlyAttributes(ethMgr.ProcessChainConfirmation)),
		events.Consume(ethDepConf, events.OnlyAttributes(ethMgr.ProcessDepositConfirmation)),
		events.Consume(ethTokConf, events.OnlyAttributes(ethMgr.ProcessTokenConfirmation)),
		events.Consume(ethTraConf, events.OnlyAttributes(ethMgr.ProcessTransferOwnershipConfirmation)),
	}

	// errGroup runs async processes and cancels their context if ANY of them returns an error.
	// Here, we don't want to stop on errors, but simply log it and continue, so errGroup doesn't cut it
	logErr := func(err error) { logger.Error(err.Error()) }
	mgr := jobs.NewMgr(logErr)
	mgr.AddJobs(js...)
	mgr.Wait()
}

func createEventMgr(ctx sdkClient.Context, stateSource events.ReadWriter, logger log.Logger) *events.Mgr {
	node, err := ctx.GetNode()
	if err != nil {
		panic(err)
	}

	return events.NewMgr(node, stateSource, pubsub.NewBus, logger)
}

func createBroadcaster(ctx sdkClient.Context, txf tx.Factory, axelarCfg app.Config, logger log.Logger) broadcasterTypes.Broadcaster {
	pipeline := broadcaster.NewPipelineWithRetry(10000, axelarCfg.MaxRetries, broadcaster.LinearBackOff(axelarCfg.MinTimeout), logger)
	return broadcaster.NewBroadcaster(ctx, txf, pipeline, logger)
}

func createTSSMgr(broadcaster broadcasterTypes.Broadcaster, sender sdk.AccAddress, genesisState *tssTypes.GenesisState, axelarCfg app.Config, logger log.Logger, valAddr string, cdc *codec.LegacyAmino) *tss.Mgr {
	create := func() (*tss.Mgr, error) {
		gg20client, err := tss.CreateTOFNDClient(axelarCfg.TssConfig.Host, axelarCfg.TssConfig.Port, axelarCfg.TssConfig.DialTimeout, logger)
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

func createBTCMgr(axelarCfg app.Config, b broadcasterTypes.Broadcaster, sender sdk.AccAddress, logger log.Logger, cdc *codec.LegacyAmino) *btc.Mgr {
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

func createEVMMgr(axelarCfg app.Config, b broadcasterTypes.Broadcaster, sender sdk.AccAddress, logger log.Logger, cdc *codec.LegacyAmino) *evm.Mgr {
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

// RWFile implements the ReadWriter interface for an underlying file
type RWFile struct {
	path string
}

// NewRWFile returns a new RWFile instance for the given file path
func NewRWFile(path string) RWFile {
	return RWFile{path: path}
}

// ReadAll returns the full content of the file
func (f RWFile) ReadAll() ([]byte, error) { return os.ReadFile(f.path) }

// WriteAll writes the given bytes to a file. Creates a new fille if it does not exist, overwrites the previous content otherwise.
func (f RWFile) WriteAll(bz []byte) error { return os.WriteFile(f.path, bz, RWALL) }
