package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	nexus "github.com/axelarnetwork/axelar-core/x/nexus/exported"

	"github.com/axelarnetwork/axelar-core/x/nexus/keeper"
	"github.com/axelarnetwork/axelar-core/x/nexus/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		GetCommandChainMaintainers(queryRoute),
		GetCommandLatestDepositAddress(),
		GetCommandTransfersForChain(),
		GetCommandFee(),
		GetCommandTransferFee(),
	)

	return queryCmd
}

// GetCommandChainMaintainers returns the query for getting chain maintainers for the given chain
func GetCommandChainMaintainers(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain-maintainers [chain]",
		Short: "Returns the chain maintainers for the given chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			bz, _, err := clientCtx.Query(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryChainMaintainers, args[0]))
			if err != nil {
				return sdkerrors.Wrap(err, "couldn't resolve chain maintainers")
			}

			var res types.QueryChainMaintainersResponse
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

			return clientCtx.PrintProto(&res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCommandLatestDepositAddress returns the query for getting the latest deposit address of some recipient
func GetCommandLatestDepositAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-deposit-address [deposit chain] [recipient chain] [recipient address]",
		Short: "Query for account by address",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.LatestDepositAddress(cmd.Context(),
				&types.LatestDepositAddressRequest{
					DepositChain:   args[0],
					RecipientChain: args[1],
					RecipientAddr:  args[2],
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

// GetCommandTransfersForChain returns the query for the transfers for a given chain
func GetCommandTransfersForChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfers-for-chain [chain] [state (pending|archived|insufficient_amount)]",
		Short: "Query for account by address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			// the Key field is read as []byte{""} if the key flag is not set, so need to reset it manually
			if len(pageReq.Key) == 0 && pageReq.Offset > 0 {
				pageReq.Key = nil
			}

			transferState := nexus.TransferStateFromString(args[1])
			if transferState == nexus.TRANSFER_STATE_UNSPECIFIED {
				return fmt.Errorf("invalid transfer state %s provided", args[1])
			}

			res, err := queryClient.TransfersForChain(cmd.Context(),
				&types.TransfersForChainRequest{
					Chain:      args[0],
					State:      transferState,
					Pagination: pageReq,
				})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "transfers")

	return cmd
}

// GetCommandFee returns the query for the fee info of an asset registered on a chain
func GetCommandFee() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee [chain] [asset]",
		Short: "Query for fees registered for an asset on a chain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			res, err := queryClient.Fee(cmd.Context(),
				&types.FeeRequest{
					Chain: args[0],
					Asset: args[1],
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

// GetCommandTransferFee returns the query for the transfers for a given chain
func GetCommandTransferFee() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-fee [source-chain] [destination-chain] [amount]",
		Short: "Returns the fee incurred on a cross-chain transfer",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)

			amount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			res, err := queryClient.TransferFee(cmd.Context(),
				&types.TransferFeeRequest{
					SourceChain:      args[0],
					DestinationChain: args[1],
					Amount:           amount,
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
