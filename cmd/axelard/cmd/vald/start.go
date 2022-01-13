package vald

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	tmEvents "github.com/axelarnetwork/tm-events/events"
	"github.com/axelarnetwork/tm-events/pubsub"
	"github.com/axelarnetwork/utils/jobs"
	"github.com/cosmos/cosmos-sdk/client"
	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/config"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/utils"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster"
	broadcasterTypes "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcaster/types"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/evm"
	evmRPC "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss"
	utils2 "github.com/axelarnetwork/axelar-core/utils"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// RW grants -rw------- file permissions
const RW = 0600

// RWX grants -rwx------ file permissions
const RWX = 0700

var once sync.Once
var cleanupCommands []func()

// GetValdCommand returns the command to start vald
func GetValdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "vald-start",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			if !cmd.Flags().Changed(flags.FlagFrom) {
				if err := cmd.Flags().Set(flags.FlagFrom, serverCtx.Viper.GetString("broadcast.broadcaster-account")); err != nil {
					return err
				}
			}
			return nil
		},
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

			cliCtx, err := sdkClient.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// dynamically adjust gas limit by simulating the tx first
			txf := tx.NewFactoryCLI(cliCtx, cmd.Flags()).WithSimulateAndExecute(true)

			valdConf := config.DefaultValdConfig()
			if err := serverCtx.Viper.Unmarshal(&valdConf); err != nil {
				panic(err)
			}

			valAddr := serverCtx.Viper.GetString("validator-addr")
			if valAddr == "" {
				return fmt.Errorf("validator address not set")
			}

			valdHome := filepath.Join(cliCtx.HomeDir, "vald")
			if _, err := os.Stat(valdHome); os.IsNotExist(err) {
				logger.Info(fmt.Sprintf("folder %s does not exist, creating...", valdHome))
				err := os.Mkdir(valdHome, RWX)
				if err != nil {
					return err
				}
			}

			var recoveryJSON []byte
			recoveryFile := serverCtx.Viper.GetString("tofnd-recovery")
			if recoveryFile != "" {
				recoveryJSON, err = ioutil.ReadFile(recoveryFile)
				if err != nil {
					return err
				}
				if len(recoveryJSON) == 0 {
					return fmt.Errorf("JSON file is empty")
				}
			}

			fPath := filepath.Join(valdHome, "state.json")
			stateSource := NewRWFile(fPath)

			logger.Info("start listening to events")
			listen(cliCtx, txf, valdConf, valAddr, recoveryJSON, stateSource, logger)
			logger.Info("shutting down")
			return nil
		},
	}
	setPersistentFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)
	values := map[string]string{
		flags.FlagGasAdjustment: "4",
		flags.FlagBroadcastMode: flags.BroadcastSync,
		flags.FlagGasPrices:     "0.05uaxl",
	}
	utils.OverwriteFlagDefaults(cmd, values, true)

	// Only set default, not actual value, so it can be overwritten by env variable
	utils.OverwriteFlagDefaults(cmd, map[string]string{
		flags.FlagChainID:        app.Name,
		flags.FlagKeyringBackend: "file",
	}, false)

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
	cmd.PersistentFlags().String("tofnd-recovery", "", "json file with recovery request")
	cmd.PersistentFlags().String("validator-addr", "", "the address of the validator operator")
	cmd.PersistentFlags().String(flags.FlagChainID, app.Name, "The network chain ID")
}

