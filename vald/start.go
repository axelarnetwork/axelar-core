package vald

import (
	"context"
	"fmt"
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
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/utils"
	"github.com/axelarnetwork/axelar-core/sdk-utils/broadcast"
	errors2 "github.com/axelarnetwork/axelar-core/utils/errors"
	"github.com/axelarnetwork/axelar-core/vald/config"
	"github.com/axelarnetwork/axelar-core/vald/evm"
	evmRPC "github.com/axelarnetwork/axelar-core/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/vald/multisig"
	grpc "github.com/axelarnetwork/axelar-core/vald/tofnd_grpc"
	"github.com/axelarnetwork/axelar-core/vald/tss"
	evmTypes "github.com/axelarnetwork/axelar-core/x/evm/types"
	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	tmEvents "github.com/axelarnetwork/tm-events/events"
	"github.com/axelarnetwork/tm-events/pubsub"
	"github.com/axelarnetwork/tm-events/tendermint"
	"github.com/axelarnetwork/utils/funcs"
	"github.com/axelarnetwork/utils/jobs"
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
			v := serverCtx.Viper

			if err := v.BindPFlag("tss.tofnd-host", cmd.PersistentFlags().Lookup("tofnd-host")); err != nil {
				return err
			}
			if err := v.BindPFlag("tss.tofnd-port", cmd.PersistentFlags().Lookup("tofnd-port")); err != nil {
				return err
			}
			if err := v.BindPFlag("tss.tofnd-dial-timeout", cmd.PersistentFlags().Lookup("tofnd-dial-timeout")); err != nil {
				return err
			}

			cliCtx, err := sdkClient.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// dynamically adjust gas limit by simulating the tx first
			txf := tx.NewFactoryCLI(cliCtx, cmd.Flags()).WithSimulateAndExecute(true)

			return runVald(cliCtx, txf, logger, v)
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
		flags.FlagGasPrices:      "0.007uaxl",
		flags.FlagKeyringBackend: "file",
	}, false)

	return cmd
}

