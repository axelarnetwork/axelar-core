package params

import (
	"fmt"

	coreaddress "cosmossdk.io/core/address"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/gogoproto/proto"
	googleproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	nexusExported "github.com/axelarnetwork/axelar-core/x/nexus/exported"
	"github.com/axelarnetwork/utils/funcs"
)

// MakeEncodingConfig creates an EncodingConfig for the application.
func MakeEncodingConfig() EncodingConfig {
	amino := codec.NewLegacyAmino()

	addrCodec := address.Bech32Codec{
		Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
	}

	interfaceRegistry := funcs.Must(types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          addrCodec,
			ValidatorAddressCodec: address.Bech32Codec{Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix()},
			CustomGetSigners:      customGetSigners(addrCodec),
		},
	}))

	marshaler := codec.NewProtoCodec(interfaceRegistry)
	txCfg := tx.NewTxConfig(marshaler, tx.DefaultSignModes)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             marshaler,
		TxConfig:          txCfg,
		Amino:             amino,
	}
}

// makeSenderDeprecatedGetSigners creates a GetSignersFunc that handles messages with both
// "sender" (string) and "sender_deprecated" (bytes) fields. This is needed for backward
// compatibility with historical transactions that used the deprecated sender field.
func makeSenderDeprecatedGetSigners(addrCodec coreaddress.Codec) signing.GetSignersFunc {
	return func(msg googleproto.Message) ([][]byte, error) {
		m := msg.ProtoReflect()
		desc := m.Descriptor()

		// Try the new sender field first (string type)
		if senderField := desc.Fields().ByName("sender"); senderField != nil {
			if senderStr := m.Get(senderField).String(); senderStr != "" {
				addrBz, err := addrCodec.StringToBytes(senderStr)
				if err != nil {
					return nil, fmt.Errorf("invalid sender address: %w", err)
				}
				return [][]byte{addrBz}, nil
			}
		}

		// Fall back to sender_deprecated field (bytes type)
		if senderDepField := desc.Fields().ByName("sender_deprecated"); senderDepField != nil {
			if senderBytes := m.Get(senderDepField).Bytes(); len(senderBytes) > 0 {
				return [][]byte{senderBytes}, nil
			}
		}

		return nil, fmt.Errorf("no sender found in message %s", desc.FullName())
	}
}

// customGetSigners returns custom signer functions for messages that need special handling.
// This includes:
// - WasmMessage: no signers (wasm-generated messages)
// - Messages with sender_deprecated: backward compatibility with historical transactions
func customGetSigners(addrCodec coreaddress.Codec) map[protoreflect.FullName]signing.GetSignersFunc {
	senderDeprecatedGetSigners := makeSenderDeprecatedGetSigners(addrCodec)

	signers := map[protoreflect.FullName]signing.GetSignersFunc{
		// WasmMessage has no signer
		protoreflect.FullName(proto.MessageName(&nexusExported.WasmMessage{})): func(msg googleproto.Message) ([][]byte, error) {
			return [][]byte{}, nil
		},
	}

	// Messages with sender_deprecated field for backward compatibility
	// These messages have both "sender" (new, string) and "sender_deprecated" (old, bytes) fields
	messagesWithSenderDeprecated := []string{
		// auxiliary
		"axelar.auxiliary.v1beta1.BatchRequest",
		// axelarnet
		"axelar.axelarnet.v1beta1.LinkRequest",
		"axelar.axelarnet.v1beta1.ConfirmDepositRequest",
		"axelar.axelarnet.v1beta1.ExecutePendingTransfersRequest",
		"axelar.axelarnet.v1beta1.RegisterIBCPathRequest",
		"axelar.axelarnet.v1beta1.AddCosmosBasedChainRequest",
		"axelar.axelarnet.v1beta1.RegisterAssetRequest",
		"axelar.axelarnet.v1beta1.RouteIBCTransfersRequest",
		"axelar.axelarnet.v1beta1.RegisterFeeCollectorRequest",
		"axelar.axelarnet.v1beta1.RetryIBCTransferRequest",
		"axelar.axelarnet.v1beta1.RouteMessageRequest",
		"axelar.axelarnet.v1beta1.CallContractRequest",
		// evm
		"axelar.evm.v1beta1.SetGatewayRequest",
		"axelar.evm.v1beta1.ConfirmGatewayTxRequest",
		"axelar.evm.v1beta1.ConfirmGatewayTxsRequest",
		"axelar.evm.v1beta1.ConfirmDepositRequest",
		"axelar.evm.v1beta1.ConfirmTokenRequest",
		"axelar.evm.v1beta1.ConfirmTransferKeyRequest",
		"axelar.evm.v1beta1.LinkRequest",
		"axelar.evm.v1beta1.CreateBurnTokensRequest",
		"axelar.evm.v1beta1.CreateDeployTokenRequest",
		"axelar.evm.v1beta1.CreatePendingTransfersRequest",
		"axelar.evm.v1beta1.CreateTransferOwnershipRequest",
		"axelar.evm.v1beta1.CreateTransferOperatorshipRequest",
		"axelar.evm.v1beta1.SignCommandsRequest",
		"axelar.evm.v1beta1.AddChainRequest",
		"axelar.evm.v1beta1.RetryFailedEventRequest",
		// multisig
		"axelar.multisig.v1beta1.StartKeygenRequest",
		"axelar.multisig.v1beta1.SubmitPubKeyRequest",
		"axelar.multisig.v1beta1.SubmitSignatureRequest",
		"axelar.multisig.v1beta1.RotateKeyRequest",
		"axelar.multisig.v1beta1.KeygenOptOutRequest",
		"axelar.multisig.v1beta1.KeygenOptInRequest",
		// nexus
		"axelar.nexus.v1beta1.RegisterChainMaintainerRequest",
		"axelar.nexus.v1beta1.DeregisterChainMaintainerRequest",
		"axelar.nexus.v1beta1.ActivateChainRequest",
		"axelar.nexus.v1beta1.DeactivateChainRequest",
		"axelar.nexus.v1beta1.RegisterAssetFeeRequest",
		"axelar.nexus.v1beta1.SetTransferRateLimitRequest",
		// permission
		"axelar.permission.v1beta1.UpdateGovernanceKeyRequest",
		"axelar.permission.v1beta1.RegisterControllerRequest",
		"axelar.permission.v1beta1.DeregisterControllerRequest",
		// reward
		"axelar.reward.v1beta1.RefundMsgRequest",
		// snapshot
		"axelar.snapshot.v1beta1.RegisterProxyRequest",
		"axelar.snapshot.v1beta1.DeactivateProxyRequest",
		// tss
		"axelar.tss.v1beta1.StartKeygenRequest",
		"axelar.tss.v1beta1.RotateKeyRequest",
		"axelar.tss.v1beta1.ProcessKeygenTrafficRequest",
		"axelar.tss.v1beta1.ProcessSignTrafficRequest",
		"axelar.tss.v1beta1.VotePubKeyRequest",
		"axelar.tss.v1beta1.VoteSigRequest",
		"axelar.tss.v1beta1.HeartBeatRequest",
		"axelar.tss.v1beta1.RegisterExternalKeysRequest",
		"axelar.tss.v1beta1.SubmitMultisigPubKeysRequest",
		"axelar.tss.v1beta1.SubmitMultisigSignaturesRequest",
		// vote
		"axelar.vote.v1beta1.VoteRequest",
	}

	for _, msgName := range messagesWithSenderDeprecated {
		signers[protoreflect.FullName(msgName)] = senderDeprecatedGetSigners
	}

	return signers
}
