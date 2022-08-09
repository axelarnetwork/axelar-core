package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/multisig/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	multisigQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	multisigQueryCmd.AddCommand(
		GetCmdKeyID(queryRoute),
		GetCmdNextKeyID(queryRoute),
		GetCmdKey(queryRoute),
		GetCmdKeygenCandidates(queryRoute),
	)

	return multisigQueryCmd
}

// GetCmdKeyID returns the key ID assigned to a given chain
func GetCmdKeyID(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key-id [chain]",
		Short: "Returns the key ID assigned to a given chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := utils.NormalizeString(args[0])
			queryClient := types.NewQueryServiceClient(clientCtx)
			res, err := queryClient.KeyID(cmd.Context(),
				&types.KeyIDRequest{
					Chain: chain,
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

// GetCmdNextKeyID returns the key ID assigned for the next rotation on a given chain
func GetCmdNextKeyID(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next-key-id [chain]",
		Short: "Returns the key ID assigned for the next rotation on a given chain and for the given key role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := utils.NormalizeString(args[0])
			queryClient := types.NewQueryServiceClient(clientCtx)
			res, err := queryClient.NextKeyID(cmd.Context(),
				&types.NextKeyIDRequest{
					Chain: chain,
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

// GetCmdKey returns the key of the given ID
func GetCmdKey(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key [key-id]",
		Short: "Returns the key of the given ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			keyID := multisig.KeyID(utils.NormalizeString(args[0]))
			queryClient := types.NewQueryServiceClient(clientCtx)
			res, err := queryClient.Key(cmd.Context(),
				&types.KeyRequest{
					KeyID: keyID,
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

// GetCmdKeygenCandidates returns the candidates chosen for a keygen corresponding to the given key ID
func GetCmdKeygenCandidates(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keygen-candidates [key-id]",
		Short: "Returns the candidates chosen for a keygen corresponding to the given key ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			keyID := multisig.KeyID(utils.NormalizeString(args[0]))
			queryClient := types.NewQueryServiceClient(clientCtx)
			res, err := queryClient.KeygenCandidates(cmd.Context(),
				&types.KeygenCandidatesRequest{
					KeyID: keyID,
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
