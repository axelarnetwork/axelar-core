package main

import (
	"fmt"
	"os"
	"time"

	"github.com/axelarnetwork/c2d2/pkg/tendermint/client"
	tmEvents "github.com/axelarnetwork/c2d2/pkg/tendermint/events"
	keyring "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/vald/broadcast"
	"github.com/axelarnetwork/axelar-core/cmd/vald/broadcast/types"
	"github.com/axelarnetwork/axelar-core/cmd/vald/btc"
	"github.com/axelarnetwork/axelar-core/cmd/vald/events"
	"github.com/axelarnetwork/axelar-core/cmd/vald/jobs"
	tss2 "github.com/axelarnetwork/axelar-core/cmd/vald/tss"
	btcTypes "github.com/axelarnetwork/axelar-core/x/bitcoin/types"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

func getStartCommand(logger log.Logger) *cobra.Command {
	return &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			hub, err := newHub()
			if err != nil {
				return err
			}

			axConf, valAddr := loadConfig()
			if valAddr == "" {
				return fmt.Errorf("validator address not set")
			}

			logger.Info("Start listening to events")
			listen(hub, axConf, valAddr, logger)
			logger.Info("Shutting down")
			return nil
		},
	}
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

func listen(hub *tmEvents.Hub, axelarCfg app.Config, valAddr string, logger log.Logger) {
	broadcaster, sender := createBroadcaster(axelarCfg, logger)
	tssMgr := createTSSMgr(broadcaster, sender, axelarCfg, logger, valAddr)
	btcMgr := createBTCMgr(axelarCfg, broadcaster, logger, valAddr)

	keygenStart := events.MustSubscribe(hub, tss.EventTypeKeygen, tss.ModuleName, tss.AttributeValueStart)
	keygenMsg := events.MustSubscribe(hub, tss.EventTypeKeygen, tss.ModuleName, tss.AttributeValueMsg)
	signStart := events.MustSubscribe(hub, tss.EventTypeSign, tss.ModuleName, tss.AttributeValueStart)
	signMsg := events.MustSubscribe(hub, tss.EventTypeSign, tss.ModuleName, tss.AttributeValueMsg)

	btcVer := events.MustSubscribe(hub, btcTypes.EventTypeVerification, btcTypes.ModuleName, btcTypes.AttributeValueStart)

	js := []jobs.Job{
		events.NewJob(keygenStart, tssMgr.ProcessKeygenStart),
		events.NewJob(keygenMsg, tssMgr.ProcessKeygenMsg),
		events.NewJob(signStart, tssMgr.ProcessSignStart),
		events.NewJob(signMsg, tssMgr.ProcessSignMsg),
		events.NewJob(btcVer, btcMgr.ProcessVerification),
	}

	// errGroup runs async processes and cancels their context if ANY of them returns an error.
	// Here, we don't want to stop on errors, but simply log it and continue, so errGroup doesn't cut it
	logErr := func(err error) { logger.Error(err.Error()) }
	mgr := jobs.NewMgr(logErr)
	mgr.AddJobs(js...)
	mgr.Wait()
}

func createBroadcaster(axelarCfg app.Config, logger log.Logger) (types.Broadcaster, sdk.AccAddress) {
	create := func() (types.Broadcaster, sdk.AccAddress, error) {
		rpc, err := broadcast.NewClient(utils.GetTxEncoder(app.MakeCodec()), axelarCfg.TendermintNodeUri)
		if err != nil {
			return nil, sdk.AccAddress{}, err
		}
		keybase, err := keyring.NewKeyring(sdk.KeyringServiceName(), axelarCfg.ClientConfig.KeyringBackend, viper.GetString(cliHomeFlag), os.Stdin)
		if err != nil {
			return nil, sdk.AccAddress{}, err
		}
		info, err := keybase.Get(axelarCfg.BroadcastConfig.From)
		if err != nil {
			return nil, sdk.AccAddress{}, err
		}
		signer, err := broadcast.NewSigner(keybase, info, axelarCfg.BroadcastConfig.KeyringPassphrase)
		if err != nil {
			return nil, sdk.AccAddress{}, err
		}
		b, err := broadcast.NewBroadcaster(signer, rpc, axelarCfg.ClientConfig, logger)
		if err != nil {
			return nil, sdk.AccAddress{}, err
		}

		backoffBroadcaster := broadcast.WithBackoff(b, broadcast.Linear, axelarCfg.MinTimeout, axelarCfg.MaxRetries)
		return backoffBroadcaster, info.GetAddress(), nil
	}
	b, addr, err := create()
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to create broadcaster"))
	}
	return b, addr
}

func createTSSMgr(broadcaster types.Broadcaster, defaultSender sdk.AccAddress, axelarCfg app.Config, logger log.Logger, valAddr string) *tss2.Mgr {
	create := func() (*tss2.Mgr, error) {
		gg20client, err := tss2.CreateTOFNDClient(axelarCfg.TssConfig.Host, axelarCfg.TssConfig.Port, logger)
		if err != nil {
			return nil, err
		}

		tssMgr := tss2.NewMgr(gg20client, 2*time.Hour, valAddr, broadcaster, defaultSender, logger)
		return tssMgr, nil
	}
	mgr, err := create()
	if err != nil {
		panic(sdkerrors.Wrap(err, "failed to create tss manager"))
	}
	return mgr
}

func createBTCMgr(axelarCfg app.Config, b types.Broadcaster, logger log.Logger, valAddr string) *btc.Mgr {
	rpc, err := btcTypes.NewRPCClient(axelarCfg.BtcConfig, logger)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
	// BTC bridge opens a grpc connection. Clean it up on process shutdown
	tmos.TrapSignal(logger, rpc.Shutdown)

	btcMgr := btc.NewMgr(rpc, valAddr, b, logger)
	return btcMgr
}
