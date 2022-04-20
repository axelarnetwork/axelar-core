package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/x/tss/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/keeper"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/axelar-core/x/tss/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	tssQueryCmd := &cobra.Command{
		Use:                        "tss",
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	tssQueryCmd.AddCommand(
		GetCmdAssignableKey(queryRoute),
		GetCmdGetSig(queryRoute),
		GetCmdGetKey(queryRoute),
		GetCmdRecovery(queryRoute),
		GetCmdGetKeyID(queryRoute),
		GetCmdNextKeyID(queryRoute),
		GetCmdGetKeySharesByKeyID(queryRoute),
		GetCmdGetKeySharesByValidator(queryRoute),
		GetCmdGetActiveOldKeys(queryRoute),
		GetCmdGetActiveOldKeysByValidator(queryRoute),
		GetCmdGetDeactivatedOperators(queryRoute),
		GetCmdExternalKeyID(queryRoute),
	)

	return tssQueryCmd
}

// GetCmdGetSig returns the query for a signature by its sigID
func GetCmdGetSig(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "signature [sig ID]",
		Short: "Query a signature by sig ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			sigID := args[0]
			bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QuerySignature, sigID))
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get signature")
			}

			var res types.QuerySignatureResponse
			if err := res.Unmarshal(bz); err != nil {
				return sdkerrors.Wrapf(err, "failed to get signature")
			}

			return cliCtx.PrintProto(&res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetKey returns the query for a key by its keyID
func GetCmdGetKey(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key [key ID]",
		Short: "Query a key by key ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			keyID := args[0]
			bz, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryKey, keyID))
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get key")
			}

			var res types.QueryKeyResponse
			if err := res.Unmarshal(bz); err != nil {
				return sdkerrors.Wrapf(err, "failed to get key")
			}

			// force the rotatedAt field to be nil, if the timestamp is for Jan 1, 1970
			// necessary because of the marshaling/unmarshaling of the Amino codec
			if res.RotatedAt != nil && res.RotatedAt.Unix() == 0 {
				res.RotatedAt = nil
			}

			return cliCtx.PrintProto(&res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdRecovery returns the command for share recovery
func GetCmdRecovery(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recover [validator address] [key ID #1] ... [key ID #N]",
		Short: "Attempt to recover the shares for the specified key ID",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			address, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to parse validator address")
			}

			keyIDs := args[1:]
			requests := make([]tofnd.RecoverRequest, len(keyIDs))
			for i, keyID := range keyIDs {
				res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QueryRecovery, keyID, address.String()))
				if err != nil {
					return sdkerrors.Wrapf(err, "failed to get recovery data")
				}

				var recResponse types.QueryRecoveryResponse
				err = recResponse.Unmarshal(res)
				if err != nil {
					return sdkerrors.Wrapf(err, "failed to get recovery data")
				}

				index := utils.IndexOf(recResponse.PartyUids, address.String())
				if index == -1 {
					// not participating
					return sdkerrors.Wrapf(err, "recovery data does not contain address %s", address.String())
				}

				requests[i] = tofnd.RecoverRequest{
					KeygenInit: &tofnd.KeygenInit{
						NewKeyUid:        keyID,
						Threshold:        recResponse.Threshold,
						PartyUids:        recResponse.PartyUids,
						PartyShareCounts: recResponse.PartyShareCounts,
						MyPartyIndex:     uint32(index),
					},
					KeygenOutput: recResponse.KeygenOutput,
				}
			}
			return cliCtx.PrintObjectLegacy(requests)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetKeyID returns the command for the keyID of the most recent key given the keyChain and keyRole
func GetCmdGetKeyID(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key-id [chain] [role]",
		Short: "Query the keyID using keyChain and keyRole",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			keyChain := args[0]
			keyRole := args[1]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QueryKeyID, keyChain, keyRole), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get keyID")
			}

			return cliCtx.PrintObjectLegacy(string(res))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetKeySharesByKeyID returns the query for a list of key shares for a given keyID