func listen(ctx sdkClient.Context, txf tx.Factory, axelarCfg config.ValdConfig, valAddr string, recoveryJSON []byte, stateSource ReadWriter, logger log.Logger) {
	encCfg := app.MakeEncodingConfig()
	cdc := encCfg.Amino
	sender, err := ctx.Keyring.Key(ctx.From)
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to read broadcaster account info from keyring"))
	}
	ctx = ctx.
		WithFromAddress(sender.GetAddress()).
		WithFromName(sender.GetName())

	bc := createBroadcaster(txf, axelarCfg, logger)

	tmClient, err := ctx.GetNode()
	if err != nil {
		panic(err)
	}

	// in order to subscribe to events, the client needs to be running
	if !tmClient.IsRunning() {
		if err := tmClient.Start(); err != nil {
			panic(fmt.Errorf("unable to start client: %v", err))
		}
	}

	stateStore := NewStateStore(stateSource)
	startBlock, err := getStartBlock(stateStore, tmClient, logger)
	if err != nil {
		panic(err)
	}

	eventBus := createEventBus(tmClient, startBlock, logger)

	tssMgr := createTSSMgr(bc, ctx, axelarCfg, logger, valAddr, cdc)
	if len(recoveryJSON) > 0 {
		if err = tssMgr.Recover(recoveryJSON); err != nil {
			panic(fmt.Errorf("unable to perform tss recovery: %v", err))
		}
	}

	evmMgr := createEVMMgr(axelarCfg, ctx, bc, logger, cdc)

	// we have two processes listening to block headers
	blockHeaderForTSS := tmEvents.MustSubscribeBlockHeader(eventBus)
	blockHeaderForStateUpdate := tmEvents.MustSubscribeBlockHeader(eventBus)

	subscribe := func(eventType, module, action string) tmEvents.FilteredSubscriber {
		return tmEvents.MustSubscribeWithAttributes(eventBus,
			eventType, module, sdk.Attribute{Key: sdk.AttributeKeyAction, Value: action})
	}

	queryHeartBeat := createNewBlockEventQuery(tssTypes.EventTypeHeartBeat, tssTypes.ModuleName, tssTypes.AttributeValueSend)
	heartbeat, err := tmEvents.Subscribe(eventBus, queryHeartBeat)
	if err != nil {
		panic(fmt.Errorf("unable to subscribe with ack event query: %v", err))
	}

	keygenStart := subscribe(tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueStart)

	querySign := createNewBlockEventQuery(tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueStart)
	signStart, err := tmEvents.Subscribe(eventBus, querySign)
	if err != nil {
		panic(fmt.Errorf("unable to subscribe with sign event query: %v", err))
	}

	keygenMsg := subscribe(tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueMsg)
	signMsg := subscribe(tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueMsg)

	evmNewChain := subscribe(evmTypes.EventTypeNewChain, evmTypes.ModuleName, evmTypes.AttributeValueUpdate)
	evmChainConf := subscribe(evmTypes.EventTypeChainConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	evmGatewayDeploymentConf := subscribe(evmTypes.EventTypeGatewayDeploymentConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	evmDepConf := subscribe(evmTypes.EventTypeDepositConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	evmTokConf := subscribe(evmTypes.EventTypeTokenConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	evmTraConf := subscribe(evmTypes.EventTypeTransferKeyConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)

	eventCtx, cancelEventCtx := context.WithCancel(context.Background())
	// stop the jobs if process gets interrupted/terminated
	cleanupCommands = append(cleanupCommands, func() {
		logger.Info("stopping listening for blocks...")
		blockHeaderForTSS.Close()
		logger.Info("block listener stopped")
		logger.Info("stop listening for events...")
		cancelEventCtx()
		<-eventBus.Done()
		logger.Info("event listener stopped")
	})

	fetchEvents := func(errChan chan<- error) { errChan <- <-eventBus.FetchEvents(eventCtx) }
	js := []jobs.Job{
		fetchEvents,
		tmEvents.Consume(blockHeaderForStateUpdate, tmEvents.OnlyBlockHeight(stateStore.SetState)),
		tmEvents.Consume(blockHeaderForTSS, tmEvents.OnlyBlockHeight(func(height int64) error {
			tssMgr.ProcessNewBlockHeader(height)
			return nil
		})),
		tmEvents.Consume(heartbeat, tssMgr.ProcessHeartBeatEvent),
		tmEvents.Consume(keygenStart, tssMgr.ProcessKeygenStart),
		tmEvents.Consume(keygenMsg, tssMgr.ProcessKeygenMsg),
		tmEvents.Consume(signStart, tssMgr.ProcessSignStart),
		tmEvents.Consume(signMsg, tssMgr.ProcessSignMsg),
		tmEvents.Consume(evmNewChain, evmMgr.ProcessNewChain),
		tmEvents.Consume(evmChainConf, evmMgr.ProcessChainConfirmation),
		tmEvents.Consume(evmGatewayDeploymentConf, evmMgr.ProcessGatewayDeploymentConfirmation),
		tmEvents.Consume(evmDepConf, evmMgr.ProcessDepositConfirmation),
		tmEvents.Consume(evmTokConf, evmMgr.ProcessTokenConfirmation),
		tmEvents.Consume(evmTraConf, evmMgr.ProcessTransferKeyConfirmation),
	}

	// errGroup runs async processes and cancels their context if ANY of them returns an error.
	// Here, we don't want to stop on errors, but simply log it and continue, so errGroup doesn't cut it
	logErr := func(err error) { logger.Error(err.Error()) }
	mgr := jobs.NewMgr(logErr)
	mgr.AddJobs(js...)
	mgr.Wait()
}

func getStartBlock(stateStore StateStore, tmClient rpcclient.Client, logger log.Logger) (int64, error) {
	startBlock, err := stateStore.GetState()
	if err != nil {
		logger.Error(err.Error())

		return 0, nil
	}

	rpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := tmClient.Status(rpcCtx)
	if err != nil {
		return 0, err
	}

	// TODO: Make it configurable instead of a hardcoded 100
	if status.SyncInfo.LatestBlockHeight-startBlock > 100 {
		logger.Info(fmt.Sprintf("block in state %d is too old and will start from the latest instead", startBlock))

		return 0, nil
	}

	return startBlock + 1, nil
}

func createNewBlockEventQuery(eventType, module, action string) tmEvents.Query {
	return tmEvents.Query{
		TMQuery: tmEvents.NewBlockHeaderEventQuery(eventType).MatchModule(module).MatchAction(action).Build(),
		Predicate: func(e tmEvents.Event) bool {
			return e.Type == eventType && e.Attributes[sdk.AttributeKeyModule] == module && e.Attributes[sdk.AttributeKeyAction] == action
		},
	}
}

func createEventBus(client rpcclient.Client, startBlock int64, logger log.Logger) *tmEvents.Bus {
	notifier := tmEvents.NewBlockNotifier(tmEvents.NewBlockClient(client), logger).StartingAt(startBlock)
	return tmEvents.NewEventBus(tmEvents.NewBlockSource(client, notifier), pubsub.NewBus, logger)
}

func createBroadcaster(txf tx.Factory, axelarCfg config.ValdConfig, logger log.Logger) broadcasterTypes.Broadcaster {
	pipeline := broadcaster.NewPipelineWithRetry(10000, axelarCfg.MaxRetries, utils2.LinearBackOff(axelarCfg.MinTimeout), logger)
	return broadcaster.NewBroadcaster(txf, pipeline, logger)
}

func createTSSMgr(broadcaster broadcasterTypes.Broadcaster, cliCtx client.Context, axelarCfg config.ValdConfig, logger log.Logger, valAddr string, cdc *codec.LegacyAmino) *tss.Mgr {
	create := func() (*tss.Mgr, error) {
		conn, err := tss.Connect(axelarCfg.TssConfig.Host, axelarCfg.TssConfig.Port, axelarCfg.TssConfig.DialTimeout, logger)
		if err != nil {
			return nil, err
		}
		logger.Debug("successful connection to tofnd gRPC server")

		// creates clients to communicate with the external tofnd process service
		gg20client := tofnd.NewGG20Client(conn)
		multiSigClient := tofnd.NewMultisigClient(conn)

		tssMgr := tss.NewMgr(gg20client, multiSigClient, cliCtx, 2*time.Hour, valAddr, broadcaster, logger, cdc)

		return tssMgr, nil
	}
	mgr, err := create()
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to create tss manager"))
	}

	return mgr
}

func createEVMMgr(axelarCfg config.ValdConfig, cliCtx client.Context, b broadcasterTypes.Broadcaster, logger log.Logger, cdc *codec.LegacyAmino) *evm.Mgr {
	rpcs := make(map[string]evmRPC.Client)

	for _, evmChainConf := range axelarCfg.EVMConfig {
		if !evmChainConf.WithBridge {
			logger.Debug(fmt.Sprintf("RPC connection is disabled for EVM chain %s. Skipping...", evmChainConf.Name))
			continue
		}

		if _, found := rpcs[strings.ToLower(evmChainConf.Name)]; found {
			msg := fmt.Errorf("duplicate bridge configuration found for EVM chain %s", evmChainConf.Name)
			logger.Error(msg.Error())
			panic(msg)
		}

		rpc, err := evmRPC.NewClient(evmChainConf.RPCAddr)
		if err != nil {
			err = sdkerrors.Wrap(err, fmt.Sprintf("Failed to create an RPC connection for EVM chain %s. Verify your RPC config.", evmChainConf.Name))
			logger.Error(err.Error())
			panic(err)
		}
		// clean up evmRPC connection on process shutdown
		cleanupCommands = append(cleanupCommands, rpc.Close)

		rpcs[strings.ToLower(evmChainConf.Name)] = rpc
		logger.Info(fmt.Sprintf("Successfully connected to EVM bridge for chain %s", evmChainConf.Name))
	}

	evmMgr := evm.NewMgr(rpcs, cliCtx, b, logger, cdc)
	return evmMgr
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
func (f RWFile) WriteAll(bz []byte) error { return os.WriteFile(f.path, bz, RW) }
