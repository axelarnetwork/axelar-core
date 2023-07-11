package vald

import (
	"context"
	"encoding/hex"
	"fmt"

	ec "github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/cosmos/cosmos-sdk/server"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	"github.com/axelarnetwork/axelar-core/vald/config"
	"github.com/axelarnetwork/axelar-core/vald/tss"
	evm "github.com/axelarnetwork/axelar-core/x/evm/types"
	multisig "github.com/axelarnetwork/axelar-core/x/multisig/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
	"github.com/axelarnetwork/utils/funcs"
)

// GetHealthCheckCommand returns the command to execute a node health check
func GetSignCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "vald-sign [key-id] [public-key] [hash to sign]",
		Short: "Sign hash with specified key",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			keyID := multisig.KeyID(args[0])
			if err := keyID.ValidateBasic(); err != nil {
				return err
			}

			pubKeyRaw, err := hex.DecodeString(args[1])
			if err != nil {
				return err
			}

			pubKey := multisig.PublicKey(pubKeyRaw)
			if err := pubKey.ValidateBasic(); err != nil {
				return err
			}

			hashRaw, err := hex.DecodeString(args[2])
			if err != nil {
				return err
			}

			if len(hashRaw) != common.HashLength {
				return fmt.Errorf("hash to sign must be 32 bytes")
			}

			hash := common.BytesToHash(hashRaw)

			valAddr, err := cmd.Flags().GetString("validator-addr")
			if err != nil {
				return err
			}

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
				fmt.Printf("signature: %s\n", hex.EncodeToString(evmSignature))
				return nil
			case *tofnd.SignResponse_Error:
				return fmt.Errorf(res.GetError())
			default:
				panic(fmt.Errorf("unknown multisig sign response %T", res.GetSignResponse()))
			}

		},
	}
	cmd.Flags().String("validator-addr", "", "the address of the validator operator, i.e axelarvaloper1..")

	return cmd
}
