package vald

import (
	"fmt"
	"path"
	"time"

	"github.com/axelarnetwork/c2d2/pkg/tendermint/client"
	tmEvents "github.com/axelarnetwork/c2d2/pkg/tendermint/events"
	sdkClient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/utils"
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
	ethTypes "github.com/axelarnetwork/axelar-core/x/ethereum/types"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetValdCommand returns the command to start vald
func GetValdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "vald-start",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			logger := serverCtx.Logger.With("module", "vald")
			cliCtx, err := sdkClient.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// dynamically adjust gas limit by simulating the tx first
			txf := tx.NewFactoryCLI(cliCtx, cmd.Flags()).WithSimulateAndExecute(true)

			hub, err := newHub()
			if err != nil {
				return err
			}

			axConf, valAddr := loadConfig()
			if valAddr == "" {
				return fmt.Errorf("validator address not set")
			}

			logger.Info("Start listening to events")
			listen(cliCtx, hub, txf, axConf, valAddr, logger)
			logger.Info("Shutting down")
			return nil
		},
	}
	setPersistentFlags(cmd)
	flags.AddTxFlagsToCmd(cmd)

	utils.OverwriteFlagDefaults(cmd, map[string]string{
		flags.FlagGasAdjustment: "1.2",
		flags.FlagBroadcastMode: flags.BroadcastSync,
	})

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

func newHub() (*tmEvents.Hub, error) {
	conf := client.Config{
		Address:  client.DefaultAddress,
		Endpoint: client.DefaultWSEndpoint,
	}

	c, err := client.NewConnectedClient(conf)
	if err != nil {
		return nil, err
	}

	hub := tmEvents.NewHub(c)
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

func listen(ctx sdkClient.Context, hub *tmEvents.Hub, txf tx.Factory, axelarCfg app.Config, valAddr string, logger log.Logger) {
	cdc := app.MakeEncodingConfig().Amino
	sender, err := ctx.Keyring.Key(axelarCfg.BroadcastConfig.From)
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to read broadcaster account info from keyring"))
	}
	ctx = ctx.
		WithFromAddress(sender.GetAddress()).
		WithFromName(sender.GetName())

	broadcaster := createBroadcaster(ctx, txf, axelarCfg, logger)
	tssMgr := createTSSMgr(broadcaster, ctx.FromAddress, axelarCfg, logger, valAddr, cdc)
	btcMgr := createBTCMgr(axelarCfg, broadcaster, ctx.FromAddress, logger, cdc)
	ethMgr := createETHMgr(axelarCfg, broadcaster, ctx.FromAddress, logger, cdc)

	keygenStart := events.MustSubscribe(hub, tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueStart)
	keygenMsg := events.MustSubscribe(hub, tssTypes.EventTypeKeygen, tssTypes.ModuleName, tssTypes.AttributeValueMsg)
	signStart := events.MustSubscribe(hub, tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueStart)
	signMsg := events.MustSubscribe(hub, tssTypes.EventTypeSign, tssTypes.ModuleName, tssTypes.AttributeValueMsg)

	btcConf := events.MustSubscribe(hub, btcTypes.EventTypeOutpointConfirmation, btcTypes.ModuleName, btcTypes.AttributeValueStart)

	ethDepConf := events.MustSubscribe(hub, ethTypes.EventTypeDepositConfirmation, ethTypes.ModuleName, ethTypes.AttributeValueStart)
	ethTokConf := events.MustSubscribe(hub, ethTypes.EventTypeTokenConfirmation, ethTypes.ModuleName, ethTypes.AttributeValueStart)

	js := []jobs.Job{
		events.Consume(keygenStart, tssMgr.ProcessKeygenStart),
		events.Consume(keygenMsg, tssMgr.ProcessKeygenMsg),
		events.Consume(signStart, tssMgr.ProcessSignStart),
		events.Consume(signMsg, tssMgr.ProcessSignMsg),
		events.Consume(btcConf, btcMgr.ProcessConfirmation),
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

func createBroadcaster(ctx sdkClient.Context, txf tx.Factory, axelarCfg app.Config, logger log.Logger) bcTypes.Broadcaster {
	pipeline := broadcast.NewPipelineWithRetry(10000, axelarCfg.MaxRetries, broadcast.LinearBackOff(axelarCfg.MinTimeout), logger)
	return broadcast.NewBroadcaster(ctx, txf, pipeline, logger)
}

func createTSSMgr(broadcaster bcTypes.Broadcaster, sender sdk.AccAddress, axelarCfg app.Config, logger log.Logger, valAddr string, cdc *codec.LegacyAmino) *tss.Mgr {
	create := func() (*tss.Mgr, error) {
		gg20client, err := tss.CreateTOFNDClient(axelarCfg.TssConfig.Host, axelarCfg.TssConfig.Port, logger)
		if err != nil {
			return nil, err
		}

		tssMgr := tss.NewMgr(gg20client, 2*time.Hour, valAddr, broadcaster, sender, logger, cdc)
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
