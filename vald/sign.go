package vald

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/utils"
	"github.com/axelarnetwork/axelar-core/vald/config"
	"github.com/axelarnetwork/axelar-core/vald/tss"
	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	multisigTypes "github.com/axelarnetwork/axelar-core/x/multisig/types"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/utils/funcs"
)

// GetSignCommand returns the command to execute a manual sign request from vald
func GetSignCommand() *cobra.Command {
	flagPubKey := "pubkey"

	cmd := &cobra.Command{
		Use:   "vald-sign [key-id] [validator-addr] [hash to sign]",
		Short: "Sign hash with the key corresponding to the key id for the given validator. If unspecified, the public key will be retrieved from the node.",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			keyID := multisig.KeyID(args[0])
			if err := keyID.ValidateBasic(); err != nil {
				return err
			}

			valAddr := strings.ToLower(args[1])
			if _, err := sdk.ValAddressFromBech32(valAddr); err != nil {
				return err
			}

			pubKeyHex, err := cmd.Flags().GetString(flagPubKey)
			if err != nil {
				return err
			}

			if pubKeyHex == "" {
				pubKeyHex, err = getPubKeyByValidator(cmd.Context(), clientCtx, valAddr, keyID)
				if err != nil {
					return err
				}
			}

			pubKeyRaw, err := utils.HexDecode(pubKeyHex)
			if err != nil {
				return err
			}

			pubKey := multisig.PublicKey(pubKeyRaw)
			if err := pubKey.ValidateBasic(); err != nil {
				return err
			}

			hashRaw, err := utils.HexDecode(args[2])
			if err != nil {
				return err
			}

			if len(hashRaw) != common.HashLength {
				return fmt.Errorf("hash to sign must be 32 bytes")
			}

			hash := common.BytesToHash(hashRaw)

			serverCtx := server.GetServerContextFromCmd(cmd)
			valdCfg := config.DefaultValdConfig()
			if err := serverCtx.Viper.Unmarshal(&valdCfg); err != nil {
				panic(err)
			}

			conn, err := tss.Connect(valdCfg.TssConfig.Host, valdCfg.TssConfig.Port, valdCfg.TssConfig.DialTimeout)
			if err != nil {
				return fmt.Errorf("failed to reach tofnd: %s", err.Error())
			}

			// creates client to communicate with the external tofnd process multisig service
			client := tofnd.NewMultisigClient(conn)

			grpcCtx, cancel := context.WithTimeout(cmd.Context(), timeout)
			defer cancel()

			res, err := client.Sign(grpcCtx, &tofnd.SignRequest{
				KeyUid:    fmt.Sprintf("%s_%d", keyID, 0),
				MsgToSign: hash.Bytes(),
				PartyUid:  valAddr,
				PubKey:    pubKey,
			})

			if err != nil {
				return sdkerrors.Wrapf(err, "failed signing")
			}

			switch res.GetSignResponse().(type) {
			case *tofnd.SignResponse_Signature:
				ecdsaSig := *funcs.Must(ec.ParseDERSignature(res.GetSignature()))
				evmSignature := funcs.Must(evm.ToSignature(ecdsaSig, hash, pubKey.ToECDSAPubKey())).ToHomesteadSig()

				signDetails := map[string]string{
					"key_id":    keyID.String(),
					"validator": valAddr,
					"msg_hash":  utils.HexEncode(hash.Bytes()),
					"pub_key":   utils.HexEncode(pubKey),
					"signature": utils.HexEncode(evmSignature),
				}
				fmt.Printf("%s", funcs.Must(json.MarshalIndent(signDetails, "", "  ")))

				return nil
			case *tofnd.SignResponse_Error:
				return fmt.Errorf(res.GetError())
			default:
				panic(fmt.Errorf("unknown multisig sign response %T", res.GetSignResponse()))
			}
		},
	}

	cmd.Flags().String(flagPubKey, "", "the public key of the validator for the key id in hex format")

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func getPubKeyByValidator(ctx context.Context, clientCtx client.Context, valAddr string, keyID multisig.KeyID) (string, error) {
	queryClient := multisigTypes.NewQueryServiceClient(clientCtx)
	res, err := queryClient.Key(ctx, &multisigTypes.KeyRequest{KeyID: keyID})
	if err != nil {
		return "", err
	}

	for _, participant := range res.Participants {
		if participant.Address == valAddr {
			return participant.PubKey, nil
		}
	}

	return "", fmt.Errorf("validator %s is not a participant for key %s", valAddr, keyID)
}
