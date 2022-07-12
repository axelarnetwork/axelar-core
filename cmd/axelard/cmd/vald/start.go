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
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/utils"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/broadcast"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/config"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/evm"
	evmRPC "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/evm/rpc"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/multisig"
	grpc "github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tofnd_grpc"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss"
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

			cliCtx, err := sdkClient.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// dynamically adjust gas limit by simulating the tx first
			txf := tx.NewFactoryCLI(cliCtx, cmd.Flags()).WithSimulateAndExecute(true)

			return runVald(cmd.Context(), cliCtx, txf, logger, v)
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

func runVald(ctx context.Context, cliCtx sdkClient.Context, txf tx.Factory, logger log.Logger, viper *viper.Viper) error {
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

	valAddr := viper.GetString("validator-addr")
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
	recoveryFile := viper.GetString("tofnd-recovery")
	if recoveryFile != "" {
		var err error
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
	listen(ctx, cliCtx, txf, valdConf, valAddr, recoveryJSON, stateSource, logger)
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
	cmd.PersistentFlags().String("tofnd-recovery", "", "json file with recovery request")
	cmd.PersistentFlags().String("validator-addr", "", "the address of the validator operator, i.e axelarvaloper1..")
	cmd.PersistentFlags().String(flags.FlagChainID, app.Name, "The network chain ID")
}

func listen(ctx context.Context, clientCtx sdkClient.Context, txf tx.Factory, axelarCfg config.ValdConfig, valAddr string, recoveryJSON []byte, stateSource ReadWriter, logger log.Logger) {
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

	// Refresh keys after the node has synced
	if err := tssMgr.RefreshKeys(ctx); err != nil {
		panic(err)
	}

	blockNotifier := tmEvents.NewBlockNotifier(robustClient, logger).StartingAt(startBlock)
	eventBus := tmEvents.NewEventBus(tmEvents.NewBlockSource(robustClient, blockNotifier, logger), pubsub.NewBus[abci.Event](), logger)

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
			return err
		}
	}

	processBlockHeader := func(blockHeight int64) error {
		tssMgr.ProcessNewBlockHeader(blockHeight)
		return stateStore.SetState(blockHeight)
	}

	blocks, blockErrs := blockNotifier.BlockHeights(ctx)

	js := []jobs.Job{
		fetchEvents,
		consume(blockErrs, funcs.Identity[error]),
		consume(blocks, processBlockHeader),
		consume(
			eventBus.Subscribe(matchEvent(tssTypes.EventTypeHeartBeat, tssTypes.ModuleName, tssTypes.AttributeValueSend)),
			mapEventTo(tssMgr.ProcessHeartBeatEvent)),
		consume(
			eventBus.Subscribe(matchEvent(tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueStart)),
			mapEventTo(tssMgr.ProcessKeygenStart)),
		consume(
			eventBus.Subscribe(matchEvent(tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueMsg)),
			mapEventTo(tssMgr.ProcessKeygenMsg)),
		consume(
			eventBus.Subscribe(matchEvent(tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueStart)),
			mapEventTo(tssMgr.ProcessSignStart)),
		consume(
			eventBus.Subscribe(matchEvent(tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueMsg)),
			mapEventTo(tssMgr.ProcessSignMsg)),
		consume(
			eventBus.Subscribe(matchEvent(evmTypes.EventTypeNewChain, evmTypes.ModuleName, evmTypes.AttributeValueUpdate)),
			mapEventTo(evmMgr.ProcessNewChain)),
		consume(
			eventBus.Subscribe(matchEvent(evmTypes.EventTypeDepositConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)),
			mapEventTo(evmMgr.ProcessDepositConfirmation)),
		consume(
			eventBus.Subscribe(matchEvent(evmTypes.EventTypeTokenConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)),
			mapEventTo(evmMgr.ProcessTokenConfirmation)),
		consume(
			eventBus.Subscribe(matchEvent(evmTypes.EventTypeTransferKeyConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)),
			mapEventTo(evmMgr.ProcessTransferKeyConfirmation)),
		consume(
			eventBus.Subscribe(matchEvent(evmTypes.EventTypeGatewayTxConfirmation, evmTypes.ModuleName, evmTypes.AttributeValueStart)),
			mapEventTo(evmMgr.ProcessGatewayTxConfirmation)),
		consume(
			eventBus.Subscribe(eventFilter[*multisigTypes.KeygenStarted]()),
			funcs.Compose(parseEvent[multisigTypes.KeygenStarted], multisigMgr.ProcessKeygenStarted)),
	}

	mgr.AddJobs(js...)
	go func() {
		select {
		case <-eventCtx.Done():
			return
		case err := <-mgr.Errs():
			logger.Error(errors.Wrap(err, "job failed").Error())
			cancelEventCtx()
		}
	}()
	<-mgr.Done()
}

