package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/evm/keeper"
	"github.com/axelarnetwork/axelar-core/x/evm/types"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/slices"
)

const (
	activated   = "activated"
	deactivated = "deactivated"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	evmQueryCmd := &cobra.Command{
		Use:                        "evm",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	evmQueryCmd.AddCommand(
		getCmdAddress(),
		getCmdAxelarGatewayAddress(),
		getCmdTokenAddress(queryRoute),
		getCmdDepositState(),
		getCmdBytecode(),
		getCmdQueryBatchedCommands(),
		getCmdLatestBatchedCommands(),
		getCmdPendingCommands(),
		getCmdCommand(),
		getCmdBurnerInfo(),
		getCmdChains(),
		getCmdConfirmationHeight(),
		getCmdERC20Tokens(),
		getCmdTokenInfo(),
		getCmdEvent(),
		getParams(),
	)

	return evmQueryCmd

}

// getCmdAddress returns the query for an EVM chain address
func getCmdAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address [chain]",
		Short: "Returns the EVM address",
		Args:  cobra.ExactArgs(1),
	}
	keyID := cmd.Flags().String("key-id", "", "the ID of the key to get the address for")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		queryClient := types.NewQueryServiceClient(clientCtx)

		req := types.KeyAddressRequest{
			Chain: utils.NormalizeString(args[0]),
			KeyID: multisig.KeyID(*keyID),
		}

		res, err := queryClient.KeyAddress(cmd.Context(), &req)
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdTokenAddress returns the query for an EVM chain master address that owns the AxelarGateway contract
func getCmdTokenAddress(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-address [chain]",
		Short: fmt.Sprintf("Query a token address by by either %s or %s", keeper.BySymbol, keeper.ByAsset),
		Args:  cobra.ExactArgs(1),
	}

	symbol := cmd.Flags().String(keeper.BySymbol, "", "lookup token by symbol")
	asset := cmd.Flags().String(keeper.ByAsset, "", "lookup token by asset name")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cliCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		var bz []byte
		switch {
		case *symbol != "" && *asset == "":
			bz, _, err = cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QTokenAddressBySymbol, args[0], *symbol))
		case *symbol == "" && *asset != "":
			bz, _, err = cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QTokenAddressByAsset, args[0], *asset))
		default:
			return fmt.Errorf("lookup must be either by asset name or symbol")
		}

		if err != nil {
			return err
		}

		var res types.QueryTokenAddressResponse
		types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

		return cliCtx.PrintProto(&res)
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdDepositState returns the query for an ERC20 deposit transaction state
func getCmdDepositState() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "deposit-state [chain] [txID] [burner address]",
		Short:      "Query the state of a deposit transaction",
		Deprecated: "will be removed in the future release",
		Args:       cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			txID := common.HexToHash(args[1])
			burnerAddress := common.HexToAddress(args[2])

			queryClient := types.NewQueryServiceClient(cliCtx)

			res, err := queryClient.DepositState(cmd.Context(), &types.DepositStateRequest{
				Chain: nexus.ChainName(chain),
				Params: &types.QueryDepositStateParams{
					TxID:          types.Hash(txID),
					BurnerAddress: types.Address(burnerAddress),
				},
			})
			if err != nil {
				return err
			}

			return cliCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdAxelarGatewayAddress returns the query for the AxelarGateway contract address
func getCmdAxelarGatewayAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway-address [chain]",
		Short: "Query the Axelar Gateway contract address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.GatewayAddress(cmd.Context(),
				&types.GatewayAddressRequest{
					Chain: utils.NormalizeString(chain),
				})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdBytecode fetches the bytecodes of an EVM contract
func getCmdBytecode() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bytecode [chain] [contract]",
		Short: "Fetch the bytecode of an EVM contract [contract] for chain [chain]",
		Long:  fmt.Sprintf("Fetch the bytecode of an EVM contract [contract] for chain [chain]. The value for [contract] can be either '%s' or '%s'.", keeper.BCToken, keeper.BCBurner),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			contract := args[1]

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.Bytecode(cmd.Context(),
				&types.BytecodeRequest{
					Chain:    utils.NormalizeString(chain),
					Contract: utils.NormalizeString(contract),
				})
			if err != nil {
				return sdkerrors.Wrapf(err, types.ErrFBytecode, contract)
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdQueryBatchedCommands returns the query to get the batched commands
func getCmdQueryBatchedCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batched-commands [chain] [batchedCommandsID]",
		Short: "Get the signed batched commands that can be wrapped in an EVM transaction to be executed in Axelar Gateway",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			idHex := args[1]

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.BatchedCommands(cmd.Context(),
				&types.BatchedCommandsRequest{
					Chain: utils.NormalizeString(chain),
					Id:    utils.NormalizeString(idHex),
				})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdLatestBatchedCommands returns the query to get the latest batched commands
func getCmdLatestBatchedCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-batched-commands [chain]",
		Short: "Get the latest batched commands that can be wrapped in an EVM transaction to be executed in Axelar Gateway",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.BatchedCommands(cmd.Context(),
				&types.BatchedCommandsRequest{
					Chain: utils.NormalizeString(chain),
				})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdPendingCommands returns the query to get the list of commands not yet added to a batch
func getCmdPendingCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-commands [chain]",
		Short: "Get the list of commands not yet added to a batch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.PendingCommands(cmd.Context(),
				&types.PendingCommandsRequest{
					Chain: utils.NormalizeString(args[0]),
				})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdCommand returns the query to get the command with the given ID on the specified chain
func getCmdCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "command [chain] [id]",
		Short: "Get information about an EVM gateway command given a chain and the command ID",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.Command(cmd.Context(),
				&types.CommandRequest{
					Chain: args[0],
					ID:    args[1],
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdBurnerInfo returns the query to get the burner info for the specified address
func getCmdBurnerInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burner-info [deposit address]",
		Short: "Get information about a burner address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.BurnerInfo(cmd.Context(),
				&types.BurnerInfoRequest{
					Address: types.Address(common.HexToAddress(args[0])),
				})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdChains returns the query to get all EVM chains
func getCmdChains() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chains",
		Short: "Return the supported EVM chains by status",
		Args:  cobra.ExactArgs(0),
	}

	status := cmd.Flags().String("status", "", fmt.Sprintf("the chain status [%s|%s]", activated, deactivated))

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		queryClient := types.NewQueryServiceClient(clientCtx)

		var chainStatus types.ChainStatus
		switch *status {
		case "":
			chainStatus = types.StatusUnspecified
		case activated:
			chainStatus = types.Activated
		case deactivated:
			chainStatus = types.Deactivated
		default:
			return fmt.Errorf("unrecognized chain status %s", *status)
		}

		res, err := queryClient.Chains(cmd.Context(), &types.ChainsRequest{
			Status: chainStatus,
		})
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// getCmdConfirmationHeight returns the query to get the minimum confirmation height for the given chain
func getCmdConfirmationHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirmation-height [chain]",
		Short: "Returns the minimum confirmation height for the given chain",
		Args:  cobra.ExactArgs(1),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		queryClient := types.NewQueryServiceClient(clientCtx)

		res, err := queryClient.ConfirmationHeight(cmd.Context(),
			&types.ConfirmationHeightRequest{
				Chain: args[0],
			})
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdEvent returns the query to an event for a chain based on the event's txID
func getCmdEvent() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event [chain] [event-id]",
		Short: "Returns an event for the given chain",
		Args:  cobra.ExactArgs(2),
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		chain := utils.NormalizeString(args[0])
		eventID := utils.NormalizeString(args[1])

		queryClient := types.NewQueryServiceClient(clientCtx)

		res, err := queryClient.Event(cmd.Context(),
			&types.EventRequest{
				Chain:   chain,
				EventId: eventID,
			})
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdERC20Tokens returns the query to get the ERC20 tokens for a given chain
func getCmdERC20Tokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "erc20-tokens [chain]",
		Short: "Returns the ERC20 tokens for the given chain",
		Args:  cobra.ExactArgs(1),
	}
	tokenType := cmd.Flags().String("token-type", "", "the token type [external|internal]")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		queryClient := types.NewQueryServiceClient(clientCtx)

		var tokenTypeEnum types.TokenType
		switch *tokenType {
		case "":
			tokenTypeEnum = types.Unspecified
		case "internal":
			tokenTypeEnum = types.Internal
		case "external":
			tokenTypeEnum = types.External
		}

		res, err := queryClient.ERC20Tokens(cmd.Context(),
			&types.ERC20TokensRequest{
				Chain: args[0],
				Type:  tokenTypeEnum,
			})
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// getCmdTokenInfo returns the query to get the details for an ERC20 token
func getCmdTokenInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token-info [chain]",
		Short: fmt.Sprintf("Returns the info of token by either %s, %s, or %s", keeper.BySymbol, keeper.ByAsset, keeper.ByAddress),
		Args:  cobra.ExactArgs(1),
	}

	symbol := cmd.Flags().String(keeper.BySymbol, "", "lookup token by symbol")
	asset := cmd.Flags().String(keeper.ByAsset, "", "lookup token by asset name")
	address := cmd.Flags().String(keeper.ByAddress, "", "lookup token by address")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}

		if !exactlyOneIsFilled(*symbol, *asset, *address) {
			return fmt.Errorf("lookup must be either by asset name, symbol, or address")
		}

		var req types.TokenInfoRequest
		switch {
		case *asset != "":
			req = types.TokenInfoRequest{
				Chain:  args[0],
				FindBy: &types.TokenInfoRequest_Asset{Asset: *asset},
			}
		case *symbol != "":
			req = types.TokenInfoRequest{
				Chain:  args[0],
				FindBy: &types.TokenInfoRequest_Symbol{Symbol: *symbol},
			}
		case *address != "":
			req = types.TokenInfoRequest{
				Chain:  args[0],
				FindBy: &types.TokenInfoRequest_Address{Address: *address},
			}
		}

		queryClient := types.NewQueryServiceClient(clientCtx)
		res, err := queryClient.TokenInfo(cmd.Context(), &req)
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func getParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params [chain]",
		Short: "Returns the params for the evm module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.ParamsRequest{Chain: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func exactlyOneIsFilled(flags ...string) bool {
	nonEmptyFlags := slices.Reduce(flags, 0, func(count int, f string) int {
		if f != "" {
			return count + 1
		}
		return count
	})
	return nonEmptyFlags == 1
}