func GetCmdGetKeySharesByKeyID(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key-shares-by-key-id [key ID]",
		Short: "Query key shares information by key ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			keyID := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryKeySharesByKeyID, keyID), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get key share information")
			}

			var keyShareResponse types.QueryKeyShareResponse
			err = keyShareResponse.Unmarshal(res)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get key share information")
			}

			return cliCtx.PrintObjectLegacy(keyShareResponse.ShareInfos)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetKeySharesByValidator returns the query for a list of key shares held by a validator address
func GetCmdGetKeySharesByValidator(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key-shares-by-validator [validator address]",
		Short: "Query key shares information by validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			validatorAddress := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryKeySharesByValidator, validatorAddress), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get key share information")
			}

			var keyShareResponse types.QueryKeyShareResponse
			err = keyShareResponse.Unmarshal(res)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get key share information")
			}

			return cliCtx.PrintObjectLegacy(keyShareResponse.ShareInfos)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetActiveOldKeys returns the query for a list of active old key IDs held by a validator address
func GetCmdGetActiveOldKeys(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active-old-keys [chain] [role]",
		Short: "Query active old key IDs by validator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			role := args[1]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%s", queryRoute, keeper.QueryActiveOldKeys, chain, role), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get key share information")
			}

			var keyIDsResponse types.QueryActiveOldKeysResponse
			err = keyIDsResponse.Unmarshal(res)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get key share information")
			}

			return cliCtx.PrintObjectLegacy(keyIDsResponse.KeyIDs)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetActiveOldKeysByValidator returns the query for a list of active old key IDs held by a validator address
func GetCmdGetActiveOldKeysByValidator(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active-old-keys-by-validator [validator address]",
		Short: "Query active old key IDs by validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			validatorAddress := args[0]
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryActiveOldKeysByValidator, validatorAddress), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get key share information")
			}

			var keyIDsResponse types.QueryActiveOldKeysValidatorResponse
			err = keyIDsResponse.Unmarshal(res)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get key share information")
			}

			return cliCtx.PrintObjectLegacy(keyIDsResponse.KeysInfo)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetDeactivatedOperators returns the list of deactivated operator addresses
func GetCmdGetDeactivatedOperators(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivated-operators",
		Short: "Fetch the list of deactivated operator addresses",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			bz, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, keeper.QueryDeactivated), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "could not get deactivated operator addresses")
			}
			var res types.QueryDeactivatedOperatorsResponse
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

			return cliCtx.PrintProto(&res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdExternalKeyID returns the keyIDs of the current set of external keys for the given chain
func GetCmdExternalKeyID(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "external-key-id [chain]",
		Short: "Returns the key IDs of the current external keys for the given chain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := args[0]
			path := fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QExternalKeyID, chain)

			bz, _, err := clientCtx.Query(path)
			if err != nil {
				return sdkerrors.Wrap(err, "could not resolve the external key IDs")
			}

			var res types.QueryExternalKeyIDResponse
			types.ModuleCdc.MustUnmarshalLengthPrefixed(bz, &res)

			return clientCtx.PrintProto(&res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdNextKeyID returns the key ID assigned for the next rotation on a given chain and for the given key role
func GetCmdNextKeyID(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next-key-id [chain] [role]",
		Short: "Returns the key ID assigned for the next rotation on a given chain and for the given key role",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := utils.NormalizeString(args[0])
			keyRole, err := exported.KeyRoleFromSimpleStr(args[1])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)
			res, err := queryClient.NextKeyID(cmd.Context(),
				&types.NextKeyIDRequest{
					Chain:   chain,
					KeyRole: keyRole,
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

// GetCmdAssignableKey returns true if a key can be assigned for the next rotation on a given chain and for the given key role
func GetCmdAssignableKey(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assignable-key [chain] [role]",
		Short: "Returns the true if a key can be assigned for the next rotation on a given chain and for the given key role",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			chain := utils.NormalizeString(args[0])
			keyRole, err := exported.KeyRoleFromSimpleStr(args[1])
			if err != nil {
				return err
			}

			queryClient := types.NewQueryServiceClient(clientCtx)
			res, err := queryClient.AssignableKey(cmd.Context(),
				&types.AssignableKeyRequest{
					Chain:   chain,
					KeyRole: keyRole,
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
