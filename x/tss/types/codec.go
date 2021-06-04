package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers concrete types on codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&StartKeygenRequest{}, "tss/StartKeygen", nil)
	cdc.RegisterConcrete(&ProcessKeygenTrafficResponse{}, "tss/KeygenTraffic", nil)
	cdc.RegisterConcrete(&ProcessSignTrafficRequest{}, "tss/SignTraffic", nil)
	cdc.RegisterConcrete(&AssignKeyRequest{}, "tss/AssignKey", nil)
	cdc.RegisterConcrete(&RotateKeyRequest{}, "tss/RotateKey", nil)
	cdc.RegisterConcrete(&VoteSigRequest{}, "tss/VoteSig", nil)
	cdc.RegisterConcrete(&VotePubKeyRequest{}, "tss/VotePubKey", nil)
	cdc.RegisterConcrete(&DeregisterRequest{}, "tss/Deregister", nil)
}

// RegisterInterfaces registers types and interfaces with the given registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&StartKeygenRequest{},
		&ProcessKeygenTrafficRequest{},
		&ProcessSignTrafficRequest{},
		&AssignKeyRequest{},
		&RotateKeyRequest{},
		&VoteSigRequest{},
		&VotePubKeyRequest{},
		&DeregisterRequest{},
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
