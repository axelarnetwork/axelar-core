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

	"github.com/cosmos/cosmos-sdk/client"
	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	tmEvents "github.com/axelarnetwork/tm-events/events"
	"github.com/axelarnetwork/tm-events/pubsub"
	"github.com/axelarnetwork/tm-events/tendermint"
	"github.com/axelarnetwork/utils/jobs"

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
			if _, err := sdk.ValAddressFromBech32(valAddr); err != nil {
				return sdkerrors.Wrap(err, "invalid validator operator address")
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
	}
	utils.OverwriteFlagDefaults(cmd, values, true)

	// Only set default, not actual value, so it can be overwritten by env variable
	utils.OverwriteFlagDefaults(cmd, map[string]string{
		flags.FlagBroadcastMode:  flags.BroadcastBlock,
		flags.FlagChainID:        app.Name,
		flags.FlagGasPrices:      "0.00005uaxl",
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
	cmd.PersistentFlags().String("validator-addr", "", "the address of the validator operator, i.e axelarvaloper1..")
	cmd.PersistentFlags().String(flags.FlagChainID, app.Name, "The network chain ID")
}

func listen(clientCtx sdkClient.Context, txf tx.Factory, axelarCfg config.ValdConfig, valAddr string, recoveryJSON []byte, stateSource ReadWriter, logger log.Logger) {
	encCfg := app.MakeEncodingConfig()
	cdc := encCfg.Amino
	sender, err := clientCtx.Keyring.Key(clientCtx.From)
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to read broadcaster account info from keyring"))
	}
	clientCtx = clientCtx.
		WithFromAddress(sender.GetAddress()).
		WithFromName(sender.GetName())

	bc := createRefundableBroadcaster(txf, clientCtx, axelarCfg, logger)

	robustClient := tendermint.NewRobustClient(func() (rpcclient.Client, error) {
		cl, err := sdkClient.NewClientFromNode(clientCtx.NodeURI)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create a new client")
		}

		err = cl.Start()
		if err != nil {
			return nil, errors.Wrap(err, "unable to start client")
		}
		return cl, nil
	})
	tssMgr := createTSSMgr(bc, clientCtx, axelarCfg, logger, valAddr, cdc)
	if len(recoveryJSON) > 0 {
		if err = tssMgr.Recover(recoveryJSON); err != nil {
			panic(fmt.Errorf("unable to perform tss recovery: %v", err))
		}
	}

	evmMgr := createEVMMgr(axelarCfg, clientCtx, bc, logger, cdc)

	stateStore := NewStateStore(stateSource)
	startBlock, err := waitTillNetworkSync(axelarCfg, stateStore, robustClient, logger)
	if err != nil {
		panic(err)
	}

	eventBus := createEventBus(robustClient, startBlock, logger)
	var subscriptions []tmEvents.FilteredSubscriber
	subscribe := func(eventType, module, action string) tmEvents.FilteredSubscriber {
		return tmEvents.MustSubscribeWithAttributes(eventBus,
			eventType, module, sdk.Attribute{Key: sdk.AttributeKeyAction, Value: action})
	}

	blockHeaderSub := tmEvents.MustSubscribeBlockHeader(eventBus)
	subscriptions = append(subscriptions, blockHeaderSub)

	queryHeartBeat := createNewBlockEventQuery(tssTypes.EventTypeHeartBeat, tssTypes.ModuleName, tssTypes.AttributeValueSend)
	heartbeat, err := tmEvents.Subscribe(eventBus, queryHeartBeat)
	if err != nil {
		panic(fmt.Errorf("unable to subscribe with ack event query: %v", err))
	}
	subscriptions = append(subscriptions, heartbeat)

	keygenStart := subscribe(tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueStart)
	subscriptions = append(subscriptions, keygenStart)

	querySign := createNewBlockEventQuery(tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueStart)
	signStart, err := tmEvents.Subscribe(eventBus, querySign)
	if err != nil {
		panic(fmt.Errorf("unable to subscribe with sign event query: %v", err))
	}
	subscriptions = append(subscriptions, signStart)

	keygenMsg := subscribe(tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueMsg)
	subscriptions = append(subscriptions, keygenMsg)
	signMsg := subscribe(tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueMsg)
	subscriptions = append(subscriptions, signMsg)

	evmNewChain := subscribe(evmTypes.EventTypeNewChain, evmTypes.ModuleName, evmTypes.AttributeValueUpdate)
	subscriptions = append(subscriptions, evmNewChain)
	evmChainConf := subscribe(evmTypes.EventTypeChainConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	subscriptions = append(subscriptions, evmNewChain)
	evmDepConf := subscribe(evmTypes.EventTypeDepositConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	subscriptions = append(subscriptions, evmDepConf)
	evmTokConf := subscribe(evmTypes.EventTypeTokenConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	subscriptions = append(subscriptions, evmTokConf)
	evmTraConf := subscribe(evmTypes.EventTypeTransferKeyConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	subscriptions = append(subscriptions, evmTraConf)
	evmGatewayTxConf := subscribe(evmTypes.EventTypeGatewayTxConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)
	subscriptions = append(subscriptions, evmGatewayTxConf)

	eventCtx, cancelEventCtx := context.WithCancel(context.Background())
	// stop the jobs if process gets interrupted/terminated
	cleanupCommands = append(cleanupCommands, func() {
		logger.Info("stop listening for events...")
		cancelEventCtx()
		<-eventBus.Done()
		logger.Info("event listener stopped")
		logger.Info("stopping subscribers...")
		for _, subscription := range subscriptions {
			<-subscription.Done()
		}
		logger.Info("subscriptions stopped")
	})

	fetchEvents := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return nil
		case err := <-eventBus.FetchEvents(ctx):
			cancelEventCtx()
			return err
		}
	}

	processBlockHeader := func(event tmEvents.Event) error {
		tssMgr.ProcessNewBlockHeader(event.Height)
		return stateStore.SetState(event.Height)
	}

	js := []jobs.Job{
		fetchEvents,
		createJob(blockHeaderSub, processBlockHeader, cancelEventCtx, logger),
		createJob(heartbeat, tssMgr.ProcessHeartBeatEvent, cancelEventCtx, logger),
		createJob(keygenStart, tssMgr.ProcessKeygenStart, cancelEventCtx, logger),
		createJob(keygenMsg, tssMgr.ProcessKeygenMsg, cancelEventCtx, logger),
		createJob(signStart, tssMgr.ProcessSignStart, cancelEventCtx, logger),
		createJob(signMsg, tssMgr.ProcessSignMsg, cancelEventCtx, logger),
		createJob(evmNewChain, evmMgr.ProcessNewChain, cancelEventCtx, logger),
		createJob(evmChainConf, evmMgr.ProcessChainConfirmation, cancelEventCtx, logger),
		createJob(evmDepConf, evmMgr.ProcessDepositConfirmation, cancelEventCtx, logger),
		createJob(evmTokConf, evmMgr.ProcessTokenConfirmation, cancelEventCtx, logger),
		createJob(evmTraConf, evmMgr.ProcessTransferKeyConfirmation, cancelEventCtx, logger),
		createJob(evmGatewayTxConf, evmMgr.ProcessGatewayTxConfirmation, cancelEventCtx, logger),
	}

	mgr := jobs.NewMgr(eventCtx)
	mgr.AddJobs(js...)
	<-mgr.Done()
}

func createJob(sub tmEvents.FilteredSubscriber, processor func(event tmEvents.Event) error, cancel context.CancelFunc, logger log.Logger) jobs.Job {
	return func(ctx context.Context) error {
		processWithLog := func(e tmEvents.Event) {
			err := processor(e)
			if err != nil {
				logger.Error(err.Error())
			}
		}
		consume := tmEvents.Consume(sub, processWithLog)
		err := consume(ctx)
		if err != nil {
			cancel()
			return err
		}
		return nil
	}

}

// Wait until the node has synced with the network
// and then return the block height to start listening to TM events from
func waitTillNetworkSync(cfg config.ValdConfig, stateStore StateStore, tmClient tmEvents.BlockInfoClient, logger log.Logger) (int64, error) {
	rpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	syncInfo, err := tmClient.LatestSyncInfo(rpcCtx)
	if err != nil {
		return 0, err
	}

	cachedHeight, err := stateStore.GetState()
	if err != nil {
		logger.Info(fmt.Sprintf("failed to retrieve the cached block height, using the latest: %s", err.Error()))
		cachedHeight = 0
	} else {
		logger.Info(fmt.Sprintf("retrieved cached block height %d", cachedHeight))
		cachedHeight++ // Skip the block that might have already been executed
	}

	// cached height must not be more than one block ahead of the node
	if cachedHeight > syncInfo.LatestBlockHeight+1 {
		return 0, fmt.Errorf("cached block height %d is ahead of the node height %d", cachedHeight, syncInfo.LatestBlockHeight)
	}

	// If the block height is older than the allowed time, then wait for the node to sync
	for syncInfo.LatestBlockTime.Add(cfg.MaxBlockTime).Before(time.Now()) {
		logger.Info(fmt.Sprintf("node height %d is old, waiting for a recent block", syncInfo.LatestBlockHeight))
		time.Sleep(cfg.MaxBlockTime)

		syncInfo, err = tmClient.LatestSyncInfo(rpcCtx)
		if err != nil {
			return 0, err
		}
	}

	nodeHeight := syncInfo.LatestBlockHeight
	startBlock := cachedHeight

	logger.Info(fmt.Sprintf("node is synced, node height: %d", nodeHeight))

	if startBlock != 0 && nodeHeight-startBlock > cfg.MaxOutOfSyncHeight {
		logger.Info(fmt.Sprintf("cached block height %d is too old, starting from the latest block", startBlock))
		startBlock = 0
	}

	return startBlock, nil
}

func createNewBlockEventQuery(eventType, module, action string) tmEvents.Query {
	return tmEvents.Query{
		TMQuery: tmEvents.NewBlockHeaderEventQuery(eventType).MatchModule(module).MatchAction(action).Build(),
		Predicate: func(e tmEvents.Event) bool {
			return e.Type == eventType && e.Attributes[sdk.AttributeKeyModule] == module && e.Attributes[sdk.AttributeKeyAction] == action
		},
	}
}

func createEventBus(client *tendermint.RobustClient, startBlock int64, logger log.Logger) *tmEvents.Bus {
	notifier := tmEvents.NewBlockNotifier(client, logger).StartingAt(startBlock)
	return tmEvents.NewEventBus(tmEvents.NewBlockSource(client, notifier, logger), pubsub.NewBus, logger)
}

func createRefundableBroadcaster(txf tx.Factory, ctx sdkClient.Context, axelarCfg config.ValdConfig, logger log.Logger) broadcasterTypes.Broadcaster {
	pipeline := broadcaster.NewPipelineWithRetry(10000, axelarCfg.MaxRetries, utils2.LinearBackOff(axelarCfg.MinTimeout), logger)
	return broadcaster.WithRefund(broadcaster.NewBroadcaster(txf, ctx, pipeline, axelarCfg.BatchThreshold, axelarCfg.BatchSizeLimit, logger))
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

		rpc, err := evmRPC.NewClient(evmChainConf.RPCAddr, evmChainConf.EnableRPCDetection)
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
