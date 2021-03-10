package main

import (
	"fmt"
	"os"
	"time"

	"github.com/axelarnetwork/c2d2/pkg/pubsub"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/client"
	"github.com/axelarnetwork/c2d2/pkg/tendermint/events"
	keyring "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/store/dbadapter"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/libs/pubsub/query"
	"github.com/tendermint/tendermint/rpc/client/http"
	tm "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/vald/broadcast"
	"github.com/axelarnetwork/axelar-core/cmd/vald/jobs"
	tss2 "github.com/axelarnetwork/axelar-core/cmd/vald/tss"
	tss "github.com/axelarnetwork/axelar-core/x/tss/types"
)

func getStartCommand(axConf app.Config, valAddr string, logger log.Logger) *cobra.Command {
	return &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			hub, err := newHub()
			if err != nil {
				return err
			}

			logger.Info("Start listening to events")
			err = listen(hub, axConf, valAddr, logger)
			if err != nil {
				return err
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
	tssMgr, err := createTSSMgr(axelarCfg, logger, valAddr)
	if err != nil {
		tmos.Exit(err.Error())
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

func createTSSMgr(axelarCfg app.Config, logger log.Logger, valAddr string) (*tss2.Mgr, error) {
	gg20client, err := tss2.CreateTOFNDClient(axelarCfg.TssConfig.Host, axelarCfg.TssConfig.Port, logger)
	if err != nil {
		return nil, err
	}
	keybase, err := keyring.NewKeyring(sdk.KeyringServiceName(), axelarCfg.ClientConfig.KeyringBackend, viper.GetString(cliHomeFlag), os.Stdin)
	if err != nil {
		return nil, err
	}
	abciClient, err := http.New(axelarCfg.TendermintNodeUri, "/websocket")
	if err != nil {
		return nil, err
	}
	b, err := broadcast.NewBroadcaster(app.MakeCodec(), keybase, dbadapter.Store{DB: dbm.NewMemDB()}, abciClient, axelarCfg.ClientConfig, logger)
	if err != nil {
		return nil, err
	}

	tssMgr := tss2.NewMgr(gg20client, 2*time.Hour, valAddr, b, logger)
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