func runVald(cliCtx sdkClient.Context, txf tx.Factory, logger log.Logger, viper *viper.Viper) error {
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

	valdConf := config.DefaultValdConfig()
	viper.RegisterAlias("broadcast.max_timeout", "rpc.timeout_broadcast_tx_commit")
	if err := viper.Unmarshal(&valdConf); err != nil {
		panic(err)
	}

	valAddr, err := sdk.ValAddressFromBech32(viper.GetString("validator-addr"))
	if err != nil {
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

	fPath := filepath.Join(valdHome, "state.json")
	stateSource := NewRWFile(fPath)

	logger.Info("start listening to events")
	listen(cliCtx, txf, valdConf, valAddr, stateSource, logger)
	logger.Info("shutting down")
	return nil
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
	cmd.PersistentFlags().String("tofnd-dial-timeout", defaultConf.DialTimeout.String(), "dialup timeout to the tss daemon")
	cmd.PersistentFlags().String("validator-addr", "", "the address of the validator operator, i.e axelarvaloper1..")
	cmd.PersistentFlags().String(flags.FlagChainID, app.Name, "The network chain ID")
}

func listen(clientCtx sdkClient.Context, txf tx.Factory, axelarCfg config.ValdConfig, valAddr sdk.ValAddress, stateSource ReadWriter, logger log.Logger) {
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
	tssMgr := createTSSMgr(bc, clientCtx, axelarCfg, logger, valAddr.String(), cdc)

	evmMgr := createEVMMgr(axelarCfg, clientCtx, bc, logger, cdc, valAddr)
	multisigMgr := createMultisigMgr(bc, clientCtx, axelarCfg, logger, valAddr)

	nodeHeight, err := waitTillNetworkSync(axelarCfg, robustClient, logger)
	if err != nil {
		panic(err)
	}

	stateStore := NewStateStore(stateSource)
	startBlock, err := getStartBlock(axelarCfg, stateStore, nodeHeight, robustClient, logger)
	if err != nil {
		panic(err)
	}

	eventBus := createEventBus(robustClient, startBlock, logger)
	subscribe := func(eventType, module, action string) <-chan tmEvents.ABCIEventWithHeight {
		return eventBus.Subscribe(func(e tmEvents.ABCIEventWithHeight) bool {
			event := tmEvents.Map(e)

			return event.Type == eventType && event.Attributes[sdk.AttributeKeyModule] == module && event.Attributes[sdk.AttributeKeyAction] == action
		})
	}

	var blockHeight int64
	blockHeaderSub := eventBus.Subscribe(func(event tmEvents.ABCIEventWithHeight) bool {
		if event.Height != blockHeight {
			blockHeight = event.Height
			return true
		}
		return false
	})

	heartbeat := subscribe(tssTypes.EventTypeHeartBeat, tssTypes.ModuleName, tssTypes.AttributeValueSend)

	evmNewChain := eventBus.Subscribe(tmEvents.Filter[*evmTypes.ChainAdded]())
	evmDepConf := eventBus.Subscribe(tmEvents.Filter[*evmTypes.ConfirmDepositStarted]())
	evmTokConf := eventBus.Subscribe(tmEvents.Filter[*evmTypes.ConfirmTokenStarted]())
	evmTraConf := eventBus.Subscribe(tmEvents.Filter[*evmTypes.ConfirmKeyTransferStarted]())
	evmGatewayTxConf := eventBus.Subscribe(tmEvents.Filter[*evmTypes.ConfirmGatewayTxStarted]())

	multisigKeygen := eventBus.Subscribe(tmEvents.Filter[*multisigTypes.KeygenStarted]())
	multisigSigning := eventBus.Subscribe(tmEvents.Filter[*multisigTypes.SigningStarted]())

	eventCtx, cancelEventCtx := context.WithCancel(context.Background())
	mgr := jobs.NewMgr(eventCtx)

	// stop the jobs if process gets interrupted/terminated
	cleanupCommands = append(cleanupCommands, func() {
		logger.Info("stop listening for events...")
		cancelEventCtx()
		<-eventBus.Done()
		logger.Info("event listener stopped")
		logger.Info("stopping subscribers...")
		<-mgr.Done()
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
		return stateStore.SetState(event.Height)
	}

	js := []jobs.Job{
		fetchEvents,
		createJob(blockHeaderSub, processBlockHeader, cancelEventCtx, logger),
		createJob(heartbeat, tssMgr.ProcessHeartBeatEvent, cancelEventCtx, logger),
		createJobTyped(evmNewChain, evmMgr.ProcessNewChain, cancelEventCtx, logger),
		createJobTyped(evmDepConf, evmMgr.ProcessDepositConfirmation, cancelEventCtx, logger),
		createJobTyped(evmTokConf, evmMgr.ProcessTokenConfirmation, cancelEventCtx, logger),
		createJobTyped(evmTraConf, evmMgr.ProcessTransferKeyConfirmation, cancelEventCtx, logger),
		createJobTyped(evmGatewayTxConf, evmMgr.ProcessGatewayTxConfirmation, cancelEventCtx, logger),
		createJobTyped(multisigKeygen, multisigMgr.ProcessKeygenStarted, cancelEventCtx, logger),
		createJobTyped(multisigSigning, multisigMgr.ProcessSigningStarted, cancelEventCtx, logger),
	}

	mgr.AddJobs(js...)
	<-mgr.Done()
}

func createJob(sub <-chan tmEvents.ABCIEventWithHeight, processor func(event tmEvents.Event) error, cancel context.CancelFunc, logger log.Logger) jobs.Job {
	return func(ctx context.Context) error {
		processWithLog := func(e tmEvents.ABCIEventWithHeight) {
			err := processor(tmEvents.Map(e))
			if err != nil {
				logger.Error(err.Error(), errors2.KeyVals(err))
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

func createJobTyped[T proto.Message](sub <-chan tmEvents.ABCIEventWithHeight, processor func(event T) error, cancel context.CancelFunc, logger log.Logger) jobs.Job {
	return func(ctx context.Context) error {
		processWithLog := func(e tmEvents.ABCIEventWithHeight) {
			event := funcs.Must(sdk.ParseTypedEvent(e.Event)).(T)
			err := processor(event)
			if err != nil {
				logger.Error(err.Error(), errors2.KeyVals(err))
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

// Wait until the node has synced with the network and return the node height
func waitTillNetworkSync(cfg config.ValdConfig, tmClient tmEvents.SyncInfoClient, logger log.Logger) (int64, error) {
	for {
		rpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		syncInfo, err := tmClient.LatestSyncInfo(rpcCtx)
		cancel()
		if err != nil {
			return 0, err
		}

		// If the block height is older than the allowed time, then wait for the node to sync
		if syncInfo.LatestBlockTime.Add(cfg.MaxLatestBlockAge).After(time.Now()) {
			return syncInfo.LatestBlockHeight, nil
		}

		logger.Info(fmt.Sprintf("node height %d is old, waiting for a recent block", syncInfo.LatestBlockHeight))
		time.Sleep(cfg.MaxLatestBlockAge)
	}
}

// Return the block height to start listening to TM events from
func getStartBlock(cfg config.ValdConfig, stateStore StateStore, nodeHeight int64, tmClient tmEvents.SyncInfoClient, logger log.Logger) (int64, error) {
	storedHeight, err := stateStore.GetState()
	if err != nil {
		logger.Info(fmt.Sprintf("failed to retrieve the stored block height, using the latest: %s", err.Error()))
		storedHeight = 0
	} else {
		logger.Info(fmt.Sprintf("retrieved stored block height %d", storedHeight))
	}

	// stored height must not be larger than node height
	if storedHeight > nodeHeight {
		return 0, fmt.Errorf("stored block height %d is ahead of the node height %d", storedHeight, nodeHeight)
	}

	logger.Info(fmt.Sprintf("node is synced, node height: %d", nodeHeight))

	startBlock := storedHeight
	if startBlock != 0 {
		// The block at the stored height might have already been processed by vald, so skip it
		startBlock++
	}

	if startBlock != 0 && nodeHeight-startBlock > cfg.MaxBlocksBehindLatest {
		logger.Info(fmt.Sprintf("stored block height %d is too old, starting from the latest block", startBlock))
		startBlock = 0
	}

	return startBlock, nil
}

func createEventBus(client *tendermint.RobustClient, startBlock int64, logger log.Logger) *tmEvents.Bus {
	notifier := tmEvents.NewBlockNotifier(client, logger).StartingAt(startBlock)
	return tmEvents.NewEventBus(tmEvents.NewBlockSource(client, notifier, logger), pubsub.NewBus[tmEvents.ABCIEventWithHeight](), logger)
}

func createRefundableBroadcaster(txf tx.Factory, ctx sdkClient.Context, axelarCfg config.ValdConfig, logger log.Logger) broadcast.Broadcaster {
	broadcaster := broadcast.WithStateManager(ctx, txf, logger, broadcast.WithResponseTimeout(axelarCfg.BroadcastConfig.MaxTimeout))
	broadcaster = broadcast.WithRetry(broadcaster, axelarCfg.MaxRetries, axelarCfg.MinSleepBeforeRetry, logger)
	broadcaster = broadcast.Batched(broadcaster, axelarCfg.BatchThreshold, axelarCfg.BatchSizeLimit, logger)
	broadcaster = broadcast.WithRefund(broadcaster)
	broadcaster = broadcast.SuppressExecutionErrs(broadcaster, logger)

	return broadcaster
}

func createMultisigMgr(broadcaster broadcast.Broadcaster, cliCtx client.Context, axelarCfg config.ValdConfig, logger log.Logger, valAddr sdk.ValAddress) *multisig.Mgr {
	conn, err := grpc.Connect(axelarCfg.TssConfig.Host, axelarCfg.TssConfig.Port, axelarCfg.TssConfig.DialTimeout, logger)
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to create multisig manager"))
	}
	logger.Debug("successful connection to tofnd gRPC server")

	return multisig.NewMgr(tofnd.NewMultisigClient(conn), cliCtx, valAddr, logger, broadcaster, timeout)
}

func createTSSMgr(broadcaster broadcast.Broadcaster, cliCtx client.Context, axelarCfg config.ValdConfig, logger log.Logger, valAddr string, cdc *codec.LegacyAmino) *tss.Mgr {
	create := func() (*tss.Mgr, error) {
		conn, err := tss.Connect(axelarCfg.TssConfig.Host, axelarCfg.TssConfig.Port, axelarCfg.TssConfig.DialTimeout, logger)
		if err != nil {
			return nil, err
		}
		logger.Debug("successful connection to tofnd gRPC server")

		// creates client to communicate with the external tofnd process service
		multiSigClient := tofnd.NewMultisigClient(conn)

		tssMgr := tss.NewMgr(multiSigClient, cliCtx, 2*time.Hour, valAddr, broadcaster, logger, cdc)

		return tssMgr, nil
	}

	mgr, err := create()
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to create tss manager"))
	}

	return mgr
}

func createEVMMgr(axelarCfg config.ValdConfig, cliCtx sdkClient.Context, b broadcast.Broadcaster, logger log.Logger, cdc *codec.LegacyAmino, valAddr sdk.ValAddress) *evm.Mgr {
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
		logger.Debug(fmt.Sprintf("created JSON-RPC client of type %T", rpc),
			"chain", evmChainConf.Name,
			"url", evmChainConf.RPCAddr,
		)

		// clean up evmRPC connection on process shutdown
		cleanupCommands = append(cleanupCommands, rpc.Close)

		rpcs[strings.ToLower(evmChainConf.Name)] = rpc
		logger.Info(fmt.Sprintf("Successfully connected to EVM bridge for chain %s", evmChainConf.Name))
	}

	evmMgr := evm.NewMgr(rpcs, cliCtx, b, logger, cdc, valAddr)
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
