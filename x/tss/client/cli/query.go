package cli

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/spf13/cobra"

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
		GetCmdGetSig(queryRoute),
		GetCmdGetKey(queryRoute),
		GetCmdRecovery(queryRoute),
		GetCmdGetKeyID(queryRoute),
		GetCmdGetKeySharesByKeyID(queryRoute),
		GetCmdGetKeySharesByValidator(queryRoute),
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
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QuerySigStatus, sigID), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get signature")
			}

			var sigResponse types.QuerySigResponse
			err = sigResponse.Unmarshal(res)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get signature")
			}

			hexSig := types.NewHexSignatureFromQuerySigResponse(&sigResponse)
			return cliCtx.PrintObjectLegacy(hexSig)
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
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryKeyStatus, keyID), nil)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get key")
			}

			var keyResponse types.QueryKeyResponse
			err = keyResponse.Unmarshal(res)
			if err != nil {
				return sdkerrors.Wrapf(err, "failed to get key")
			}

			return cliCtx.PrintObjectLegacy(keyResponse)
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
				res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/%s/%s", queryRoute, keeper.QueryRecovery, keyID))
				if err != nil {
					return sdkerrors.Wrapf(err, "failed to get recovery data")
				}

				var recResponse types.QueryRecoveryResponse
				err = recResponse.Unmarshal(res)
				if err != nil {
					return sdkerrors.Wrapf(err, "failed to get recovery data")
				}

				var index int32 = -1
				for i, participant := range recResponse.PartyUids {
					if address.String() == participant {
						index = int32(i)
						break
					}
				}
				// not participating
				if index == -1 {
					return sdkerrors.Wrapf(err, "recovery data does not contain address %s", address.String())
				}

				requests[i] = tofnd.RecoverRequest{
					KeygenInit: &tofnd.KeygenInit{
						NewKeyUid:        keyID,
						Threshold:        recResponse.Threshold,
						PartyUids:        recResponse.PartyUids,
						PartyShareCounts: recResponse.PartyShareCounts,
						MyPartyIndex:     index,
					},
					ShareRecoveryInfos: recResponse.ShareRecoveryInfos,
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
		Use:   "keyID [key chain] [key role]",
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
		Use:   "keySharesKeyID [key ID]",
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

			return cliCtx.PrintProto(&keyShareResponse)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// GetCmdGetKeySharesByValidator returns the query for a list of key shares held by a validator address
func GetCmdGetKeySharesByValidator(queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keySharesValidator [validator address]",
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

			return cliCtx.PrintProto(&keyShareResponse)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
