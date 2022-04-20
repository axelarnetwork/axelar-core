package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	reward "github.com/axelarnetwork/axelar-core/x/reward/exported"
	"github.com/axelarnetwork/axelar-core/x/tss/tofnd"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&HeartBeatRequest{}, "tss/HeartBeatRequest", nil)
	cdc.RegisterConcrete(&StartKeygenRequest{}, "tss/StartKeygen", nil)
	cdc.RegisterConcrete(&ProcessKeygenTrafficResponse{}, "tss/KeygenTraffic", nil)
	cdc.RegisterConcrete(&ProcessSignTrafficRequest{}, "tss/SignTraffic", nil)
	cdc.RegisterConcrete(&RotateKeyRequest{}, "tss/RotateKey", nil)
	cdc.RegisterConcrete(&VoteSigRequest{}, "tss/VoteSig", nil)
	cdc.RegisterConcrete(&VotePubKeyRequest{}, "tss/VotePubKey", nil)
	cdc.RegisterConcrete(&RegisterExternalKeysRequest{}, "tss/RegisterExternalKey", nil)
	cdc.RegisterConcrete(&SubmitMultisigPubKeysRequest{}, "tss/SubmitMultisigPubKeys", nil)
	cdc.RegisterConcrete(&SubmitMultisigSignaturesRequest{}, "tss/SubmitMultisigSignatures", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&HeartBeatRequest{},
		&StartKeygenRequest{},
		&ProcessKeygenTrafficRequest{},
		&ProcessSignTrafficRequest{},
		&RotateKeyRequest{},
		&VoteSigRequest{},
		&VotePubKeyRequest{},
		&RegisterExternalKeysRequest{},
		&SubmitMultisigPubKeysRequest{},
		&SubmitMultisigSignaturesRequest{},
	)
	registry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
		&tofnd.MessageOut_SignResult{},
		&tofnd.MessageOut_KeygenResult{},
		&tofnd.MessageOut_CriminalList{},
		&QueryRecoveryResponse{},
		&KeygenVoteData{},
	)

	registry.RegisterImplementations((*reward.Refundable)(nil),
		&HeartBeatRequest{},
		&ProcessKeygenTrafficRequest{},
		&VotePubKeyRequest{},
		&ProcessSignTrafficRequest{},
		&VoteSigRequest{},
		&SubmitMultisigPubKeysRequest{},
		&SubmitMultisigSignaturesRequest{},
	)
}

var amino = codec.NewLegacyAmino()

// ModuleCdc defines the module codec
var ModuleCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
