package vald

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/config"
	"github.com/axelarnetwork/axelar-core/cmd/axelard/cmd/vald/tss"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/axelarnetwork/axelar-core/x/snapshot/keeper"
	"github.com/axelarnetwork/axelar-core/x/snapshot/types"
	snapshotTypes "github.com/axelarnetwork/axelar-core/x/snapshot/types"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	tssTypes "github.com/axelarnetwork/axelar-core/x/tss/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
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

var (
	allGood bool = true
)

// GetHealthCheckCommand returns the command to execute a node health check
func GetHealthCheckCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "health-check",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			serverCtx := server.GetServerContextFromCmd(cmd)

			execCheck(context.Background(), clientCtx, serverCtx, flagSkipTofnd, checkTofnd)
			execCheck(cmd.Context(), clientCtx, serverCtx, flagSkipBroadcaster, checkBroadcaster)
			execCheck(nil, clientCtx, serverCtx, flagSkipOperator, checkOperator)

			// enforce a non-zero exit code in case health checks fail without printing cobra output
			if !allGood {
				os.Exit(1)
			}

			return nil
		},
	}

	defaultConf := tssTypes.DefaultConfig()
	cmd.PersistentFlags().String(flagTofndHost, defaultConf.Host, "host name for tss daemon")
	cmd.PersistentFlags().String(flagTofndPort, defaultConf.Port, "port for tss daemon")
	cmd.PersistentFlags().String(flagOperatorAddr, "", "broadcaster address")
	cmd.PersistentFlags().Bool(flagSkipTofnd, false, "skip tofnd check")
	cmd.PersistentFlags().Bool(flagSkipBroadcaster, false, "skip broadcaster check")
	cmd.PersistentFlags().Bool(flagSkipOperator, false, "skip operator check")

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

type checkCmd func(ctx context.Context, clientCtx client.Context, serverCtx *server.Context) error

func execCheck(ctx context.Context, clientCtx client.Context, serverCtx *server.Context, flag string, cmd checkCmd) {
	fmt.Printf("%s check: ", strings.TrimPrefix(flag, "skip-"))
	if serverCtx.Viper.GetBool(flag) {
		fmt.Println("skipped")
		return
	}

	err := cmd(ctx, clientCtx, serverCtx)
	if err != nil {
		fmt.Printf("failed (%s)\n", err.Error())
		allGood = false
		return
	}

	fmt.Println("passed")
}

func checkTofnd(ctx context.Context, clientCtx client.Context, serverCtx *server.Context) error {
	valdCfg := config.DefaultValdConfig()
	if err := serverCtx.Viper.Unmarshal(&valdCfg); err != nil {
		panic(err)
	}

	nopLogger := server.ZeroLogWrapper{Logger: zerolog.New(io.Discard)}
	gg20client, err := tss.CreateTOFNDClient(valdCfg.TssConfig.Host, valdCfg.TssConfig.Port, valdCfg.TssConfig.DialTimeout, nopLogger)
	if err != nil {
		return fmt.Errorf("failed to reach tofnd: %s", err.Error())
	}

	grpcCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	request := &tofnd.KeyPresenceRequest{
		// we do not need to look for a key ID that exists to obtain a successful healthcheck,
		// all we need to do is obtain err == nil && response != FAIL
		// TODO: this kind of check should have its own dedicated GRPC
		KeyUid: keyID,
	}

	response, err := gg20client.KeyPresence(grpcCtx, request)

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
	json.Unmarshal(bz, &reply)

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

func checkOperator(_ context.Context, clientCtx client.Context, serverCtx *server.Context) error {
	addr := serverCtx.Viper.GetString(flagOperatorAddr)
	if addr == "" {
		return fmt.Errorf("no operator address specified")
	}

	bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s", snapshotTypes.QuerierRoute, keeper.QValidators))
	if err != nil {
		return err
	}

	var resValidators types.QueryValidatorsResponse
	types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &resValidators)

	for _, v := range resValidators.Validators {
		if v.OperatorAddress == addr {
			if v.TssIllegibilityInfo.Jailed ||
				v.TssIllegibilityInfo.MissedTooManyBlocks ||
				v.TssIllegibilityInfo.NoProxyRegistered ||
				v.TssIllegibilityInfo.Tombstoned ||
				v.TssIllegibilityInfo.TssSuspended {
				return fmt.Errorf("health check to operator %s failed due to the following issues: %v",
					addr, string(snapshotTypes.ModuleCdc.MustMarshalJSON(&v.TssIllegibilityInfo)))
			}
			return nil
		}
	}

	return fmt.Errorf("operator address %s not found amongst current set of validators", addr)
}
