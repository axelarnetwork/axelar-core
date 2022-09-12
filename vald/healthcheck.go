package vald

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/vald/config"
	"github.com/axelarnetwork/axelar-core/vald/tss"
	"github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	snapshotTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
)

const (
	keyID      = "testkey"
	tokenDenom = "uaxl"
	minBalance = 5000000
	timeout    = time.Hour

	flagSkipTofnd       = "skip-tofnd"
	flagSkipBroadcaster = "skip-broadcaster"
	flagSkipOperator    = "skip-operator"
	flagOperatorAddr    = "operator-addr"
	flagTofndHost       = "tofnd-host"
	flagTofndPort       = "tofnd-port"
)

// GetHealthCheckCommand returns the command to execute a node health check
func GetHealthCheckCommand() *cobra.Command {
	var skipTofnd bool
	var skipBroadcaster bool
	var skipOperator bool

	cmd := &cobra.Command{
		Use: "health-check",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			serverCtx := server.GetServerContextFromCmd(cmd)

			ok := execCheck(context.Background(), clientCtx, serverCtx, "tofnd", skipTofnd, checkTofnd) &&
				execCheck(cmd.Context(), clientCtx, serverCtx, "broadcaster", skipBroadcaster, checkBroadcaster)

			// enforce a non-zero exit code in case health checks fail without printing cobra output
			if !ok {
				os.Exit(1)
			}

			return nil
		},
	}

	defaultConf := tssTypes.DefaultConfig()
	cmd.PersistentFlags().String(flagTofndHost, defaultConf.Host, "host name for tss daemon")
	cmd.PersistentFlags().String(flagTofndPort, defaultConf.Port, "port for tss daemon")
	cmd.PersistentFlags().String(flagOperatorAddr, "", "operator address")
	cmd.PersistentFlags().BoolVar(&skipTofnd, flagSkipTofnd, false, "skip tofnd check")
	cmd.PersistentFlags().BoolVar(&skipBroadcaster, flagSkipBroadcaster, false, "skip broadcaster check")
	cmd.PersistentFlags().BoolVar(&skipOperator, flagSkipOperator, false, "skip operator check")

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

type checkCmd func(ctx context.Context, clientCtx client.Context, serverCtx *server.Context) error

func execCheck(ctx context.Context, clientCtx client.Context, serverCtx *server.Context, name string, skip bool, check checkCmd) bool {
	if skip {
		fmt.Printf("%s check: skipped\n", name)
		return true
	}

	err := check(ctx, clientCtx, serverCtx)
	if err != nil {
		fmt.Printf("%s check: failed (%s)\n", name, err.Error())
		return false
	}

	fmt.Printf("%s check: passed\n", name)
	return true
}

func checkTofnd(ctx context.Context, clientCtx client.Context, serverCtx *server.Context) error {
	valdCfg := config.DefaultValdConfig()
	if err := serverCtx.Viper.Unmarshal(&valdCfg); err != nil {
		panic(err)
	}

	nopLogger := server.ZeroLogWrapper{Logger: zerolog.New(io.Discard)}

	conn, err := tss.Connect(valdCfg.TssConfig.Host, valdCfg.TssConfig.Port, valdCfg.TssConfig.DialTimeout, nopLogger)
	if err != nil {
		return fmt.Errorf("failed to reach tofnd: %s", err.Error())
	}
	nopLogger.Debug("successful connection to tofnd gRPC server")

	// creates client to communicate with the external tofnd process multisig service
	client := tofnd.NewMultisigClient(conn)

	grpcCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	request := &tofnd.KeyPresenceRequest{
		// we do not need to look for a key ID that exists to obtain a successful healthcheck,
		// all we need to do is obtain err == nil && response != FAIL
		// TODO: this kind of check should have its own dedicated GRPC
		KeyUid: keyID,
	}

	response, err := client.KeyPresence(grpcCtx, request)

	if err != nil {
		return fmt.Errorf("failed to invoke tofnd grpc: %s", err.Error())
	}

	if response.Response == tofnd.RESPONSE_FAIL ||
		response.Response == tofnd.RESPONSE_UNSPECIFIED {
		return fmt.Errorf("obtained FAIL response, tofnd not properly configured")
	}

	return nil
}

func checkBroadcaster(ctx context.Context, clientCtx client.Context, serverCtx *server.Context) error {
	str := serverCtx.Viper.GetString(flagOperatorAddr)
	if str == "" {
		return fmt.Errorf("no operator address specified")
	}
	operator, err := sdk.ValAddressFromBech32(str)
	if err != nil {
		return err
	}

	bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", snapshotTypes.QuerierRoute, keeper.QProxy, operator.String()))
	if err != nil {
		return err
	}

	reply := struct {
		Address string `json:"address"`
		Status  string `json:"status"`
	}{}
	if err := json.Unmarshal(bz, &reply); err != nil {
		return err
	}

	broadcaster, err := sdk.AccAddressFromBech32(reply.Address)
	if err != nil {
		return err
	}

	if reply.Status != "active" {
		return fmt.Errorf("broadcaster for operator %s not active", operator.String())
	}

	queryClient := bankTypes.NewQueryClient(clientCtx)
	params := bankTypes.NewQueryBalanceRequest(broadcaster, tokenDenom)

	grpcCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	res, err := queryClient.Balance(grpcCtx, params)
	if err != nil {
		return err
	}

	if res.Balance.Amount.LTE(sdk.NewInt(minBalance)) {
		return fmt.Errorf("broadcaster does not have enough funds (minimum balance is %d%s)", minBalance, tokenDenom)
	}

	return nil
}
