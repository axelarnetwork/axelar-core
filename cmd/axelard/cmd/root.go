package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"cosmossdk.io/log"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	cmtcfg "github.com/cometbft/cometbft/config"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/app/params"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/utils"
	"github.com/axelarnetwork/axelar-core/config"
	axelarnet "github.com/axelarnetwork/axelar-core/x/axelarnet/exported"
	"github.com/axelarnetwork/utils/funcs"
)

var (
	minGasPrice = "0.007" + axelarnet.NativeAsset
)

// NewRootCmd creates a new root command for axelard. It is called once in the
// main function.
func NewRootCmd() (*cobra.Command, params.EncodingConfig) {
	app.SetConfig()

	encodingConfig := app.MakeEncodingConfig()

	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastSync).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("")

	rootCmd := &cobra.Command{
		Use:   app.Name + "d",
		Short: "Axelar App",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx = initClientCtx.WithCmdContext(cmd.Context())
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = clientconfig.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			// This needs to go after ReadFromClientConfig, as that function
			// sets the RPC client needed for SIGN_MODE_TEXTUAL. This sign mode
			// is only available if the client is online.
			if !initClientCtx.Offline {
				enabledSignModes := append(tx.DefaultSignModes, signing.SignMode_SIGN_MODE_TEXTUAL)
				txConfigOpts := tx.ConfigOptions{
					EnabledSignModes:           enabledSignModes,
					TextualCoinMetadataQueryFn: authtxconfig.NewGRPCCoinMetadataQueryFn(initClientCtx),
				}
				txConfig, err := tx.NewTxConfigWithOptions(
					initClientCtx.Codec,
					txConfigOpts,
				)
				if err != nil {
					return err
				}

				initClientCtx = initClientCtx.WithTxConfig(txConfig)
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			axelarAppTemplate, axelarAppConfig := initAppConfig()

			cmConfig := cmtcfg.DefaultConfig()
			err = server.InterceptConfigsPreRunHandler(cmd, axelarAppTemplate, axelarAppConfig, cmConfig)
			if err != nil {
				return err
			}

			serverCtx, err := server.InterceptConfigsAndCreateContext(cmd, axelarAppTemplate, axelarAppConfig, cmConfig)
			if err != nil {
				return err
			}

			// InterceptConfigsPreRunHandler initializes a console logger with an improper time format with no way of changing the config,
			// so we need to overwrite the logger
			logger, err := sdkLoggerWithTimeFormat(serverCtx, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			serverCtx.Logger = logger.With(log.ModuleKey, "server")

			err = server.SetCmdServerContext(cmd, serverCtx)
			if err != nil {
				return err
			}

			// we don't have direct access to the definition of the start command, so this is the only place we can add additional seeds
			if cmd.Use == "start" {
				if err := extendSeeds(cmd); err != nil {
					return err
				}
			}

			return nil
		},
	}

	initRootCmd(rootCmd, encodingConfig)

	// we "pre"-instantiate the application for getting the injected/configured encoding configuration
	// note, this is not necessary when using app wiring, as depinject can be directly used (see root_v2.go)
	temp := tempDir()
	// cleanup temp dir after we are done with the tempApp, so we don't leave behind a
	// new temporary directory for every invocation. See https://github.com/CosmWasm/wasmd/issues/2017
	defer funcs.MustNoErr(os.RemoveAll(temp))

	tempApp := app.NewAxelarApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		encodingConfig,
		simtestutil.NewAppOptionsWithFlagHome(temp),
		[]wasmkeeper.Option{},
	)
	defer func() {
		funcs.MustNoErr(tempApp.Close())
	}()

	// add keyring to autocli opts
	autoCliOpts := tempApp.AutoCliOpts()
	autoCliOpts.ClientCtx = initClientCtx
	funcs.MustNoErr(autoCliOpts.EnhanceRootCommand(rootCmd))

	utils.OverwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagBroadcastMode:    flags.BroadcastSync,
		flags.FlagChainID:          app.Name,
		flags.FlagGasPrices:        minGasPrice,
		flags.FlagKeyringBackend:   "file",
		flags.FlagSkipConfirmation: "true",
	}, false)

	return rootCmd, encodingConfig
}

func extendSeeds(cmd *cobra.Command) error {
	serverCtx := server.GetServerContextFromCmd(cmd)
	seeds, err := config.ReadSeeds(serverCtx.Viper)
	if errors.As(err, &viper.ConfigFileNotFoundError{}) {
		serverCtx.Logger.Info("file seeds.toml not found")
		return nil
	}
	if err != nil {
		return err
	}

	serverCtx.Config = config.MergeSeeds(serverCtx.Config, seeds)
	if err := server.SetCmdServerContext(cmd, serverCtx); err != nil {
		return err
	}

	serverCtx.Logger.Info(fmt.Sprintf("adding %d seeds from seeds.toml", len(seeds)))

	return nil
}

// sdkLoggerWithTimeFormat creates an SDK logger with RFC3339 time format
// It reads the log level and format from the server context.
func sdkLoggerWithTimeFormat(ctx *server.Context, out io.Writer) (log.Logger, error) {
	var opts []log.Option
	if ctx.Viper.GetString(flags.FlagLogFormat) == flags.OutputFormatJSON {
		opts = append(opts, log.OutputJSONOption())
	}
	opts = append(opts,
		log.ColorOption(!ctx.Viper.GetBool(flags.FlagLogNoColor)),
		// We use CometBFT flag (cmtcli.TraceFlag) for trace logging.
		log.TraceOption(ctx.Viper.GetBool(server.FlagTrace)),
		log.TimeFormatOption(time.RFC3339),
	)

	// check and set filter level or keys for the logger if any
	logLvlStr := ctx.Viper.GetString(flags.FlagLogLevel)
	if logLvlStr == "" {
		return log.NewLogger(out, opts...), nil
	}

	logLvl, err := zerolog.ParseLevel(logLvlStr)
	switch {
	case err != nil:
		// If the log level is not a valid zerolog level, then we try to parse it as a key filter.
		filterFunc, err := log.ParseLogLevel(logLvlStr)
		if err != nil {
			return nil, err
		}

		opts = append(opts, log.FilterOption(filterFunc))
	default:
		opts = append(opts, log.LevelOption(logLvl))
	}

	return log.NewLogger(out, opts...), nil
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	type AxelarAppConfig struct {
		serverconfig.Config
	}

	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()

	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In simapp, we set the default min gas price.
	srvCfg.MinGasPrices = minGasPrice

	axelarAppConfig := AxelarAppConfig{
		Config: *srvCfg,
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate

	return customAppTemplate, axelarAppConfig
}

func tempDir() string {
	dir, err := os.MkdirTemp("", "axelard")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}

	return dir
}
