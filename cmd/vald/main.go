package main

import (
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/axelarnetwork/c2d2/pkg/pubsub"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/client"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/events"
	"github.com/cosmos/cosmos-sdk/client/flags"
	keyring "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/libs/pubsub/query"
	"github.com/tendermint/tendermint/rpc/client/http"
	tm "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/vald/broadcast"
	tss2 "github.com/axelarnetwork/axelar-core/cmd/vald/tss"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

const cliHomeFlag = "clihome"

func main() {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(app.Bech32PrefixAccAddr, app.Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(app.Bech32PrefixValAddr, app.Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(app.Bech32PrefixConsAddr, app.Bech32PrefixConsPub)
	config.Seal()

	rootCmd := &cobra.Command{
		Use:   "vald",
		Short: "Validator Daemon ",
	}

	rootCmd.PersistentFlags().String(cliHomeFlag, app.DefaultCLIHome, "directory for cli config and data")
	_ = viper.BindPFlag(cliHomeFlag, rootCmd.Flags().Lookup(cliHomeFlag))

	rootCmd.PersistentFlags().String("tofnd-host", "", "host name for tss daemon")
	_ = viper.BindPFlag("tofnd_host", rootCmd.PersistentFlags().Lookup("tofnd-host"))

	rootCmd.PersistentFlags().String("tofnd-port", "50051", "port for tss daemon")
	_ = viper.BindPFlag("tofnd_port", rootCmd.PersistentFlags().Lookup("tofnd-port"))

	// rootCmd.PersistentFlags().String(flags.FlagNode, "tcp://localhost:26657", "<host>:<port> to Tendermint RPC interface for this chain")
	// _ = viper.BindPFlag(flags.FlagNode, rootCmd.PersistentFlags().Lookup(flags.FlagNode))

	// rootCmd.PersistentFlags().Uint64("gas", uint64(flags.DefaultGasLimit),
	// 	fmt.Sprintf("gas limit to set per-transaction (default %d)", flags.DefaultGasLimit))
	// _ = viper.BindPFlag("gas", rootCmd.PersistentFlags().Lookup("gas"))

	startCommand := getStartCommand()
	rootCmd.AddCommand(flags.PostCommands(startCommand)...)

	executor := cli.PrepareMainCmd(rootCmd, "AX", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		tmos.Exit(err.Error())
	}
}

func getStartCommand() *cobra.Command {
	return &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {

			conf := client.Config{
				Address:  client.DefaultAddress,
				Endpoint: client.DefaultWSEndpoint,
			}
			logger := log.NewTMLogger(os.Stdout).With("external", "main")

			c, err := client.NewConnectedClient(conf)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}
			hub := events.NewHub(c)

			logger.Info("Start listening to events")
			axConf := loadConfig()

			err = listen(&hub, axConf, logger)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}

			logger.Info("Shutting down")
			return nil
		},
	}
}

func loadConfig() app.Config {
	// need to merge in cli config because axelard now has its own broadcasting client
	conf := app.DefaultConfig()
	homeDir := viper.GetString(cli.HomeFlag)
	cliHomeDir := viper.GetString(cliHomeFlag)
	cliCfgFile := path.Join(cliHomeDir, "config", "config.toml")
	viper.SetConfigFile(cliCfgFile)
	if err := viper.MergeInConfig(); err != nil {
		panic(err)
	}
	cfgFile := path.Join(homeDir, "config", "config.toml")
	viper.SetConfigFile(cfgFile)

	if err := viper.Unmarshal(&conf); err != nil {
		panic(err)
	}
	// for some reason gas is not being filled
	conf.Gas = viper.GetInt("gas")

	return conf
}

func listen(hub *events.Hub, axelarCfg app.Config, logger log.Logger) error {
	// start a gRPC client
	tofndServerAddress := axelarCfg.TssConfig.Host + ":" + axelarCfg.TssConfig.Port
	logger.Info(fmt.Sprintf("initiate connection to tofnd gRPC server: %s", tofndServerAddress))
	conn, err := grpc.Dial(tofndServerAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		tmos.Exit(err.Error())
	}
	logger.Debug("successful connection to tofnd gRPC server")
	gg20client := tofnd.NewGG20Client(conn)

	keybase, err := keyring.NewKeyring(sdk.KeyringServiceName(), axelarCfg.ClientConfig.KeyringBackend, viper.GetString(cliHomeFlag), os.Stdin)
	if err != nil {
		tmos.Exit(err.Error())
	}
	abciClient, err := http.New(axelarCfg.TendermintNodeUri, "/websocket")
	if err != nil {
		tmos.Exit(err.Error())
	}

	b, err := broadcast.NewBroadcaster(app.MakeCodec(), keybase, dbadapter.Store{DB: dbm.NewMemDB()}, abciClient, axelarCfg.ClientConfig, logger)
	if err != nil {
		tmos.Exit(err.Error())
	}

	tssMgr := tss2.NewTSSMgr(gg20client, 2*time.Hour, axelarCfg.BroadcastConfig.From, b, logger)

	keygen, err := subscribeToEvent(hub, tss.EventTypeKeygen, tss.ModuleName)
	if err != nil {
		return err
	}
	sign, err := subscribeToEvent(hub, tss.EventTypeSign, tss.ModuleName)
	if err != nil {
		return err
	}

	processors := []func(chan<- error){
		func(e chan<- error) { tssMgr.ProcessKeygen(keygen, e) },
		func(e chan<- error) { tssMgr.ProcessSign(sign, e) }}

	waitFor(processors, logger)

	return nil
}

func waitFor(processors []func(chan<- error), logger log.Logger) {
	errChan := make(chan error, 100)
	wg1 := &sync.WaitGroup{}
	wg1.Add(len(processors))

	for _, process := range processors {
		go func(f func(chan<- error)) {
			defer wg1.Done()
			f(errChan)
		}(process)
	}

	wg2 := &sync.WaitGroup{}
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		for err := range errChan {
			logger.Error(err.Error())
		}
	}()
	// when all events are processed, wait until all errors are logged
	wg1.Wait()
	close(errChan)
	wg2.Done()
}

func subscribeToEvent(hub *events.Hub, eventType string, module string) (pubsub.Subscriber, error) {
	bus, err := hub.Subscribe(query.MustParse(fmt.Sprintf("%s='%s' AND %s.%s='%s'",
		tm.EventTypeKey, tm.EventTx, eventType, sdk.AttributeKeyModule, module)))
	if err != nil {
		return nil, err
	}
	subscriber, err := bus.Subscribe()
	if err != nil {
		return nil, err
	}
	return subscriber, nil
}