func mapEventTo(f func(event tmEvents.Event) error) func(event abci.Event) error {
	return mapEventTo(f)
}

func consume[T any](sub <-chan T, processor func(T) error) jobs.Job {
	return func(ctx context.Context) error {
		errs := make(chan error, 1)
		go func() {
			for {
				select {
				case <-ctx.Done():
					errs <- ctx.Err()
					return
				case x, ok := <-sub:
					if !ok {
						errs <- nil
					}

					if err := processor(x); err != nil {
						errs <- err
						return
					}
				}
			}
		}()
		return <-errs
	}
}

func parseEvent[T proto.Message](event abci.Event) T {
	return funcs.Must(sdk.ParseTypedEvent(event)).(T)
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

func createRefundableBroadcaster(txf tx.Factory, ctx sdkClient.Context, axelarCfg config.ValdConfig, logger log.Logger) broadcast.Broadcaster {
	broadcaster := broadcast.WithStateManager(ctx, txf, logger, broadcast.WithResponseTimeout(axelarCfg.BroadcastConfig.MaxTimeout))
	broadcaster = broadcast.WithRetry(broadcaster, axelarCfg.MaxRetries, axelarCfg.MinSleepBeforeRetry, logger)
	broadcaster = broadcast.Batched(broadcaster, axelarCfg.BatchThreshold, axelarCfg.BatchSizeLimit, logger)
	broadcaster = broadcast.WithRefund(broadcaster)
	broadcaster = broadcast.SuppressExecutionErrs(broadcaster, logger)

	return broadcaster
}

func createMultisigMgr(broadcaster broadcast.Broadcaster, cliCtx client.Context, axelarCfg config.ValdConfig, logger log.Logger, valAddr string) *multisig.Mgr {
	conn, err := grpc.Connect(axelarCfg.TssConfig.Host, axelarCfg.TssConfig.Port, axelarCfg.TssConfig.DialTimeout, logger)
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to create multisig manager"))
	}
	logger.Debug("successful connection to tofnd gRPC server")

	return multisig.NewMgr(tofnd.NewMultisigClient(conn), cliCtx, funcs.Must(sdk.ValAddressFromBech32(valAddr)), logger, broadcaster, timeout)
}

func createTSSMgr(broadcaster broadcast.Broadcaster, cliCtx client.Context, axelarCfg config.ValdConfig, logger log.Logger, valAddr string, cdc *codec.LegacyAmino) *tss.Mgr {
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

func createEVMMgr(axelarCfg config.ValdConfig, cliCtx client.Context, b broadcast.Broadcaster, logger log.Logger, cdc *codec.LegacyAmino) *evm.Mgr {
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

func eventFilter[T proto.Message]() func(e abci.Event) bool {
	return func(e abci.Event) bool {
		typedEvent, err := sdk.ParseTypedEvent(e)
		if err != nil {
			return false
		}

		return proto.MessageName(typedEvent) == proto.MessageName(*new(T))
	}
}

func matchEvent(eventType, module, action string) func(e abci.Event) bool {
	return func(e abci.Event) bool {
		event := tmEvents.Map(e)
		return event.Type == eventType && event.Attributes[sdk.AttributeKeyModule] == module && event.Attributes[sdk.AttributeKeyAction] == action
	}
}
