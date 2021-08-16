package vald

import (
	"context"
	"io"
	"math/rand"
	"time"

	"github.com/axelarnetwork/axelar-core/app"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

var alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12234567890")

const (
	idLen int = 10

	defaultTimeout time.Duration = 2 * time.Hour
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetTofndPingCommand returns the command to ping tofnd
func GetTofndPingCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "tofnd-ping",
		RunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			logger := server.ZeroLogWrapper{Logger: zerolog.New(io.Discard)}

			axelarCfg := app.DefaultConfig()
			if err := serverCtx.Viper.Unmarshal(&axelarCfg); err != nil {
				panic(err)
			}

			timeout, err := time.ParseDuration(serverCtx.Viper.GetString("context-timeout"))
			if err != nil {
				return err
			}

			gg20client, err := tss.CreateTOFNDClient(axelarCfg.TssConfig.Host, axelarCfg.TssConfig.Port, axelarCfg.TssConfig.DialTimeout, logger)
			if err != nil {
				logger.Error("failed to reach tofnd: %s", err.Error())
				return nil
			}

			logger.Info("successfully created grpc")

			grpcCtx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			request := &tofnd.KeyPresenceRequest{
				// we do not need to look for a key ID that exists to obtain a successful healthcheck,
				// all we need to do is obtain err == nil && response != FAIL
				KeyUid: randomID(idLen),
			}

			response, err := gg20client.KeyPresence(grpcCtx, request)

			if err != nil {
				logger.Error("failed to invoke tofnd grpc: %s", err.Error())
				return nil
			}

			logger.Info("grpc call successful")

			if response.Response == tofnd.KeyPresenceResponse_RESPONSE_FAIL {
				logger.Error("tofnd healthcheck failed")
				return nil
			}

			logger.Info("healthcheck passed, we are good!")
			return nil
		},
	}

	defaultConf := tssTypes.DefaultConfig()
	cmd.PersistentFlags().String("tofnd-host", defaultConf.Host, "host name for tss daemon")
	cmd.PersistentFlags().String("tofnd-port", defaultConf.Port, "port for tss daemon")
	cmd.PersistentFlags().String("context-timeout", defaultTimeout.String(), "context timeout for the grpc")
	return cmd
}

func randomID(length int) string {
	buffer := make([]rune, length)
	for i := range buffer {
		buffer[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(buffer)
}
