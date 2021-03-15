package main

import (
	"fmt"
	"os"
	"time"

	"github.com/axelarnetwork/c2d2/pkg/pubsub"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/client"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/events"
	keyring "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/pubsub/query"
	tm "github.com/tendermint/tendermint/types"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/vald/broadcast"
	"github.com/axelarnetwork/axelar-core/cmd/vald/jobs"
	tss2 "github.com/axelarnetwork/axelar-core/cmd/vald/tss"
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
				logger.Error("validator address not set")
				os.Exit(1)
			}

			logger.Info("Start listening to events")
			err = listen(hub, axConf, valAddr, logger)
			if err != nil {
				logger.Error(err.Error())
				os.Exit(1)
			}
			logger.Info("Shutting down")
			return nil
		},
	}
}

func newHub() (*events.Hub, error) {
	conf := client.Config{
		Address:  client.DefaultAddress,
		Endpoint: client.DefaultWSEndpoint,
	}

	c, err := client.NewConnectedClient(conf)
	if err != nil {
		return nil, err
	}

	hub := events.NewHub(c)
	return &hub, nil
}

func listen(hub *events.Hub, axelarCfg app.Config, valAddr string, logger log.Logger) error {
	broadcaster, sender, err := createBroadcaster(axelarCfg, logger)
	if err != nil {
		return err
	}

	tssMgr, err := createTSSMgr(broadcaster, sender, axelarCfg, logger, valAddr)
	if err != nil {
		return err
	}

	keygen, err := subscribeToEvent(hub, tss.EventTypeKeygen, tss.ModuleName)
	if err != nil {
		return err
	}
	sign, err := subscribeToEvent(hub, tss.EventTypeSign, tss.ModuleName)
	if err != nil {
		return err
	}

	js := []jobs.Job{
		func(e chan<- error) { tssMgr.ProcessKeygen(keygen, e) },
		func(e chan<- error) { tssMgr.ProcessSign(sign, e) }}

	// errGroup runs async processes and cancels their context if ANY of them returns an error.
	// Here, we don't want to stop on errors, but simply log it and continue, so errGroup doesn't cut it
	logErr := func(err error) { logger.Error(err.Error()) }
	mgr := jobs.NewMgr(logErr)
	mgr.AddJobs(js...)
	mgr.Wait()

	return nil
}

func createBroadcaster(axelarCfg app.Config, logger log.Logger) (*broadcast.Broadcaster, sdk.AccAddress, error) {
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
	return b, info.GetAddress(), nil
}

func createTSSMgr(broadcaster *broadcast.Broadcaster, defaultSender sdk.AccAddress, axelarCfg app.Config, logger log.Logger, valAddr string) (*tss2.Mgr, error) {
	gg20client, err := tss2.CreateTOFNDClient(axelarCfg.TssConfig.Host, axelarCfg.TssConfig.Port, logger)
	if err != nil {
		return nil, err
	}

	xboBroadcaster := broadcast.WithExponentialBackoff(broadcaster, axelarCfg.MinTimeout, axelarCfg.MaxRetries)
	tssMgr := tss2.NewMgr(gg20client, 2*time.Hour, valAddr, xboBroadcaster, defaultSender, logger)
	return tssMgr, nil
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
